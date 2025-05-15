package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const engineURL = "http://engine:8000/api/events"

type HTTPEvent struct {
	Timestamp   time.Time         `json:"timestamp"`
	SrcIP       string            `json:"src_ip,omitempty"`
	Method      string            `json:"method"`
	URI         string            `json:"uri"`
	Headers     map[string]string `json:"headers"`
	PayloadSize int               `json:"payload_size"`
	Description string            `json:"description,omitempty"`
}

func sendHTTPEvent(event HTTPEvent) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling HTTP event: %v", err)
		return
	}

	resp, err := http.Post(engineURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending HTTP event to engine: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Engine responded with status: %v", resp.Status)
	} else {
		log.Printf("Successfully sent event to engine: %v %v", event.Method, event.URI)
	}
}
