package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	monitoringActive  = false
	mu                sync.Mutex
	captureCancelFunc context.CancelFunc
	watchCtx          context.Context
	resetCtx          context.Context
)

func isMonitoringActive() bool {
	mu.Lock()
	defer mu.Unlock()
	return monitoringActive
}

func setMonitoringActive(active bool) {
	mu.Lock()
	monitoringActive = active
	mu.Unlock()
}

func watchFilesControlled(ctx context.Context, watchDir string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopped file watching")
			return
		default:
			watchFiles(watchDir)
			time.Sleep(1 * time.Second)
		}
	}
}

func resetRequestCountsControlled(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopped resetting request counts")
			return
		default:
			resetRequestCounts()
			time.Sleep(1 * time.Second)
		}
	}
}

func capturePacketsControlled(ctx context.Context) {
	device := getDevice()
	handle, err := pcap.OpenLive(device, 1024, false, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	var filter = "tcp port 80 or tcp port 8080 or tcp port 3000"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Only capturing TCP port 80, 8080, or 3000 packets.")

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopped packet capture")
			return
		case packet := <-packetSource.Packets():
			appLayer := packet.ApplicationLayer()
			if appLayer != nil {
				payload := appLayer.Payload()
				if len(payload) > 0 {
					payloadStr := string(payload)
					if strings.HasPrefix(payloadStr, "GET") || strings.HasPrefix(payloadStr, "POST") {
						lines := strings.Split(payloadStr, "\r\n")
						if len(lines) > 0 {
							requestLine := lines[0]
							parts := strings.Split(requestLine, " ")
							if len(parts) >= 2 {
								method := parts[0]
								uri := parts[1]
								headers := make(map[string]string)
								for _, line := range lines[1:] {
									if line == "" {
										break
									}
									headerParts := strings.SplitN(line, ":", 2)
									if len(headerParts) == 2 {
										headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
									}
								}
								srcIP := ""
								if netLayer := packet.NetworkLayer(); netLayer != nil {
									srcIP = netLayer.NetworkFlow().Src().String()
								}
								event := HTTPEvent{
									Timestamp:   time.Now(),
									SrcIP:       srcIP,
									Method:      method,
									URI:         uri,
									Headers:     headers,
									PayloadSize: len(payload),
								}
								sendHTTPEvent(event)
								// Call monitorHTTPRequest to track and detect high frequency requests
								monitorHTTPRequest(method, uri, headers, len(payload))
							}
						}
					}
				}
			}
		}
	}
}

func main() {
	watchDir := os.Getenv("WATCH_DIR")
	if watchDir == "" {
		watchDir = "/var/www/html"
	}

	log.Println("SecureShield Agent started")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start controlled file monitoring
	watchCtx, _ = context.WithCancel(ctx)
	go watchFilesControlled(watchCtx, watchDir)

	// Start controlled reset HTTP request counts
	resetCtx, _ = context.WithCancel(ctx)
	go resetRequestCountsControlled(resetCtx)

	// Variables to control capturePackets goroutine
	var captureCtx context.Context

	http.HandleFunc("/activate", func(w http.ResponseWriter, r *http.Request) {
		setMonitoringActive(true)
		log.Println("Monitoring activated")

		// Start capturePackets if not already running
		if captureCancelFunc == nil {
			captureCtx, captureCancelFunc = context.WithCancel(ctx)
			go capturePacketsControlled(captureCtx)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"agent monitoring activated"}`))
	})

	http.HandleFunc("/deactivate", func(w http.ResponseWriter, r *http.Request) {
		setMonitoringActive(false)
		log.Println("Monitoring deactivated")

		// Stop capturePackets if running
		if captureCancelFunc != nil {
			captureCancelFunc()
			captureCancelFunc = nil
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"agent monitoring deactivated"}`))
	})

	http.HandleFunc("/status", statusHandler)

	go func() {
		log.Println("Starting HTTP server for agent control on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	for {
		time.Sleep(10 * time.Second)
		if isMonitoringActive() {
			log.Println("Agent heartbeat - monitoring...")
		} else {
			log.Println("Agent heartbeat - monitoring paused")
		}
	}
}

func stopAll() {
	if captureCancelFunc != nil {
		captureCancelFunc()
		captureCancelFunc = nil
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := "inactive"
	if isMonitoringActive() {
		status = "active"
	}
	w.Write([]byte(`{"status":"` + status + `"}`))
}
