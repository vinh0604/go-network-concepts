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
		Msg:     stringPtr("Hello, World!"),
	}
	// Marshal the payload
	payloadBytes, err := json.Marshal(testPayload)
	assert.NoError(err, "Failed to marshal test payload")

	// Write the message to the pipe in a goroutine
	go func() {
		err := writeMessage(client, payloadBytes)
		assert.NoError(err, "Failed to write to pipe")
	}()

	// Read the message using ReadNextMessage
	readBuf := &ReadBuffer{}
	receivedPayload, err := ReadNextMessage(server, readBuf)
	assert.NoError(err, "ReadNextMessage failed")

	// Compare the received payload with the original
	assert.Equal(testPayload.MsgType, receivedPayload.MsgType, "MsgType mismatch")
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
		err := writeMessage(client, invalidPayload)
		assert.NoError(err, "Failed to write to pipe")
	}()

	// Read the message using ReadNextMessage
	readBuf := &ReadBuffer{}
	receivedPayload, err := ReadNextMessage(server, readBuf)

	// Check that an error was returned
	assert.Error(err, "Expected an error for invalid JSON payload")
	assert.Nil(receivedPayload, "Expected nil payload for invalid JSON")
}

func TestReadNextMessageMultipleMessages(t *testing.T) {
	assert := assert.New(t)

	// Create a pipe for testing
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Test payload
	testPayload1 := chatmodels.Payload{
		MsgType: chatmodels.MsgTypeHello,
		Nick:    stringPtr("TestUser"),
	}
	testPayload2 := chatmodels.Payload{
		MsgType: chatmodels.MsgTypeChat,
		Msg:     stringPtr("Hello, World!"),
	}
	// Marshal the payload
	payload1Bytes, err := json.Marshal(testPayload1)
	assert.NoError(err, "Failed to marshal test payload 1")
	payload2Bytes, err := json.Marshal(testPayload2)
	assert.NoError(err, "Failed to marshal test payload 2")

	// Write the message to the pipe in a goroutine
	go func() {
		message := make([]byte, 2+len(payload1Bytes)+2+len(payload2Bytes))
		binary.BigEndian.PutUint16(message[:2], uint16(len(payload1Bytes)))
		copy(message[2:], payload1Bytes)
		binary.BigEndian.PutUint16(message[2+len(payload1Bytes):2+len(payload1Bytes)+2], uint16(len(payload2Bytes)))
		copy(message[2+len(payload1Bytes)+2:], payload2Bytes)
		_, err := client.Write(message)
		assert.NoError(err, "Failed to write to pipe")
	}()

	// Read the message using ReadNextMessage
	readBuf := &ReadBuffer{}
	receivedPayload1, err := ReadNextMessage(server, readBuf)
	assert.NoError(err, "ReadNextMessage 1 failed")
	assert.Equal(testPayload1.MsgType, receivedPayload1.MsgType, "MsgType mismatch")
	assert.Equal(*testPayload1.Nick, *receivedPayload1.Nick, "Nick mismatch")

	receivedPayload2, err := ReadNextMessage(server, readBuf)
	assert.NoError(err, "ReadNextMessage 2 failed")
	assert.Equal(testPayload2.MsgType, receivedPayload2.MsgType, "MsgType mismatch")
	assert.Equal(*testPayload2.Msg, *receivedPayload2.Msg, "Msg mismatch")
}

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func writeMessage(conn net.Conn, payloadBytes []byte) error {
	messageLen := uint16(len(payloadBytes))
	message := make([]byte, 2+len(payloadBytes))
	binary.BigEndian.PutUint16(message[:2], messageLen)
	copy(message[2:], payloadBytes)

	_, err := conn.Write(message)
	return err
}

func TestConnectionManager(t *testing.T) {
	assert := assert.New(t)

	cm := NewConnectionManager()
	go cm.Run()

	// Test Add method
	conn1 := &net.TCPConn{}
	cm.Add(conn1, "user1")
	conn2 := &net.TCPConn{}
	cm.Add(conn2, "user2")

	// Test List method
	connections := cm.List()
	assert.Len(connections, 2, "Expected 2 connections")
	assert.Contains(connections, ConnectionInfo{Conn: conn1, Nick: "user1"})
	assert.Contains(connections, ConnectionInfo{Conn: conn2, Nick: "user2"})

	// Test Remove method
	removedNick := cm.Remove(conn1)
	assert.Equal("user1", removedNick, "Expected removed nick to be 'user1'")
	connections = cm.List()
	assert.Len(connections, 1, "Expected 1 connection after removal")
	assert.Contains(connections, ConnectionInfo{Conn: conn2, Nick: "user2"})
	assert.NotContains(connections, ConnectionInfo{Conn: conn1, Nick: "user1"})

	// Test adding a connection with the same nick
	conn3 := &net.TCPConn{}
	cm.Add(conn3, "user2")
	connections = cm.List()
	assert.Len(connections, 2, "Expected 2 connections")
	assert.Contains(connections, ConnectionInfo{Conn: conn2, Nick: "user2"})
	assert.Contains(connections, ConnectionInfo{Conn: conn3, Nick: "user2"})

	// Test removing a non-existent connection
	nonExistentConn := &net.TCPConn{}
	removedNick = cm.Remove(nonExistentConn)
	assert.Equal("", removedNick, "Expected empty string for non-existent connection")
	connections = cm.List()
	assert.Len(connections, 2, "Expected no change in connections")
}
