package chatutils

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vinh0604/go-network-concepts/internal/chatmodels"
)

func TestReadNextMessage(t *testing.T) {
	assert := assert.New(t)

	// Create a pipe for testing
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Test payload
	testPayload := chatmodels.Payload{
		MsgType: chatmodels.MsgTypeChat,
		Nick:    stringPtr("TestUser"),
		Msg:     stringPtr("Hello, World!"),
	}

	// Write the message to the pipe in a goroutine
	go func() {
		// Marshal the payload
		payloadBytes, err := json.Marshal(testPayload)
		assert.NoError(err, "Failed to marshal test payload")

		// Prepare the message with length prefix
		messageLen := uint16(len(payloadBytes))
		message := make([]byte, 2+len(payloadBytes))
		binary.BigEndian.PutUint16(message[:2], messageLen)
		copy(message[2:], payloadBytes)

		_, err = client.Write(message)
		assert.NoError(err, "Failed to write to pipe")
	}()

	// Read the message using ReadNextMessage
	readBuf := &ReadBuffer{}
	receivedPayload, err := ReadNextMessage(server, readBuf)
	assert.NoError(err, "ReadNextMessage failed")

	// Compare the received payload with the original
	assert.Equal(testPayload.MsgType, receivedPayload.MsgType, "MsgType mismatch")
	assert.Equal(*testPayload.Nick, *receivedPayload.Nick, "Nick mismatch")
	assert.Equal(*testPayload.Msg, *receivedPayload.Msg, "Msg mismatch")
}

func TestReadNextMessageInvalidPayload(t *testing.T) {
	assert := assert.New(t)

	// Create a pipe for testing
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Invalid JSON payload
	invalidPayload := []byte(`{"MsgType": "chat", "Nick": "TestUser", "Msg": "Hello, World!"`)

	// Write the message to the pipe in a goroutine
	go func() {
		messageLen := uint16(len(invalidPayload))
		message := make([]byte, 2+len(invalidPayload))
		binary.BigEndian.PutUint16(message[:2], messageLen)
		copy(message[2:], invalidPayload)

		_, err := client.Write(message)
		assert.NoError(err, "Failed to write to pipe")
	}()

	// Read the message using ReadNextMessage
	readBuf := &ReadBuffer{}
	receivedPayload, err := ReadNextMessage(server, readBuf)

	// Check that an error was returned
	assert.Error(err, "Expected an error for invalid JSON payload")
	assert.Nil(receivedPayload, "Expected nil payload for invalid JSON")
}

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}
