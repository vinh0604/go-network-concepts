package chatutils

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"testing"

	"github.com/vinh0604/go-network-concepts/internal/chatmodels"
)

func TestReadNextMessage(t *testing.T) {
	// Create a pipe for testing
	client, server := net.Pipe()

	// Test payload
	testPayload := chatmodels.Payload{
		MsgType: chatmodels.MsgTypeChat,
		Nick:    stringPtr("TestUser"),
		Msg:     stringPtr("Hello, World!"),
	}

	// Marshal the payload
	payloadBytes, err := json.Marshal(testPayload)
	if err != nil {
		t.Fatalf("Failed to marshal test payload: %v", err)
	}

	// Prepare the message with length prefix
	messageLen := uint16(len(payloadBytes))
	message := make([]byte, 2+len(payloadBytes))
	binary.BigEndian.PutUint16(message[:2], messageLen)
	copy(message[2:], payloadBytes)

	// Write the message to the pipe in a goroutine
	go func() {
		_, err := server.Write(message)
		if err != nil {
			t.Errorf("Failed to write to pipe: %v", err)
		}
		server.Close()
	}()

	// Read the message using ReadNextMessage
	readBuf := &ReadBuffer{}
	receivedPayload, err := ReadNextMessage(client, readBuf)
	if err != nil {
		t.Fatalf("ReadNextMessage failed: %v", err)
	}

	// Compare the received payload with the original
	if receivedPayload.MsgType != testPayload.MsgType {
		t.Errorf("Expected MsgType %s, got %s", testPayload.MsgType, receivedPayload.MsgType)
	}
	if *receivedPayload.Nick != *testPayload.Nick {
		t.Errorf("Expected Nick %s, got %s", *testPayload.Nick, *receivedPayload.Nick)
	}
	if *receivedPayload.Msg != *testPayload.Msg {
		t.Errorf("Expected Msg %s, got %s", *testPayload.Msg, *receivedPayload.Msg)
	}

	client.Close()
}

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}
