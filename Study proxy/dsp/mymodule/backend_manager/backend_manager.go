package backend_manager

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Backend struct {
	net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

var backendQueue chan *Backend
var requestBytes map[string]int64
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
