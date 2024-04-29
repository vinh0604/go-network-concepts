package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	for i := 0; i < 10; i++ {
		result, err := tcpChecksum(i)
		if err != nil {
			fmt.Println("TCP Checksum 1: false, err:", err)
			return
		}

		fmt.Printf("TCP Checksum %d: %t\n", i, result)
	}
}

func tcpChecksum(packetIdx int) (bool, error) {
	addrsFile, err := os.ReadFile(fmt.Sprintf("./data/tcp_data/tcp_addrs_%d.txt", packetIdx))
	if err != nil {
		return false, err
	}
	addrsLine := strings.Split(string(addrsFile), "\n")[0]
	addrs := strings.Split(addrsLine, " ")
	sourceBytes, err := getIpBytes(addrs[0])
	if err != nil {
		return false, err
	}
	dstBytes, err := getIpBytes(addrs[1])
	if err != nil {
		return false, err
	}

	tcpDataFile, err := os.ReadFile(fmt.Sprintf("./data/tcp_data/tcp_data_%d.dat", packetIdx))
	if err != nil {
		return false, err
	}
	checksum := binary.BigEndian.Uint16(tcpDataFile[16:18])
	fmt.Println("Checksum:", fmt.Sprintf("0x%04x", checksum))

	tcpLen := len(tcpDataFile)
	tcpPseudoHeader := append(*sourceBytes, *dstBytes...)
	tcpPseudoHeader = append(tcpPseudoHeader, byte(0))
	tcpPseudoHeader = append(tcpPseudoHeader, byte(6))
	tcpLenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(tcpLenBytes, uint16(tcpLen))
	tcpPseudoHeader = append(tcpPseudoHeader, tcpLenBytes...)

	tcpWithZeroChecksum := append(tcpDataFile[:16], byte(0), byte(0))
	tcpWithZeroChecksum = append(tcpWithZeroChecksum, tcpDataFile[18:]...)
	if tcpLen%2 == 1 {
		tcpWithZeroChecksum = append(tcpWithZeroChecksum, byte(0))
	}

	dataToChecksum := append(tcpPseudoHeader, tcpWithZeroChecksum...)
	var total uint32 = 0
	for i := 0; i < len(dataToChecksum); i += 2 {
		// same as uint32(dataToChecksum[i])<<8 + uint32(dataToChecksum[i+1])
		total += uint32(binary.BigEndian.Uint16(dataToChecksum[i : i+2]))
	}
	// Handle potential carry-over from higher-order bits by folding it into lower bits
	// (total >> 16) to move any potential carry-over from the higher-order bits into the lower 16 bits.
	// (total & 0xffff) to isolate the lower-order bits.
	total = (total >> 16) + (total & 0xffff)
	// Repeat to ensure all carry-over is accounted for
	total += (total >> 16)

	calculatedChecksum := uint16(^total)
	fmt.Println("Calculated Checksum:", fmt.Sprintf("0x%04x", calculatedChecksum))

	return calculatedChecksum == checksum, nil
}

func getIpBytes(ip string) (*[]byte, error) {
	ipBytes := strings.Split(ip, ".")
	ipByteArr := make([]byte, 4)
	for i, ipByte := range ipBytes {
		ipValue, err := strconv.Atoi(ipByte)
		if err != nil {
			return nil, err
		}

		ipByteArr[i] = byte(ipValue)
	}

	return &ipByteArr, nil
}
