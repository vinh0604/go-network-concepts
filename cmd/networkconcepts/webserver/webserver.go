package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]

	port := 8080
	var err error
	if len(args) > 1 {
		port, err = strconv.Atoi(args[0])
		if err != nil {
			panic(err)
		}
	}

	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := sock.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Remote Address: %s %s\n", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	fmt.Printf("Local Address: %s %s\n", conn.LocalAddr().Network(), conn.LocalAddr().String())

	buffer := make([]byte, 4096)
	var sb strings.Builder
	var last3Bytes []byte = []byte{}
	for {
		mLen, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading:", err.Error())
			}
			break
		}

		sb.Write(buffer[:mLen])
		if bytes.Contains(append(last3Bytes, buffer[:mLen]...), []byte("\r\n\r\n")) {
			break
		}

		if mLen >= 3 {
			last3Bytes = buffer[mLen-3:]
		}
	}
	fmt.Println(sb.String())

	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!"))
}
