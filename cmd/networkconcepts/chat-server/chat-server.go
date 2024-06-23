package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/vinh0604/go-network-concepts/internal/chatmodels"
)

const PAYLOAD_LEN_SIZE = 2

type clientInfo struct {
	conn         *net.Conn
	chatPayload  *chatmodels.Payload
	disconnected bool
}

type chatBuffer struct {
	bufLen int
	buf    []byte
}

func main() {
	var err error

	args := flag.Args()
	port := 8080
	if len(args) > 1 {
		port, err = strconv.Atoi(args[0])
		if err != nil {
			panic(err)
		}
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		clientCh := make(chan clientInfo)
		go handleConn(conn, clientCh)

		go func() {
			clients := map[net.Conn]string{}
			for {
				client := <-clientCh
				if client.disconnected {
					if nick, ok := clients[*client.conn]; ok {
						fmt.Printf("Client %s (nick=%s) left.\n", (*client.conn).RemoteAddr().String(), nick)
						delete(clients, *client.conn)
					}
					continue
				}

				if client.chatPayload != nil {
					if client.chatPayload.MsgType == chatmodels.MsgTypeHello {
						fmt.Printf("Client %s (nick=%s) joined.\n", (*client.conn).RemoteAddr().String(), *client.chatPayload.Nick)
						clients[*client.conn] = *client.chatPayload.Nick
						announce := chatmodels.Payload{
							MsgType: chatmodels.MsgTypeJoin,
							Nick:    client.chatPayload.Nick,
						}
						go relay(*client.chatPayload.Nick, client.conn, &clients, announce)
					} else if client.chatPayload.MsgType == chatmodels.MsgTypeChat {
						nick, ok := clients[*client.conn]
						if !ok {
							fmt.Printf("Client %s not registered\n", (*client.conn).RemoteAddr().String())
							continue
						}

						fmt.Printf("Client %s (nick=%s) sent a message.\n", (*client.conn).RemoteAddr().String(), nick)
						chat := chatmodels.Payload{
							MsgType: chatmodels.MsgTypeChat,
							Nick:    &nick,
							Msg:     client.chatPayload.Msg,
						}
						go relay(nick, client.conn, &clients, chat)
					} else {
						fmt.Printf("Client %s sent an unknown message type: %s\n", (*client.conn).RemoteAddr().String(), client.chatPayload.MsgType)
					}
				}
			}
		}()
	}
}

func relay(nick string, clientConn *net.Conn, clients *map[net.Conn]string, payload chatmodels.Payload) {
	jsonStr, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to relay %s message from %s: %s\n", payload.MsgType, nick, err)
		return
	}

	payloadLen := len(jsonStr)
	outBytes := []byte{
		byte(payloadLen >> 8),
		byte(payloadLen & 0xFF),
	}
	outBytes = append(outBytes, jsonStr...)
	for conn := range *clients {
		if conn != *clientConn {
			conn.Write(outBytes)
		}
	}
}

func handleConn(conn net.Conn, clientCh chan clientInfo) {
	defer conn.Close()

	chatBuf := chatBuffer{}
	for {
		payload, err := readNextMessage(conn, &chatBuf)

		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading:", err.Error())
			} else {
				clientCh <- clientInfo{
					conn:         &conn,
					disconnected: true,
				}
			}
			break
		}

		if payload.MsgType == chatmodels.MsgTypeHello {
			if payload.Nick != nil {
				clientCh <- clientInfo{
					conn:         &conn,
					chatPayload:  payload,
					disconnected: false,
				}
			} else {
				fmt.Printf("Client %s sent a hello message without a nickname\n", conn.RemoteAddr().String())
			}
		} else if payload.MsgType == chatmodels.MsgTypeChat {
			if payload.Msg != nil {
				clientCh <- clientInfo{
					conn:         &conn,
					chatPayload:  payload,
					disconnected: false,
				}
			} else {
				fmt.Printf("Client %s sent a chat message without a message\n", conn.RemoteAddr().String())
			}
		} else {
			fmt.Printf("Client %s sent an unknown message type: %s\n", conn.RemoteAddr().String(), payload.MsgType)
		}
	}
}

func readNextMessage(conn net.Conn, chatBuf *chatBuffer) (*chatmodels.Payload, error) {
	buf := make([]byte, 1024)
	sb := &strings.Builder{}
	var err error
	mLen := chatBuf.bufLen
	lenBytes := chatBuf.buf
	bytesToRead := 0

	for {
		if mLen == 0 {
			mLen, err = conn.Read(buf)
			if err != nil {
				return nil, err
			}
		}

		if bytesToRead > 0 {
			if mLen >= bytesToRead {
				sb.Write(buf[:bytesToRead])
				var payload chatmodels.Payload
				json.Unmarshal([]byte(sb.String()), &payload)

				chatBuf.bufLen = mLen - bytesToRead
				chatBuf.buf = buf[bytesToRead:mLen]
				return &payload, nil
			} else {
				sb.Write(buf[:mLen])
				bytesToRead -= mLen
				mLen = 0
			}
		} else if mLen+len(lenBytes) >= PAYLOAD_LEN_SIZE {
			bytesToRead = int(binary.BigEndian.Uint16(append(lenBytes, buf[:PAYLOAD_LEN_SIZE-len(lenBytes)]...)))

			buf = buf[PAYLOAD_LEN_SIZE-len(lenBytes):]
			mLen -= PAYLOAD_LEN_SIZE - len(lenBytes)
			lenBytes = []byte{}
			continue
		} else {
			lenBytes = buf[:mLen]
			mLen = 0
		}
	}

}
