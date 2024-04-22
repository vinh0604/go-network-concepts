package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

const WORD_LENGTH_SIZE = 2

func main() {
	var err error
	if len(os.Args) < 2 {
		fmt.Println("usage: wordclient <port>")
		return
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	sock, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer
	for {
		packet, err := getNextWordPacket(&buffer, sock)
		if err != nil {
			if err != io.EOF {
				panic(err)
			}

			fmt.Println("Connection closed")
			break
		}

		fmt.Println("Word:", extractWord(packet))
	}
}

func getNextWordPacket(buffer *bytes.Buffer, sock net.Conn) ([]byte, error) {
	if buffer.Len() >= WORD_LENGTH_SIZE {
		wordLen := binary.BigEndian.Uint16(buffer.Bytes()[0:WORD_LENGTH_SIZE])
		if buffer.Len() >= int(wordLen)+WORD_LENGTH_SIZE {
			output := buffer.Bytes()[:int(wordLen)+WORD_LENGTH_SIZE]
			buffer.Next(int(wordLen) + WORD_LENGTH_SIZE)
			return output, nil
		} else {
			err := receive(buffer, sock)
			if err != nil {
				return nil, err
			}
			return getNextWordPacket(buffer, sock)
		}
	} else {
		err := receive(buffer, sock)
		if err != nil {
			return nil, err
		}
		return getNextWordPacket(buffer, sock)
	}
}

func receive(buffer *bytes.Buffer, sock net.Conn) error {
	readBuffer := make([]byte, 5)
	mLen, err := sock.Read(readBuffer)
	if err != nil {
		return err
	}
	buffer.Write(readBuffer[:mLen])
	return nil
}

func extractWord(packet []byte) string {
	return string(packet[WORD_LENGTH_SIZE:])
}
