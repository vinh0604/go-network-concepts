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

	cm := chatutils.NewConnectionManager()
	go cm.Run()

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
					disconnectedNick := cm.Remove(conn)
					if disconnectedNick != nil {
						fmt.Printf("Client %s (nick=%s) left.\n", (*client.conn).RemoteAddr().String(), *disconnectedNick)
					}
					continue
				}

				if client.chatPayload != nil {
					if client.chatPayload.MsgType == chatmodels.MsgTypeHello {
						cm.Add(conn, *client.chatPayload.Nick)
						fmt.Printf("Client %s (nick=%s) joined.\n", (*client.conn).RemoteAddr().String(), *client.chatPayload.Nick)
						announce := chatmodels.Payload{
							MsgType: chatmodels.MsgTypeJoin,
							Nick:    client.chatPayload.Nick,
						}
						conns := cm.List()
						go relay(*client.chatPayload.Nick, *client.conn, conns, announce)
					} else if client.chatPayload.MsgType == chatmodels.MsgTypeChat {
						nick := cm.GetNick(*client.conn)
						if nick == nil {
							fmt.Printf("Client %s not registered\n", (*client.conn).RemoteAddr().String())
							continue
						}

						fmt.Printf("Client %s (nick=%s) sent a message.\n", (*client.conn).RemoteAddr().String(), *nick)
						chat := chatmodels.Payload{
							MsgType: chatmodels.MsgTypeChat,
							Nick:    nick,
							Msg:     client.chatPayload.Msg,
						}
						conns := cm.List()
						go relay(*nick, *client.conn, conns, chat)
					} else {
						fmt.Printf("Client %s sent an unknown message type: %s\n", (*client.conn).RemoteAddr().String(), client.chatPayload.MsgType)
					}
				}
			}
		}()
	}
}

func relay(nick string, clientConn net.Conn, clients []chatutils.ConnectionInfo, payload chatmodels.Payload) {
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
	for _, connInfo := range clients {
		if connInfo.Conn != clientConn {
			fmt.Printf("Relaying to %s\n", connInfo.Conn.RemoteAddr().String())
			connInfo.Conn.Write(outBytes)
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
