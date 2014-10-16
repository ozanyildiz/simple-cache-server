package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// error handling
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// error handling
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		n, err = c.Write(buf[0:n])
		if err != nil {
			c.Close()
			break
		}
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
}
