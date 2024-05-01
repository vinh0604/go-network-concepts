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

func TestComputeSubnetMask_withValidNotation(t *testing.T) {
	result, err := computeSubnetMask(24)
	assert.NoError(t, err, "Expected no error when computing a valid subnet mask")
	assert.Equal(t, []byte{255, 255, 255, 0}, result, "Expected the correct subnet mask")

	result, _ = computeSubnetMask(16)
	assert.Equal(t, []byte{255, 255, 0, 0}, result, "Expected the correct subnet mask")

	result, _ = computeSubnetMask(8)
	assert.Equal(t, []byte{255, 0, 0, 0}, result, "Expected the correct subnet mask")

	result, _ = computeSubnetMask(25)
	assert.Equal(t, []byte{255, 255, 255, 128}, result, "Expected the correct subnet mask")
}

func TestComputeSubnetMask_withInvalidNotation(t *testing.T) {
	_, err := computeSubnetMask(35)
	assert.Error(t, err, "Expected an error when computing an invalid subnet mask")
}

func TestGetNetworkNumber(t *testing.T) {
	result, err := getNetworkNumber([]byte{198, 51, 100, 10}, 24)
	assert.NoError(t, err, "Expected no error when getting a subnet")
	assert.Equal(t, []byte{198, 51, 100, 0}, result, "Expected the correct subnet")

	result, _ = getNetworkNumber([]byte{198, 51, 100, 140}, 25)
	assert.Equal(t, []byte{198, 51, 100, 128}, result, "Expected the correct subnet")
}

func TestGetHostBits(t *testing.T) {
	result, err := getHostBits([]byte{198, 51, 100, 10}, 24)
	assert.NoError(t, err, "Expected no error when getting a subnet")
	assert.Equal(t, []byte{0, 0, 0, 10}, result, "Expected the correct subnet")

	result, _ = getHostBits([]byte{198, 51, 100, 140}, 25)
	assert.Equal(t, []byte{0, 0, 0, 12}, result, "Expected the correct subnet")
}

func TestIpsSameSubnet(t *testing.T) {
	result, err := ipsSameSubnet("192.168.1.1", "192.168.1.3", 24)
	assert.NoError(t, err, "Expected no error when checking if IPs are in the same subnet")
	assert.True(t, result, "Expected the IPs to be in the same subnet")

	result, _ = ipsSameSubnet("192.168.2.1", "192.168.1.3", 24)
	assert.False(t, result, "Expected the IPs to be in the same subnet")
}

func TestRouterForIp(t *testing.T) {
	routers := map[string]RouterInfo{
		"10.34.166.1": RouterInfo{24},
		"10.34.194.1": RouterInfo{24},
		"10.34.98.1":  RouterInfo{24},
	}

	result, err := routerForIp(routers, "10.34.166.170")
	assert.NoError(t, err, "Expected no error when getting the router for an IP")
	assert.Equal(t, "10.34.166.1", result, "Expected the correct router")

	_, err = routerForIp(routers, "10.35.166.170")
	assert.Error(t, err, "Expected an error when getting the router for an IP")
}
