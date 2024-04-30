package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpStringToBytes_withValidIp(t *testing.T) {
	result, err := ipStringToBytes("192.167.23.5")
	assert.NoError(t, err, "Expected no error when converting a valid IP address")
	assert.Equal(t, []byte{192, 167, 23, 5}, result, "Expected the correct byte representation of the IP address")
}

func TestIpStringToBytes_withInvalidIp(t *testing.T) {
	_, err := ipStringToBytes("192.167.23")
	assert.Error(t, err, "Expected an error when converting an invalid IP address")

	_, err = ipStringToBytes("a.b.c.d")
	assert.Error(t, err, "Expected an error when converting an invalid IP address")
}

func TestIpBytesToInt32_withValidIpBytes(t *testing.T) {
	result, err := ipBytesToInt32([]byte{192, 167, 23, 5})
	assert.NoError(t, err, "Expected no error when converting a valid IP address")
	assert.Equal(t, uint32(3232175877), result, "Expected the correct integer representation of the IP address")
}

func TestIpBytesToInt32_withInvalidIpBytes(t *testing.T) {
	_, err := ipBytesToInt32([]byte{192, 167, 23})
	assert.Error(t, err, "Expected an error when converting an invalid IP address")
}

func TestIpInt32ToBytes_withInvalidIpBytes(t *testing.T) {
	result := ipInt32ToBytes(uint32(3232175877))
	assert.Equal(t, []byte{192, 167, 23, 5}, result, "Expected the correct byte representation of the IP address")
}

func TestIpBytesToString_withInvalidIpBytes(t *testing.T) {
	result := ipBytesToString([]byte{192, 167, 23, 5})
	assert.Equal(t, "192.167.23.5", result, "Expected the correct byte representation of the IP address")
}
