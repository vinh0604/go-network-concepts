package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	var err error
	sock, err := net.Dial("tcp", "time.nist.gov:37")

	if err != nil {
		panic(err)
	}

	var bytesBuffer bytes.Buffer
	buffer := make([]byte, 4096)
	for {
		mLen, err := sock.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading:", err.Error())
			}
			break
		}

		bytesBuffer.Write(buffer[:mLen])
	}

	bytesReceived := bytesBuffer.Bytes()
	if len(bytesReceived) != 4 {
		panic("Invalid response from server")
	}

	epochFrom1900 := binary.BigEndian.Uint32(bytesReceived)
	secondsDelta := 2208988800
	fmt.Println("NIST time :", epochFrom1900)
	fmt.Println("System time :", time.Now().Unix()+int64(secondsDelta))
}
