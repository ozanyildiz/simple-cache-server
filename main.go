package main

import (
	"log"
	"net"
	"strings"
)

var cache = make(map[string]string)

// Commands
var GET string = "get"
var SET string = "set"

var LINE_ENDING = "\r\n"

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

		message := strings.TrimSuffix(string(buf[0:n]), LINE_ENDING)
		commandPieces := strings.Split(message, " ")
		log.Printf("%s", commandPieces)
		command := commandPieces[0]
		if command == GET {
			key := commandPieces[1]
			log.Printf("Get command. cache[%s] = %s", key, cache[key])
		} else if command == SET {
			key := commandPieces[1]
			value := commandPieces[2]
			cache[key] = value
			log.Printf("Set command. cache[%s] = %s", key, cache[key])
		}
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
}
