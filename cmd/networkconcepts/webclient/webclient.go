package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		panic("Not enough arguments")
	}

	host := args[0]
	port := 80
	var err error
	if len(args) > 1 {
		port, err = strconv.Atoi(args[1])
		if err != nil {
			panic(err)
		}
	}

	sock, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	sock.Write([]byte(
		fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", host)))
	var sb strings.Builder
	buffer := make([]byte, 4096)
	for {
		mLen, err := sock.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading:", err.Error())
			}
			break
		}

		sb.Write(buffer[:mLen])
	}

	fmt.Println(sb.String())
	defer sock.Close()
}
