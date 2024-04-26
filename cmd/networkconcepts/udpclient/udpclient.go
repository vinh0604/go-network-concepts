package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

const MAX_PACKET_SIZE = 512

func main() {
	var err error
	if len(os.Args) < 4 {
		fmt.Println("usage: udpclient <host> <port> <data>")
		return
	}

	host := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	data := os.Args[3]

	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataBytes := []byte(data)

	for start := 0; start < len(dataBytes); start += MAX_PACKET_SIZE {
		end := start + MAX_PACKET_SIZE
		if end > len(dataBytes) {
			end = len(dataBytes)
		}
		conn.Write(dataBytes[start:end])
	}
}
