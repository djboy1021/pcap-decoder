package pcapparser

import (
	"encoding/binary"
	"fmt"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func isDualMode(packetData *[]byte) bool {
	fmt.Println((*packetData)[1247])
	return (*packetData)[1247] == 0x39
}

func getProductID(packetData *[]byte) byte {
	return (*packetData)[1247]
}

func getTime(packetData *[]byte) uint32 {
	return binary.LittleEndian.Uint32((*packetData)[1242:1246])
}
