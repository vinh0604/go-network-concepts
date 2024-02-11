package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"
)

func main() {
	ALLOWED_METHODS := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	METHODS_WITH_PAYLOAD := []string{"POST", "PUT", "PATCH", "DELETE"}
	var port int
	flag.IntVar(&port, "p", 80, "Port to connect to")
	var method string = "GET"
	flag.StringVar(&method, "X", "GET", "HTTP Method")
	method = strings.ToUpper(method)
	var payload string
	flag.StringVar(&payload, "d", "", "Payload")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		panic("Not enough arguments")
	}
	host := args[0]

	if !slices.Contains(ALLOWED_METHODS, method) {
		panic(fmt.Sprintf("Method %s not allowed", method))
	}

	var err error
	sock, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	if slices.Contains(METHODS_WITH_PAYLOAD, method) && payload != "" {
		sock.Write([]byte(
			fmt.Sprintf("%s / HTTP/1.1\r\nHost: %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", method, host, len(payload), payload)))
	} else {
		sock.Write([]byte(
			fmt.Sprintf("%s / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", method, host)))
	}

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
