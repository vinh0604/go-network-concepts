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
	requestHeaders := sb.String()
	var firstHeaderLine = strings.Split(requestHeaders, "\r\n")[0]
	requestMethod := firstHeaderLine[:strings.Index(firstHeaderLine, " ")+1]
	fmt.Printf("Request method: %s\n", requestMethod)

	var requestPath = firstHeaderLine[strings.Index(firstHeaderLine, " ")+1 : strings.LastIndex(firstHeaderLine, " ")]
	var pathComponents = strings.Split(requestPath, "/")
	var filePath = pathComponents[len(pathComponents)-1]

	var errorMessage string
	var responseCode string
	var contentType string
	var responseBody []byte
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errorMessage = fmt.Sprintf("File not found: %s", filePath)
		responseCode = "404 Not Found"
	} else if err != nil {
		errorMessage = fmt.Sprintf("Internal Server Error: %s", err.Error())
		responseCode = "500 Internal Server Error"
	} else {
		var content, err = os.ReadFile(filePath)
		if err != nil {
			errorMessage = fmt.Sprintf("Internal Server Error: %s", err.Error())
			responseCode = "500 Internal Server Error"
		} else {
			if strings.HasSuffix(filePath, ".html") {
				contentType = "text/html"
				responseBody = content
			} else if strings.HasSuffix(filePath, ".txt") {
				contentType = "text/plain"
				responseBody = content
			} else {
				errorMessage = fmt.Sprintf("File content is not supported: %s", filePath)
				responseCode = "400 Bad Request"
			}
		}
	}

	if errorMessage != "" {
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", responseCode, len(errorMessage), errorMessage)))
	} else {
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s; charset=utf-8\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", contentType, len(responseBody), responseBody)))
	}
}
