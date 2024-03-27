package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	var err error
	currDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var rootDir string
	flag.StringVar(&rootDir, "d", currDir, "Serving directory")
	flag.Parse()
	fmt.Printf("Serving directory: %s\n", rootDir)
	if rootInfo, err := os.Stat(rootDir); os.IsNotExist(err) {
		panic(err)
	} else if !rootInfo.IsDir() {
		panic("Root path is not a directory")
	}
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		panic(err)
	}

	args := flag.Args()
	port := 8080
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
		go handleConnection(conn, rootDir)
	}
}

func handleConnection(conn net.Conn, rootDir string) {
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
	filePath, err := url.QueryUnescape(requestPath[1:])

	if err != nil {
		fmt.Printf("Error decoding URI: %v\n", err)
	}
	if filePath == "" {
		filePath = "."
	}
	fmt.Printf("File requested: %s\n", filePath)

	var errorMessage string
	var responseCode string
	var contentType string
	var responseBody []byte
	filePath, err = filepath.Abs(filePath)
	if err != nil {
		errorMessage = fmt.Sprintf("Internal Server Error: %s", err.Error())
		responseCode = "500 Internal Server Error"
	} else if !strings.HasPrefix(filePath, rootDir) {
		errorMessage = fmt.Sprintf("Access denied: %s", filePath)
		responseCode = "403 Forbidden"
	} else if fileInfo, err := os.Stat(filePath); os.IsNotExist(err) {
		errorMessage = fmt.Sprintf("File not found: %s", filePath)
		responseCode = "404 Not Found"
	} else if err != nil {
		errorMessage = fmt.Sprintf("Internal Server Error: %s", err.Error())
		responseCode = "500 Internal Server Error"
	} else {
		if fileInfo.IsDir() {
			renderDir(rootDir, filePath, &responseBody, &errorMessage, &responseCode, &contentType)
		} else {
			renderFile(filePath, &responseBody, &errorMessage, &responseCode, &contentType)
		}

	}

	if errorMessage != "" {
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", responseCode, len(errorMessage), errorMessage)))
	} else {
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", contentType, len(responseBody), responseBody)))
	}
}

func renderDir(rootDir string, dirPath string, responseBody *[]byte, errorMessage *string, responseCode *string, contentType *string) {
	relPath, err := filepath.Rel(rootDir, dirPath)
	if err != nil {
		*errorMessage = fmt.Sprintf("Internal Server Error: %s", err.Error())
		*responseCode = "500 Internal Server Error"
		return
	}
	files, err := os.ReadDir(dirPath)
	if err != nil {
		*errorMessage = fmt.Sprintf("Internal Server Error: %s", err.Error())
		*responseCode = "500 Internal Server Error"
	}

	var sb strings.Builder
	sb.WriteString("<html><head><title>Directory Listing</title></head><body><h1>Directory Listing</h1><ul>")
	for _, file := range files {
		sb.WriteString(fmt.Sprintf("<li><a href=\"/%s/%s\">%s</a></li>", relPath, file.Name(), file.Name()))
	}

	sb.WriteString("</ul></body></html>")
	*contentType = "text/html; charset=utf-8"
	*responseBody = []byte(sb.String())
}

func renderFile(filePath string, responseBody *[]byte, errorMessage *string, responseCode *string, contentType *string) {
	var content, err = os.ReadFile(filePath)
	if err != nil {
		*errorMessage = fmt.Sprintf("Internal Server Error: %s", err.Error())
		*responseCode = "500 Internal Server Error"
	} else {
		if strings.HasSuffix(filePath, ".html") || strings.HasSuffix(filePath, ".htm") {
			*contentType = "text/html; charset=utf-8"
			*responseBody = content
		} else if strings.HasSuffix(filePath, ".txt") || strings.HasSuffix(filePath, ".log") || strings.HasSuffix(filePath, ".csv") || strings.HasSuffix(filePath, ".md") {
			*contentType = "text/plain; charset=utf-8"
			*responseBody = content
		} else if strings.HasSuffix(filePath, ".jpg") {
			*contentType = "image/jpeg"
			*responseBody = content
		} else if strings.HasSuffix(filePath, ".png") {
			*contentType = "image/png"
			*responseBody = content
		} else if strings.HasSuffix(filePath, ".pdf") {
			*contentType = "application/pdf"
			*responseBody = content
		} else {
			*errorMessage = fmt.Sprintf("File content is not supported: %s", filePath)
			*responseCode = "400 Bad Request"
		}
	}
}
