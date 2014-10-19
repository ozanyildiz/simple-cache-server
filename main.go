package main

import (
	"log"
	"net"
	"strconv"
	"strings"
)

var cache = make(map[string]string)

// Commands
var GET string = "get"
var SET string = "set"
var DELETE string = "delete"
var QUIT string = "quit"
var STATS string = "stats"

var LINE_ENDING = "\r\n"

// Responses
var VALUE string = "VALUE"
var END string = "END"
var STORED string = "STORED"
var DELETED string = "DELETED"
var NOT_FOUND string = "NOT_FOUND"

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

	var stats = make(map[string]int)
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
			response := ""
			for i := range commandPieces[1:] {
				key := commandPieces[i+1]
				_, exists := cache[key]
				if exists {
					response += VALUE + " " + key + LINE_ENDING
					response += cache[key] + LINE_ENDING
					stats["get_hits"] += 1
				} else {
					stats["get_misses"] += 1
				}
				stats["cmd_get"] += 1
			}
			response += END + LINE_ENDING
			c.Write([]byte(response))
		} else if command == SET {
			currentKey = commandPieces[1]
			isExpectingCommand = true
			stats["cmd_set"] += 1
		} else if command == DELETE {
			key := commandPieces[1]
			_, exists := cache[key]
			if exists {
				delete(cache, key)
				c.Write([]byte(DELETED + LINE_ENDING))
				stats["delete_hits"] += 1
			} else {
				c.Write([]byte(NOT_FOUND))
				stats["delete_misses"] += 1
			}
		} else if command == STATS {
			response := "cmd_get " + strconv.Itoa(stats["cmd_get"]) + LINE_ENDING
			response += "cmd_set " + strconv.Itoa(stats["cmd_set"]) + LINE_ENDING
			response += "get_hits " + strconv.Itoa(stats["get_hits"]) + LINE_ENDING
			response += "get_misses " + strconv.Itoa(stats["get_misses"]) + LINE_ENDING
			response += "delete_hits " + strconv.Itoa(stats["delete_hits"]) + LINE_ENDING
			response += "delete_misses " + strconv.Itoa(stats["delete_misses"]) + LINE_ENDING
			response += "curr_items " + strconv.Itoa(len(cache)) + LINE_ENDING
			response += "limits_items " + "not implemented" + LINE_ENDING
			response += END + LINE_ENDING
			c.Write([]byte(response))
		} else if command == QUIT {
			err = c.Close()
			if err != nil {
				// error handling
			}
		}
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
}
