package netfunc

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

func IpStringToBytes(ipv4 string) ([]byte, error) {
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

func IpBytesToInt32(ipBytes []byte) (uint32, error) {
	if len(ipBytes) != 4 {
		return 0, errors.New("invalid IP address")
	}
	return uint32(ipBytes[0])<<24 | uint32(ipBytes[1])<<16 | uint32(ipBytes[2])<<8 | uint32(ipBytes[3]), nil
}

func IpInt32ToBytes(ipInt32 uint32) []byte {
	return []byte{
		byte(ipInt32 >> 24),
		byte(ipInt32 >> 16),
		byte(ipInt32 >> 8),
		byte(ipInt32),
	}
}

func IpBytesToString(ipBytes []byte) string {
	return strconv.Itoa(int(ipBytes[0])) + "." + strconv.Itoa(int(ipBytes[1])) + "." + strconv.Itoa(int(ipBytes[2])) + "." + strconv.Itoa(int(ipBytes[3]))
}

func ComputeSubnetMask(notation uint8) ([]byte, error) {
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

func GetNetworkNumber(ipBytes []byte, notation uint8) ([]byte, error) {
	if len(ipBytes) != 4 {
		return nil, errors.New("invalid IP address")
	}
	subnetMask, err := ComputeSubnetMask(notation)
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

func GetHostBits(ipBytes []byte, notation uint8) ([]byte, error) {
	if len(ipBytes) != 4 {
		return nil, errors.New("invalid IP address")
	}
	subnetMask, err := ComputeSubnetMask(notation)
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

func IpsSameSubnet(ip1 string, ip2 string, subnetNotation uint8) (bool, error) {
	ip1Bytes, err := IpStringToBytes(ip1)
	if err != nil {
		return false, err
	}

	ip2Bytes, err := IpStringToBytes(ip2)
	if err != nil {
		return false, err
	}

	network1, err := GetNetworkNumber(ip1Bytes, subnetNotation)
	if err != nil {
		return false, err
	}

	network2, err := GetNetworkNumber(ip2Bytes, subnetNotation)
	if err != nil {
		return false, err
	}

	return slices.Equal(network1, network2), nil
}

type RouterInfo struct {
	NetmaskNotation uint8
}

func RouterForIp(routers map[string]RouterInfo, ip string) (string, error) {
	_, err := IpStringToBytes(ip)
	if err != nil {
		return "", err
	}

	for routerIp, info := range routers {
		if result, _ := IpsSameSubnet(ip, routerIp, info.NetmaskNotation); result {
			return routerIp, nil
		}
	}

	return "", errors.New("no router found for IP")
}
