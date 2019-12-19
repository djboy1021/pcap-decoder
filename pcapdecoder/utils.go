package pcapdecoder

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

const piOver180 = math.Pi / 180

func radians(degrees float64) float64 {
	return degrees * piOver180
}

func degrees(radians float64) float64 {
	deg := radians / piOver180
	if deg < 0 {
		deg += 360
	}

	return deg
}

func normalizeAngle(angle float64) float64 {
	if angle < 0 {
		angle += 360
	}
	return angle
}
