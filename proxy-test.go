package main

import (
	"io"
	"log"
	"net"
	"time"
)

// For the PoC, we hardcode an expiration date. 
// In production, this would read from the OEM license file or service.
var licenseExpiration = time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

func main() {
	// 1. Listen on a custom port (the "Front Door" for OEM)
	listener, err := net.Listen("tcp", ":7688")
	if err != nil {
		log.Fatalf("Failed to bind to port: %v", err)
	}
	log.Println("Neo4j License Proxy listening on :7688...")

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle each incoming connection in a lightweight goroutine
		go handleConnection(clientConn)
	}
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// 2. License Validation Check (The "Bouncer")
	if time.Now().After(licenseExpiration) {
		log.Println("Connection rejected: OEM license has expired.")
		// MVP: Immediately close the TCP socket. 
		// (A more advanced version could write a mock Bolt fatal error frame here before closing).
		return
	}

	log.Println("License valid. Forwarding connection to Neo4j...")

	// 3. Dial the actual externalized Neo4j backend
	neo4jConn, err := net.Dial("tcp", "localhost:7687")
	if err != nil {
		log.Printf("Failed to connect to Neo4j backend: %v", err)
		return
	}
	defer neo4jConn.Close()

	// 4. Layer 4 Bidirectional Streaming
	// This streams bytes blindly without parsing the Bolt protocol or Cypher queries.
	done := make(chan struct{})

	// Stream: Client -> Neo4j
	go func() {
		io.Copy(neo4jConn, clientConn)
		done <- struct{}{}
	}()

	// Stream: Neo4j -> Client
	go func() {
		io.Copy(clientConn, neo4jConn)
		done <- struct{}{}
	}()

	// Wait for one of the streams to close (e.g., client disconnects)
	<-done
	log.Println("Connection closed.")
}