package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func getDevice() string {
	dev := os.Getenv("CAPTURE_DEVICE")
	if dev == "" {
		return "eth0"
	}
	return dev
}

func capturePackets() {
	device := getDevice() // Get device from env or default
	snapshotLen := int32(1024)
	promiscuous := false
	timeout := 30 * time.Second

	handle, err := pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
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
	for packet := range packetSource.Packets() {
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
