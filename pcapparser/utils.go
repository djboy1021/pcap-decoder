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

func getPrecisionAzimuth(currAzimuth uint16, nextAzimuth uint16, rowIndex uint8, productID byte) float32 {
	var azimuthGap uint16

	if nextAzimuth < currAzimuth {
		azimuthGap = (36000 - currAzimuth) + nextAzimuth // same as nextAzimuth + 36000 - currAzimuth
	} else {
		azimuthGap = nextAzimuth - currAzimuth
	}

	var precisionAzimuth uint16
	if productID == 0x28 {
		K := float64((1 + rowIndex) / 2)
		precisionAzimuth = currAzimuth + uint16(math.Round(float64(azimuthGap)*K*0.04166667)) // 2.304/55.296 = 0.0416667

	} else if productID == 0x22 {
		K := float64(rowIndex)
		if rowIndex < 16 {
			// Precision_Azimuth[K] := Azimuth[datablock_n] + (AzimuthGap * 2.304 μs * K) / 55.296 μs);
			precisionAzimuth = currAzimuth + uint16(math.Round(float64(azimuthGap)*K*0.04166667)) // 2.304/55.296 = 0.0416667
		} else {
			// Precision_Azimuth[K] := Azimuth[datablock_n] + (AzimuthGap * 2.304 μs * ((K-16) + 55.296 μs)) / (2 * 55.296 μs);
			precisionAzimuth = currAzimuth + uint16((float64(azimuthGap)*(float64(K-16)+55.296))*0.02083333) // 2.304/55.296 = 0.0416667
		}
	} else {
		panic(string(productID) + "is not supported")
	}

	var paFloat float32
	if productID == 0x28 {
		paFloat = float32(precisionAzimuth)/100 + float32(dictionary.VLP32AzimuthOffset[rowIndex])/1000
	}

	if paFloat > 360 {
		paFloat -= 360
	}

	return paFloat
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
