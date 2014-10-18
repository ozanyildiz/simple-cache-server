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
var DELETE string = "delete"

var LINE_ENDING = "\r\n"

// Responses
var VALUE string = "VALUE"
var END string = "END"
var STORED string = "STORED"
var DELETED string = "DELETED"

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
	isExpectingCommand := false
	currentKey := ""

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

		if isExpectingCommand {
			cache[currentKey] = message
			isExpectingCommand = false
			c.Write([]byte(STORED + LINE_ENDING))
			continue
		}

		commandPieces := strings.Split(message, " ")
		log.Printf("%s", commandPieces)
		command := commandPieces[0]

		if command == GET {
			response := VALUE
			for i := range commandPieces[1:] {
				key := commandPieces[i+1]
				response += key + LINE_ENDING + cache[key] + LINE_ENDING
			}
			response += END + LINE_ENDING
			c.Write([]byte(response))
		} else if command == SET {
			currentKey = commandPieces[1]
			isExpectingCommand = true
		} else if command == DELETE {
			key := commandPieces[1]
			_, exists := cache[key]
			if exists {
				delete(cache, key)
			} else {
				c.Write([]byte(DELETED + LINE_ENDING))
			}
		}
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
}
