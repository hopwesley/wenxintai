package srv

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	opText  byte = 0x1
	opClose byte = 0x8
	opPing  byte = 0x9
)

type wsConn struct {
	conn net.Conn
	mu   sync.Mutex
}

func (c *wsConn) WriteJSON(v interface{}) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.writeFrame(opText, payload)
}

func (c *wsConn) WriteClose() error {
	return c.writeFrame(opClose, nil)
}

func (c *wsConn) WritePing() error {
	return c.writeFrame(opPing, []byte("ping"))
}

func (c *wsConn) Close() error {
	return c.conn.Close()
}

func (c *wsConn) writeFrame(opcode byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	head := []byte{0x80 | opcode}
	pLen := len(payload)
	switch {
	case pLen < 126:
		head = append(head, byte(pLen))
	case pLen <= 0xFFFF:
		head = append(head, 126)
		ext := make([]byte, 2)
		binary.BigEndian.PutUint16(ext, uint16(pLen))
		head = append(head, ext...)
	default:
		head = append(head, 127)
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(pLen))
		head = append(head, ext...)
	}

	if _, err := c.conn.Write(head); err != nil {
		return err
	}

	if pLen > 0 {
		if _, err := c.conn.Write(payload); err != nil {
			return err
		}
	}

	return nil
}

type wsUpgrader struct {
	allowedOrigins   map[string]struct{}
	handshakeTimeout time.Duration
}

func newWSUpgrader(origins []string, timeout time.Duration) *wsUpgrader {
	allowed := map[string]struct{}{}
	for _, o := range origins {
		allowed[o] = struct{}{}
	}
	return &wsUpgrader{allowedOrigins: allowed, handshakeTimeout: timeout}
}

func (u *wsUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (*wsConn, error) {
	if len(u.allowedOrigins) > 0 {
		origin := r.Header.Get("Origin")
		if _, ok := u.allowedOrigins[origin]; !ok {
			return nil, fmt.Errorf("origin not allowed: %s", origin)
		}
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("websocket upgrade not supported")
	}

	conn, buf, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	if u.handshakeTimeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(u.handshakeTimeout))
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		_ = conn.Close()
		return nil, fmt.Errorf("missing Sec-WebSocket-Key")
	}

	accept := computeAcceptKey(key)
	response := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", accept)

	if _, err := buf.WriteString(response); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := buf.Flush(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	_ = conn.SetDeadline(time.Time{})
	return &wsConn{conn: conn}, nil
}

func computeAcceptKey(key string) string {
	h := sha1.New()
	_, _ = io.WriteString(h, key+"258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
