package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	var err error
	if len(os.Args) < 2 {
		fmt.Println("usage: udpserver <port>")
		return
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for {
		buffer := make([]byte, 4096)
		size, remoteAddr, err := conn.ReadFrom(buffer)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Remote Address: %s %s\n", remoteAddr.Network(), remoteAddr.String())
		fmt.Println("Received:", string(buffer[:size]))

		_, err = conn.WriteTo([]byte("ACK"), remoteAddr)
		if err != nil {
			panic(err)
		}
	}
}
