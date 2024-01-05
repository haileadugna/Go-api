package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("Error accepting connection: %v", err)
            continue
        }

        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    reader := bufio.NewReader(conn)
    _, err := http.ReadRequest(reader)
	if err != nil {
		if err == io.EOF {
			log.Printf("Client disconnected before sending request")
		} else {
			log.Printf("Error reading request: %v", err)
		}
		return
	}

    // Add your proxy logic here
}