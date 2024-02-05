package connection_handler

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
)
import "mymodule/mymodule/backend_manager"

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

		backend_manager.HandleRequest(conn, req)
	}
}
