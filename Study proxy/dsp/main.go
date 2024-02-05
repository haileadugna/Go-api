package main

import "mymodule/mymodule/connection_handler"
import "mymodule/statistics"
import "mymodule/backend_manager"
import "mymodule/rpc_server"

import (
	"log"
	"net"
	// "net/http"
	"sync"
)


var requestBytes map[string]int64
var requestLock sync.Mutex

func main() {
	// Your main function code here
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
