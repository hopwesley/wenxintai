package stream

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hopwesley/wenxintai/server/internal/store"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func NewWSHandler(repo store.Repo, broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		assessmentID := parseWSPath(r.URL.Path)
		if assessmentID == "" {
			http.NotFound(w, r)
			return
		}
		if _, err := repo.GetAssessmentByID(r.Context(), assessmentID); err != nil {
			if err == store.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "failed to load assessment", http.StatusInternalServerError)
			return
		}

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "websocket not supported", http.StatusInternalServerError)
			return
		}
		conn, rw, err := hj.Hijack()
		if err != nil {
			return
		}

		if err := performHandshake(r, rw); err != nil {
			conn.Close()
			return
		}

		lastID := r.URL.Query().Get("last_event_id")
		events, cancel := broker.Subscribe(assessmentID, lastID)
		defer cancel()

		ctx, cancelCtx := context.WithCancel(r.Context())
		defer cancelCtx()

		go readLoop(ctx, conn, rw, cancelCtx)

		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				conn.Close()
				return
			case <-ticker.C:
				if err := writePing(conn); err != nil {
					conn.Close()
					return
				}
			case evt, ok := <-events:
				if !ok {
					conn.Close()
					return
				}
				payload := map[string]interface{}{
					"id":   evt.ID,
					"type": evt.Type,
				}
				if len(evt.Data) > 0 {
					payload["data"] = json.RawMessage(evt.Data)
				}
				message, err := json.Marshal(payload)
				if err != nil {
					continue
				}
				if err := writeTextFrame(conn, message); err != nil {
					conn.Close()
					return
				}
			}
		}
	}
}

func parseWSPath(path string) string {
	trimmed := strings.TrimPrefix(path, "/ws/assessments/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return ""
	}
	return trimmed
}

func performHandshake(r *http.Request, rw *bufio.ReadWriter) error {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return fmt.Errorf("not a websocket upgrade")
	}
	if !strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		return fmt.Errorf("connection header missing upgrade")
	}
	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		return fmt.Errorf("unsupported websocket version")
	}
	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		return fmt.Errorf("missing websocket key")
	}
	acceptKey := computeAcceptKey(key)
	response := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", acceptKey)
	if _, err := rw.WriteString(response); err != nil {
		return err
	}
	return rw.Flush()
}

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func readLoop(ctx context.Context, conn net.Conn, rw *bufio.ReadWriter, cancel context.CancelFunc) {
	defer cancel()
	for {
		opcode, payload, err := readFrame(rw)
		if err != nil {
			return
		}
		switch opcode {
		case 0x8: // close
			return
		case 0x9: // ping
			_ = writeControlFrame(conn, 0xA, payload)
		default:
			// ignore other frames
		}
	}
}

func readFrame(rw *bufio.ReadWriter) (byte, []byte, error) {
	header := make([]byte, 2)
	if err := ioReadFull(rw, header); err != nil {
		return 0, nil, err
	}
	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	length := int64(header[1] & 0x7F)

	switch length {
	case 126:
		extended := make([]byte, 2)
		if err := ioReadFull(rw, extended); err != nil {
			return 0, nil, err
		}
		length = int64(binary.BigEndian.Uint16(extended))
	case 127:
		extended := make([]byte, 8)
		if err := ioReadFull(rw, extended); err != nil {
			return 0, nil, err
		}
		length = int64(binary.BigEndian.Uint64(extended))
	}

	var maskKey [4]byte
	if masked {
		if err := ioReadFull(rw, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, length)
	if err := ioReadFull(rw, payload); err != nil {
		return 0, nil, err
	}

	if masked {
		for i := int64(0); i < length; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, payload, nil
}

func writeTextFrame(conn net.Conn, payload []byte) error {
	return writeFrame(conn, 0x1, payload)
}

func writePing(conn net.Conn) error {
	return writeFrame(conn, 0x9, nil)
}

func writeControlFrame(conn net.Conn, opcode byte, payload []byte) error {
	return writeFrame(conn, opcode, payload)
}

func writeFrame(conn net.Conn, opcode byte, payload []byte) error {
	var header []byte
	payloadLen := len(payload)
	b0 := byte(0x80 | (opcode & 0x0F))
	header = append(header, b0)

	switch {
	case payloadLen <= 125:
		header = append(header, byte(payloadLen))
	case payloadLen <= 65535:
		header = append(header, 126)
		extended := make([]byte, 2)
		binary.BigEndian.PutUint16(extended, uint16(payloadLen))
		header = append(header, extended...)
	default:
		header = append(header, 127)
		extended := make([]byte, 8)
		binary.BigEndian.PutUint64(extended, uint64(payloadLen))
		header = append(header, extended...)
	}

	if _, err := conn.Write(header); err != nil {
		return err
	}
	if payloadLen > 0 {
		if _, err := conn.Write(payload); err != nil {
			return err
		}
	}
	return nil
}

func ioReadFull(r *bufio.ReadWriter, buf []byte) error {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}
