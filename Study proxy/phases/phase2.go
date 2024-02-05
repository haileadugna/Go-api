package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
)

var requestBytes map[string]int64
var requestLock sync.Mutex

func init() {
	requestBytes = make(map[string]int64)
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
