package main

import (
    "sync"
    "time"
)

var (
    requestCounts   = make(map[string]int)
    requestCountsMu sync.Mutex
)

func monitorHTTPRequest(method, uri string, headers map[string]string, payloadSize int) {
    now := time.Now()

    // Track request counts per URI for rate limiting detection
    requestCountsMu.Lock()
    requestCounts[uri]++
    count := requestCounts[uri]
    requestCountsMu.Unlock()

    event := HTTPEvent{
        Timestamp:   now,
        Method:      method,
        URI:         uri,
        Headers:     headers,
        PayloadSize: payloadSize,
    }

    // Detect high frequency requests (simple threshold example)
    if count > 100 {
        event.Description = "High frequency requests detected - possible DDoS"
    }

    sendHTTPEvent(event)
}

// Reset request counts every minute
func resetRequestCounts() {
    for {
        time.Sleep(1 * time.Minute)
        requestCountsMu.Lock()
        requestCounts = make(map[string]int)
        requestCountsMu.Unlock()
    }
}
