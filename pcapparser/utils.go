package pcapparser

import (
	"encoding/binary"
	"math"
	"pcap-decoder/dictionary"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func isDualMode(packetData *[]byte) bool {
	return (*packetData)[1246] == 0x57
}

func getProductID(packetData *[]byte) byte {
	return (*packetData)[1247]
}

func getTime(packetData *[]byte) uint32 {
	return binary.LittleEndian.Uint32((*packetData)[1242:1246])
}

func getAzimuthGap(currAzimuth uint16, nextAzimuth uint16) uint16 {
	if nextAzimuth < currAzimuth {
		return (36000 - currAzimuth) + nextAzimuth // same as nextAzimuth + 36000 - currAzimuth
	}
	return nextAzimuth - currAzimuth
}

// returns the elevation angle in degrees
func getElevationAngle(productID byte, rowIndex uint8) float64 {
	var elevAngle float64

	switch productID {
	case 0x22:
		elevAngle = float64(dictionary.VLP16ElevationAngles[rowIndex%16]) / 1000
	case 0x28:
		elevAngle = float64(dictionary.VLP32ElevationAngles[rowIndex]) / 1000
	}

	return elevAngle
}

// returns the elevation angle in degrees
func getRawElevationAngle(productID byte, rowIndex uint8) int16 {
	var elevAngle int16

	switch productID {
	case 0x22:
		elevAngle = dictionary.VLP16ElevationAngles[rowIndex%16]
	case 0x28:
		elevAngle = dictionary.VLP32ElevationAngles[rowIndex]
	}

	return elevAngle
}

func rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}
