package pcapparser

import (
	"fmt"
	"math"
	"pcap-decoder/dictionary"
)

// LidarPoint contains the point information in spherical system.
type LidarPoint struct {
	rowIndex    uint8
	productID   byte
	distance    uint16
	azimuth     uint16
	nextAzimuth uint16
	Intensity   byte
}

// GetXYZ returns the XYZ Coordinates
func (p LidarPoint) GetXYZ() (float64, float64, float64) {
	azimuth := rad(p.Azimuth())
	elevAngle := rad(p.Bearing())

	cosEl := math.Cos(elevAngle)
	sinEl := math.Sin(elevAngle)
	sinAzimuth := math.Sin(azimuth)
	cosAzimuth := math.Cos(azimuth)

	distance := float64(p.Distance())
	X := distance * cosEl * sinAzimuth
	Y := distance * cosEl * cosAzimuth
	Z := distance * sinEl

	return X, Y, Z
}

// Distance returns the distance in mm
func (p LidarPoint) Distance() float64 {
	return 2 * float64(p.distance)
}

// Bearing returns the elevation angle in radians
func (p LidarPoint) Bearing() float64 {
	var elevAngle float64

	switch p.productID {
	case 0x22:
		elevAngle = float64(dictionary.VLP16ElevationAngles[p.rowIndex%16]) / 1000
	case 0x28:
		elevAngle = float64(dictionary.VLP32ElevationAngles[p.rowIndex]) / 1000
	}

	return elevAngle
}

// Azimuth returns the azimuth angle in radians
func (p LidarPoint) Azimuth() float64 {
	var azimuthOffset float64

	if p.productID == 0x28 {
		azimuthOffset = float64(dictionary.VLP32AzimuthOffset[p.rowIndex]) / 1000
	}

	azimuthGap := getAzimuthGap(p.azimuth, p.nextAzimuth)
	angleTimeOffset := getAngleTimeOffset(p.productID, p.rowIndex, azimuthGap)
	precisionAzimuth := (float64(p.azimuth)+angleTimeOffset)/100 + azimuthOffset

	if precisionAzimuth > 360 {
		precisionAzimuth -= 360
	}

	return precisionAzimuth
}

func getAngleTimeOffset(productID byte, rowIndex uint8, azimuthGap uint16) float64 {
	var K float64
	var angleTimeOffset float64
	switch productID {
	case 0x28:
		if rowIndex%2 == 0 {
			K = float64(rowIndex)
		} else {
			K = float64(rowIndex - 1)
		}
		// K /= 2
		// angleTimeOffset = float64(azimuthGap) * K  / 24 // 2.304/55.296 = 1/24

		angleTimeOffset = float64(azimuthGap) * K / 48 // K/2/24 = K /48

	case 0x22:
		var time float64
		var totalTime float64
		if rowIndex < 16 {
			time = 2304 * float64(rowIndex)
			totalTime = 55296
			// time = float64(rowIndex)
			// totalTime = 24
		} else {
			time = 55296 + 2304*float64(rowIndex-16)
			totalTime = 110592
		}
		angleTimeOffset = float64(azimuthGap) * time / totalTime

	default:
		panic(fmt.Sprintf("product ID 0x%x is not supported", productID))
	}

	return angleTimeOffset
}
