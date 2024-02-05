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
)

func main() {
	rpc.Register(new(RpcServer))
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":8079")
	if err != nil {
		log.Fatalf("Error starting RPC server: %v", err)
	}
	go http.Serve(l, nil)

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
	for {
		if conn, err := ln.Accept(); err == nil {
			go handleConnection(conn)
		}
	}
}

var requestBytes map[string]int64
var requestLock sync.Mutex

func init() {
	requestBytes = make(map[string]int64)
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
		if be, err := net.Dial("tcp", "google.com"); err == nil {
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

type Empty struct{}
type Stats struct {
	RequestBytes map[string]int64
}

type RpcServer struct{}

func (r *RpcServer) GetStats(args *Empty, reply *Stats) error {
	requestLock.Lock()
	defer requestLock.Unlock()

	reply.RequestBytes = make(map[string]int64)
	for k, v := range requestBytes {
		reply.RequestBytes[k] = v
	}
	return nil
}

func updateStats(req *http.Request, resp *http.Response) int64 {
	requestLock.Lock()
	defer requestLock.Unlock()

	bytes := requestBytes[req.URL.Path] + resp.ContentLength
	requestBytes[req.URL.Path] = bytes
	return bytes
}

func handleRequest(be_reader *bufio.Reader, conn net.Conn, req *http.Request) {
	if resp, err := http.ReadResponse(be_reader, req); err == nil {
		bytes := updateStats(req, resp)
		resp.Header.Set("X-Bytes", strconv.FormatInt(bytes, 10))

		if err := resp.Write(conn); err == nil {
			log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
		}
	}
}
