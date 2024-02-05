package main

import (
    "bufio"
    "net"
    "net/http"
    "strconv"
    "sync"
)

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