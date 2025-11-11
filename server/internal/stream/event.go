package stream

import "encoding/json"

type Payload struct {
	Type string
	Data interface{}
}

type Event struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func NewStatusEvent(phase string) Payload {
	return Payload{Type: "status", Data: map[string]string{"phase": phase}}
}

func NewTokenEvent(text string) Payload {
	return Payload{Type: "token", Data: map[string]string{"text": text}}
}

func NewProgressEvent(tokens int) Payload {
	return Payload{Type: "progress", Data: map[string]int{"tokens": tokens}}
}

func NewFinalEvent(reportID string) Payload {
	return Payload{Type: "final", Data: map[string]string{"report_id": reportID}}
}

func NewErrorEvent(code, message string) Payload {
	return Payload{Type: "error", Data: map[string]string{"code": code, "message": message}}
}
