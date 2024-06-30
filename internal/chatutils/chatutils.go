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
