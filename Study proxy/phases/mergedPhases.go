package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

type Empty struct{}
type Stats struct {
	RequestBytes map[string]int64
}

type RpcServer struct{}

var requestLock sync.Mutex
var requestBytes map[string]int64

func (r *RpcServer) GetStats(args *Empty, reply *Stats) error {
	requestLock.Lock()
	defer requestLock.Unlock()

	reply.RequestBytes = make(map[string]int64)
	for k, v := range requestBytes {
		reply.RequestBytes[k] = v
	}
	return nil
}

func main() {
	// Start the RPC server
	rpc.Register(&RpcServer{})
	rpc.HandleHTTP()
	go func() {
		if err := http.ListenAndServe(":8079", nil); err != nil {
			log.Fatalf("Error starting RPC server: %v", err)
		}
	}()

	// Start the HTTP server for handling requests on :8080
	http.HandleFunc("/", handleHTTP)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	// Start the TCP server to handle backend connections
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
	for {
		if conn, err := ln.Accept(); err == nil {
			go handleConnection(conn)
		}
	}
}

func handleConnection(conn net.Conn) {
	// The existing handleConnection logic...
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("Failed to read request: %v", err)
			}
			return
		}

		// Connect to a backend and send the request along.
		if be, err := net.Dial("tcp", "127.0.0.1:8081"); err == nil {
			be_reader := bufio.NewReader(be)
			if err := req.Write(be); err == nil {
				if resp, err := http.ReadResponse(be_reader, req); err == nil {
					if err := resp.Write(conn); err == nil {
						log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
					}
				}
			}
		}
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle HTTP requests on http://127.0.0.1:8080
	// You can serve Go documentation or any other content here.
	// For example, you can use the http.ServeFile function to serve an HTML file.
	http.ServeFile(w, r, "path/to/your/documentation.html")
}
