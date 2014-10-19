package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
)

// Global variables
var CACHE = make(map[string]string)

// Commands
var GET string = "get"
var SET string = "set"
var DELETE string = "delete"
var QUIT string = "quit"
var STATS string = "stats"

// Responses
var VALUE string = "VALUE"
var END string = "END"
var STORED string = "STORED"
var DELETED string = "DELETED"
var NOT_FOUND string = "NOT_FOUND"
var ERROR string = "ERROR"

var KEY_LIMIT int = 255   // # of characters
var DATA_LIMIT int = 8192 // 8kb
var LINE_ENDING = "\r\n"

// Options
var DEFAULT_PORT int = 11212
var DEFAULT_ITEMS int = 65535

func main() {
	port := flag.Int("port", DEFAULT_PORT, "The port the server listens on")
	items := flag.Int("item", DEFAULT_ITEMS, "Total number of items the server can store")
	flag.Parse()

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}
		go handleConnection(conn, *items)
	}
}

func handleConnection(c net.Conn, items int) {
	buf := make([]byte, items)
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
			CACHE[currentKey] = message
			isExpectingCommand = false
			c.Write([]byte(STORED + LINE_ENDING))
			continue
		}

		commandPieces := strings.Split(message, " ")
		command := commandPieces[0]

		if command == GET {
			response := ""
			for i := range commandPieces[1:] {
				key := commandPieces[i+1]
				_, exists := CACHE[key]
				if exists {
					response += VALUE + " " + key + LINE_ENDING
					response += CACHE[key] + LINE_ENDING
					stats["get_hits"] += 1
				} else {
					stats["get_msses"] += 1
				}
				stats["cmd_get"] += 1
			}
			response += END + LINE_ENDING
			c.Write([]byte(response))
		} else if command == SET {
			currentKey = commandPieces[1]
			if len(currentKey) >= KEY_LIMIT {
				c.Write([]byte(ERROR + LINE_ENDING))
				continue
			}
			isExpectingCommand = true
			stats["cmd_set"] += 1
		} else if command == DELETE {
			key := commandPieces[1]
			_, exists := CACHE[key]
			if exists {
				delete(CACHE, key)
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
			response += "curr_items " + strconv.Itoa(len(CACHE)) + LINE_ENDING
			response += "limits_items " + strconv.Itoa(DEFAULT_ITEMS) + LINE_ENDING
			response += END + LINE_ENDING
			c.Write([]byte(response))
		} else if command == QUIT {
			err = c.Close()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			c.Write([]byte(ERROR + " NO CAMMAND FOUND" + LINE_ENDING))
		}
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
}
