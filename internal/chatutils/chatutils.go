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
	buf := make([]byte, 1024)
	sb := &strings.Builder{}
	var err error
	mLen := len(readBuf.bufferBytes)
	lenBytes := readBuf.bufferBytes
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

				readBuf.bufferBytes = buf[bytesToRead:mLen]
				return &payload, nil
			} else {
				sb.Write(buf[:mLen])
				bytesToRead -= mLen
				mLen = 0
			}
		} else if mLen+len(lenBytes) >= payloadLenBytesSize {
			bytesToRead = int(binary.BigEndian.Uint16(append(lenBytes, buf[:payloadLenBytesSize-len(lenBytes)]...)))

			buf = buf[payloadLenBytesSize-len(lenBytes):]
			mLen -= payloadLenBytesSize - len(lenBytes)
			lenBytes = []byte{}
			continue
		} else {
			lenBytes = buf[:mLen]
			mLen = 0
		}
	}
}
