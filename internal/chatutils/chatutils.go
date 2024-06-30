package chatutils

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"strings"

	"github.com/vinh0604/go-network-concepts/internal/chatmodels"
)

const payloadLenBytesSize = 2

type ReadBuffer struct {
	bufferBytes []byte
}

func ReadNextMessage(conn net.Conn, readBuf *ReadBuffer) (*chatmodels.Payload, error) {
	messageBuffer := &strings.Builder{}
	payloadLength, err := readPayloadLength(conn, readBuf)
	if err != nil {
		return nil, err
	}

	for messageBuffer.Len() < payloadLength {
		if len(readBuf.bufferBytes) == 0 {
			if err := readIntoBuffer(conn, readBuf); err != nil {
				return nil, err
			}
		}

		bytesToRead := min(payloadLength-messageBuffer.Len(), len(readBuf.bufferBytes))
		messageBuffer.Write(readBuf.bufferBytes[:bytesToRead])
		readBuf.bufferBytes = readBuf.bufferBytes[bytesToRead:]
	}

	var payload chatmodels.Payload
	if err := json.Unmarshal([]byte(messageBuffer.String()), &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

func readPayloadLength(conn net.Conn, readBuf *ReadBuffer) (int, error) {
	for len(readBuf.bufferBytes) < payloadLenBytesSize {
		if err := readIntoBuffer(conn, readBuf); err != nil {
			return 0, err
		}
	}

	payloadLength := int(binary.BigEndian.Uint16(readBuf.bufferBytes[:payloadLenBytesSize]))
	readBuf.bufferBytes = readBuf.bufferBytes[payloadLenBytesSize:]
	return payloadLength, nil
}

func readIntoBuffer(conn net.Conn, readBuf *ReadBuffer) error {
	tempBuf := make([]byte, 1024)
	n, err := conn.Read(tempBuf)
	if err != nil {
		return err
	}
	readBuf.bufferBytes = append(readBuf.bufferBytes, tempBuf[:n]...)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type ConnectionInfo struct {
	conn net.Conn
	nick string
}

type ConnectionManager struct {
	addCh    chan ConnectionInfo
	removeCh chan net.Conn
	listCh   chan chan []ConnectionInfo
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		addCh:    make(chan ConnectionInfo),
		removeCh: make(chan net.Conn),
		listCh:   make(chan chan []ConnectionInfo),
	}
}

func (cm *ConnectionManager) Run() {
	conns := make(map[net.Conn]string)
	for {
		select {
		case connInfo := <-cm.addCh:
			conns[connInfo.conn] = connInfo.nick
		case conn := <-cm.removeCh:
			delete(conns, conn)
		case respCh := <-cm.listCh:
			listResult := make([]ConnectionInfo, 0, len(conns))
			for conn, nick := range conns {
				listResult = append(listResult, ConnectionInfo{conn, nick})
			}
			respCh <- listResult
		}
	}
}

func (cm *ConnectionManager) Add(conn net.Conn, nick string) {
	cm.addCh <- ConnectionInfo{conn, nick}
}

func (cm *ConnectionManager) Remove(conn net.Conn) {
	cm.removeCh <- conn
}

func (cm *ConnectionManager) List() []ConnectionInfo {
	respCh := make(chan []ConnectionInfo)
	cm.listCh <- respCh
	return <-respCh
}
