package parsepcap

import (
	"encoding/binary"
	"math"
	"pcap-decoder/dictionary"
)

func getTime(packetData *[]byte) uint32 {
	return binary.LittleEndian.Uint32((*packetData)[1242:1246])
}

func getProductID(packetData *[]byte) byte {
	return (*packetData)[1247]
}

func isDualMode(packetData *[]byte) bool {
	return (*packetData)[1247] == 0x39
}

func getDistanceAndReflectivity(packetData *[]byte, colIndex uint8, rowIndex uint8) (uint32, byte) {
	blkIndex := 46 + int(colIndex)*100 + int(rowIndex)*3

	distance := uint32((*packetData)[blkIndex+1])<<8 + uint32((*packetData)[blkIndex])
	distance = distance << 1
	reflectivity := (*packetData)[blkIndex+2]

	return distance, reflectivity
}

func getCurrentAndNextRawAzimuths(currPacketData *[]byte, nextPacketData *[]byte, colIndex uint8) (uint16, uint16) {
	// Current Azimuth
	blkIndex := int(colIndex)*100 + 44
	currAzimuth := binary.LittleEndian.Uint16((*currPacketData)[blkIndex : blkIndex+2])

	// Next Azimuth
	var nextAzimuth uint16
	if colIndex >= 11 {
		nextAzimuth = binary.LittleEndian.Uint16((*nextPacketData)[44:46])
	} else {
		blkIndex = int(colIndex+1)*100 + 44
		nextAzimuth = binary.LittleEndian.Uint16((*currPacketData)[blkIndex : blkIndex+2])
	}

	// Adjust if exceeds 360degrees
	if currAzimuth > 36000 {
		currAzimuth -= 36000
	}
	if nextAzimuth > 36000 {
		nextAzimuth -= 36000
	}

	return currAzimuth, nextAzimuth
}

func getPrecisionAzimuth(currAzimuth uint16, nextAzimuth uint16, rowIndex uint8, productID byte) uint16 {
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

	if precisionAzimuth >= 36000 {
		precisionAzimuth -= 36000
	}

	return precisionAzimuth
}

func rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

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

func getLaserID(productID byte, rowIndex uint8) uint8 {
	switch productID {
	case 0x22:
		return rowIndex % 16
	}

	return rowIndex
}

func getTimeStamp(data *[]byte, rowIndex uint8, colIndex uint8) uint32 {
	productID := getProductID(data)
	firingTime := getTime(data)

	if productID == 0x28 {
		return firingTime + dictionary.SingleModeVLP32TimingOffsetTable[rowIndex][colIndex]
	}

	switch productID {
	case 0x28:
		return firingTime + dictionary.SingleModeVLP32TimingOffsetTable[rowIndex][colIndex]
	case 0x22:
		return firingTime + dictionary.SingleModeVLP16TimingOffsetTable[rowIndex][colIndex]
	}

	return firingTime
}
