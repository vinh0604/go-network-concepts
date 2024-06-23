package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/vinh0604/go-network-concepts/internal/chatmodels"
	"github.com/vinh0604/go-network-concepts/internal/chatutils"
)

type clientInfo struct {
	conn         *net.Conn
	chatPayload  *chatmodels.Payload
	disconnected bool
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

	clients := map[net.Conn]string{}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		clientCh := make(chan clientInfo)
		go handleConn(conn, clientCh)

		go func() {
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
			fmt.Printf("Relaying to %s\n", conn.RemoteAddr().String())
			conn.Write(outBytes)
		}
	}
}

func handleConn(conn net.Conn, clientCh chan clientInfo) {
	defer conn.Close()

	readBuf := chatutils.ReadBuffer{}
	for {
		payload, err := chatutils.ReadNextMessage(conn, &readBuf)

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
