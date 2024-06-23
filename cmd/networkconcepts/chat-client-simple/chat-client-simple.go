package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/vinh0604/go-network-concepts/internal/chatmodels"
)

func main() {
	var err error

	args := flag.Args()
	host := "localhost"
	port := 8080
	if len(args) > 2 {
		host = args[0]
		port, err = strconv.Atoi(args[1])
		if err != nil {
			panic(err)
		}
	}

	sock, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}
	defer sock.Close()

	nick := "vinh"
	payload := chatmodels.Payload{
		MsgType: chatmodels.MsgTypeHello,
		Nick:    &nick,
	}

	out, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	outLen := len(out)
	outLenBytes := []byte{
		byte(outLen >> 8),
		byte(outLen & 0xFF),
	}

	sock.Write(append(outLenBytes, out...))
	time.Sleep(2 * time.Second)

	msg := "Hello, everyone!"
	payload = chatmodels.Payload{
		MsgType: chatmodels.MsgTypeChat,
		Msg:     &msg,
	}
	out, err = json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	outLen = len(out)
	outLenBytes = []byte{
		byte(outLen >> 8),
		byte(outLen & 0xFF),
	}
	sock.Write(append(outLenBytes, out...))
}
