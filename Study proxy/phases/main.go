// package main

// import (
// 	"bufio"
// 	"io"
// 	"log"
// 	"net"
// 	"net/http"
// 	"strconv"
// 	"sync"
// 	"time"
// )

// type Backend struct {
// 	net.Conn
// 	Reader *bufio.Reader
// 	Writer *bufio.Writer
// }

// var backendQueue chan *Backend
// var requestBytes map[string]int64
// var requestLock sync.Mutex

// func init() {
// 	backendQueue = make(chan *Backend, 10)
// 	requestBytes = make(map[string]int64)
// }

// func getBackend() (*Backend, error) {
// 	select {
// 	case be := <-backendQueue:
// 		return be, nil
// 	case <-time.After(100 * time.Millisecond):
// 		be, err := net.Dial("tcp", "127.0.0.1:8081")
// 		if err != nil {
// 			return nil, err
// 		}
// 		return &Backend{
// 			Conn:   be,
// 			Reader: bufio.NewReader(be),
// 			Writer: bufio.NewWriter(be),
// 		}, nil
// 	}
// }

// func queueBackend(be *Backend) {
// 	select {
// 	case backendQueue <- be:
// 	case <-time.After(1 * time.Second):
// 		be.Close()
// 	}
// }

// func handleRequest(conn net.Conn, req *http.Request) {
// 	be, err := getBackend()
// 	if err != nil {
// 		log.Printf("Failed to get backend: %v", err)
// 		return
// 	}
// 	defer queueBackend(be)

// 	if err := req.Write(be); err != nil {
// 		log.Printf("Failed to write request to backend: %v", err)
// 		return
// 	}

// 	resp, err := http.ReadResponse(be.Reader, req)
// 	if err != nil {
// 		log.Printf("Failed to read response from backend: %v", err)
// 		return
// 	}

// 	requestLock.Lock()
// 	bytes := requestBytes[req.URL.Path] + resp.ContentLength
// 	requestBytes[req.URL.Path] = bytes
// 	requestLock.Unlock()

// 	resp.Header.Set("X-Bytes", strconv.FormatInt(bytes, 10))

// 	if err := resp.Write(conn); err == nil {
// 		log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
// 	}
// }

// func acceptConnection(ln net.Listener) {
// 	for {
// 		conn, err := ln.Accept()
// 		if err != nil {
// 			log.Fatalf("Error accepting connection: %s", err)
// 		}
// 		go handleConnection(conn)
// 	}
// }

// func handleConnection(conn net.Conn) {
// 	defer conn.Close()

// 	reader := bufio.NewReader(conn)

// 	for {
// 		req, err := http.ReadRequest(reader)
// 		if err != nil {
// 			if err == io.EOF {
// 				log.Printf("Failed to read request: %v", err)
// 			}
// 			return
// 		}

// 		handleRequest(conn, req)
// 	}
// }

// func main() {
// 	// Start the backend server (for demonstration purposes)
// 	go func() {
// 		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 			w.Write([]byte("Hello from backend hi!"))
// 		})
// 		log.Fatal(http.ListenAndServe(":8081", nil))
// 	}()

// 	// Start the proxy server
// 	ln, err := net.Listen("tcp", ":8080")
// 	if err != nil {
// 		log.Fatalf("Error starting server: %s", err)
// 	}
// 	defer ln.Close()

// 	acceptConnection(ln)
// }
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

type Empty struct{}
type Stats struct {
	RequestBytes map[string]int64
}

type RpcServer struct{}

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

func updateStats(req *http.Request, resp *http.Response) int64 {
	requestLock.Lock()
	defer requestLock.Unlock()

	bytes := requestBytes[req.URL.Path] + resp.ContentLength
	requestBytes[req.URL.Path] = bytes
	return bytes
}

func handleRequest(conn net.Conn) {
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

		bytes := updateStats(req, resp)
		resp.Header.Set("X-Bytes", strconv.FormatInt(bytes, 10))

		if err := resp.Write(conn); err != nil {
			log.Printf("Failed to write response to client: %v", err)
			return
		}

		log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
	}
}

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
	rpc.Register(&RpcServer{})
	rpc.HandleHTTP()
	go func() {
		l, err := net.Listen("tcp", ":8079")
		if err != nil {
			log.Fatalf("Error starting RPC server: %v", err)
		}
		log.Fatal(http.Serve(l, nil))
	}()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}

	for {
		if conn, err := ln.Accept(); err == nil {
			go handleRequest(conn)
		}
	}
}
