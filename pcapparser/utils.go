package pcapparser

import (
	"encoding/binary"
	"fmt"
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

func getAngleTimeOffset(productID byte, rowIndex uint8, azimuthGap uint16) uint16 {
	var K float64
	var angleTimeOffset float64
	switch productID {
	case 0x28:
		if rowIndex%2 == 0 {
			K = float64(rowIndex)
		} else {
			K = float64(rowIndex - 1)
		}
		K /= 2
		angleTimeOffset = float64(azimuthGap) * K * 2304 / 55296 // 2.304/55.296 = 0.0416667

	case 0x22:
		var time float64
		var totalTime float64
		if rowIndex < 16 {
			time = 2304 * float64(rowIndex)
			totalTime = 55296
		} else {
			time = 55296 + 2304*float64(rowIndex-16)
			totalTime = 110592
		}
		angleTimeOffset = float64(azimuthGap) * time / totalTime

	default:
		panic(fmt.Sprintf("product ID 0x%x is not supported", productID))
	}

	return uint16(math.Round(angleTimeOffset))
}

func getAzimuthGap(currAzimuth uint16, nextAzimuth uint16) uint16 {
	if nextAzimuth < currAzimuth {
		return (36000 - currAzimuth) + nextAzimuth // same as nextAzimuth + 36000 - currAzimuth
	}
	return nextAzimuth - currAzimuth
}

func getPrecisionAzimuth(currAzimuth uint16, nextAzimuth uint16, rowIndex uint8, productID byte) float32 {
	var azimuthGap uint16

	if nextAzimuth < currAzimuth {
		azimuthGap = (36000 - currAzimuth) + nextAzimuth // same as nextAzimuth + 36000 - currAzimuth
	} else {
		azimuthGap = nextAzimuth - currAzimuth
	}

	var precisionAzimuth uint16

	switch productID {
	case 0x28:
		angleTimeOffset := getAngleTimeOffset(productID, rowIndex, azimuthGap)
		fmt.Println("calcOffset", angleTimeOffset, rowIndex, currAzimuth, nextAzimuth)
		precisionAzimuth = currAzimuth + angleTimeOffset
	case 0x22:
		K := float64(rowIndex)
		if rowIndex < 16 {
			// Precision_Azimuth[K] := Azimuth[datablock_n] + (AzimuthGap * 2.304 μs * K) / 55.296 μs);
			precisionAzimuth = currAzimuth + uint16(math.Round(float64(azimuthGap)*K*0.04166667)) // 2.304/55.296 = 0.0416667
		} else {
			// Precision_Azimuth[K] := Azimuth[datablock_n] + (AzimuthGap * 2.304 μs * ((K-16) + 55.296 μs)) / (2 * 55.296 μs);
			precisionAzimuth = currAzimuth + uint16((float64(azimuthGap)*(float64(K-16)+55.296))*0.02083333) // 2.304/(2*55.296) = 0.02083333
		}
	default:
		panic(string(productID) + "is not supported")
	}

	// var paFloat float32
	// if productID == 0x28 {
	// 	paFloat = float32(precisionAzimuth)/100 + float32(dictionary.VLP32AzimuthOffset[rowIndex])/1000
	// }

	paFloat := float32(precisionAzimuth)
	// fmt.Println("asdfsd", dictionary.VLP32AzimuthOffset[rowIndex], currAzimuth, nextAzimuth)

	// if paFloat > 360 {
	// 	paFloat -= 360
	// }

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

func getXYZCoordinates(distance *uint32, azimuth uint16, productID byte, rowIndex uint8) (int, int, int) {
	var azimuthOffset float64

	if productID == 0x28 {
		azimuthOffset = float64(dictionary.VLP32AzimuthOffset[rowIndex]) / 1000
	}

	elevAngle := getElevationAngle(productID, rowIndex)

	cosEl := math.Cos(rad((elevAngle)))
	sinEl := math.Sin(rad((elevAngle)))
	sinAzimuth := math.Sin(rad((azimuthOffset) + float64(azimuth)/100))
	cosAzimuth := math.Cos(rad((azimuthOffset) + float64(azimuth)/100))

	X := math.Round(float64(*distance) * cosEl * sinAzimuth)
	Y := math.Round(float64(*distance) * cosEl * cosAzimuth)
	Z := math.Round(float64(*distance) * sinEl)

	return int(X), int(Y), int(Z)
}
