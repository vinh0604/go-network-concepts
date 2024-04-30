package main

import (
	"errors"
	"strconv"
	"strings"
)

func ipStringToBytes(ipv4 string) ([]byte, error) {
	parts := strings.Split(ipv4, ".")

	if len(parts) != 4 {
		return nil, errors.New("invalid IPv4 address")
	}

	ipBytes := make([]byte, 4)
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		ipBytes[i] = byte(num)
	}

	return ipBytes, nil
}

func ipBytesToInt32(ipBytes []byte) (uint32, error) {
	if len(ipBytes) != 4 {
		return 0, errors.New("invalid IP address")
	}
	return uint32(ipBytes[0])<<24 | uint32(ipBytes[1])<<16 | uint32(ipBytes[2])<<8 | uint32(ipBytes[3]), nil
}

func ipInt32ToBytes(ipInt32 uint32) []byte {
	return []byte{
		byte(ipInt32 >> 24),
		byte(ipInt32 >> 16),
		byte(ipInt32 >> 8),
		byte(ipInt32),
	}
}

func ipBytesToString(ipBytes []byte) string {
	return strconv.Itoa(int(ipBytes[0])) + "." + strconv.Itoa(int(ipBytes[1])) + "." + strconv.Itoa(int(ipBytes[2])) + "." + strconv.Itoa(int(ipBytes[3]))
}

func computeSubnetMask(notation uint8) ([]byte, error) {
	if notation > 32 {
		return nil, errors.New("invalid subnet notation")
	}

	runOfOne := (1 << notation) - 1
	runOfOne <<= 32 - notation
	return []byte{
		byte(runOfOne >> 24),
		byte(runOfOne >> 16),
		byte(runOfOne >> 8),
		byte(runOfOne),
	}, nil
}

func getSubnet(ipBytes []byte, notation uint8) ([]byte, error) {
	if len(ipBytes) != 4 {
		return nil, errors.New("invalid IP address")
	}
	subnetMask, err := computeSubnetMask(notation)
	if err != nil {
		return nil, err
	}

	return []byte{
		ipBytes[0] & subnetMask[0],
		ipBytes[1] & subnetMask[1],
		ipBytes[2] & subnetMask[2],
		ipBytes[3] & subnetMask[3],
	}, nil
}

func getHostBits(ipBytes []byte, notation uint8) ([]byte, error) {
	if len(ipBytes) != 4 {
		return nil, errors.New("invalid IP address")
	}
	subnetMask, err := computeSubnetMask(notation)
	if err != nil {
		return nil, err
	}

	return []byte{
		ipBytes[0] & ^subnetMask[0],
		ipBytes[1] & ^subnetMask[1],
		ipBytes[2] & ^subnetMask[2],
		ipBytes[3] & ^subnetMask[3],
	}, nil
}
