package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"
)

// Part I: Basic Proxy Design and Implementation

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
	if r.URL.Path == "/" {
		log.Println("Received HTTP request, redirecting to: https://go.dev/doc/")
	}
	http.Redirect(w, r, "https://go.dev/doc/", http.StatusTemporaryRedirect)
}

// Part III: Statistics Gathering Functionality

type Empty struct{}
type Stats struct {
	RequestBytes map[string]int64
}

type RpcServer struct{}

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

// Part IV: Optimizing Backend Connections

type Backend struct {
	net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

var backendQueue chan *Backend
var requestLock sync.Mutex

func init() {
	backendQueue = make(chan *Backend, 10)
	requestBytes = make(map[string]int64)
}

func getBackend() (*Backend, error) {
	select {
	case be := <-backendQueue:
		return be, nil
	case <-time.After(100 * time.Millisecond):
		be, err := net.Dial("tcp", "127.0.0.1:8081")
		if err != nil {
			return nil, err
		}
		return &Backend{
			Conn:   be,
			Reader: bufio.NewReader(be),
			Writer: bufio.NewWriter(be),
		}, nil
	}
}

func queueBackend(be *Backend) {
	select {
	case backendQueue <- be:
	case <-time.After(1 * time.Second):
		be.Close()
	}
}

func handleRequest(conn net.Conn, req *http.Request) {
	be, err := getBackend()
	if err != nil {
		log.Printf("Failed to get backend: %v", err)
		return
	}
	defer queueBackend(be)

	if err := req.Write(be); err != nil {
		log.Printf("Failed to write request to backend: %v", err)
		return
	}

	resp, err := http.ReadResponse(be.Reader, req)
	if err != nil {
		log.Printf("Failed to read response from backend: %v", err)
		return
	}

	requestLock.Lock()
	bytes := requestBytes[req.URL.Path] + resp.ContentLength
	requestBytes[req.URL.Path] = bytes
	requestLock.Unlock()

	resp.Header.Set("X-Bytes", strconv.FormatInt(bytes, 10))

	if err := resp.Write(conn); err == nil {
		log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
	}
}

// Part V: RPCs for Communication Between Proxy and Backend

// Code for RPC server and client remains the same as in Part III

// Note: Make sure to import necessary packages like "strconv" and "time" at the beginning of the code.
