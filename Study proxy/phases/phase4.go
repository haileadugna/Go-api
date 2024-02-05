package main

import (
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


// server side
func main() {
	rpc.Register(&RpcServer{})
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":8079")
	if err != nil {
		log.Fatalf("Error starting RPC server: %v", err)
	}
	go http.Serve(l, nil)
}

// client side
// func main() {
// 	client, err := rpc.DialHTTP("tcp", "localhost:8079")
// 	if err != nil {
// 		log.Fatalf("Error connecting to RPC server: %v", err)
// 	}
	
// 	var reply Stats
// 	err = client.Call("RpcServer.GetStats", &Empty{}, &reply)
// 	if err != nil {
// 		log.Fatalf("Error calling RPC server: %v", err)
// 	}
// }