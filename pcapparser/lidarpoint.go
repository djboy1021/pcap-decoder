package pcapparser

import (
	"math"
	"pcap-decoder/dictionary"
)

// LidarPoint contains the point information in spherical system.
type LidarPoint struct {
	LaserID     uint8
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
		elevAngle = float64(dictionary.VLP16ElevationAngles[p.LaserID%16]) / 1000
	case 0x28:
		elevAngle = float64(dictionary.VLP32ElevationAngles[p.LaserID]) / 1000
	}

	return elevAngle
}

// Azimuth returns the azimuth angle in radians
func (p LidarPoint) Azimuth() float64 {
	var azimuthOffset float64

	if p.productID == 0x28 {
		azimuthOffset = float64(dictionary.VLP32AzimuthOffset[p.LaserID]) / 1000
	}

	azimuthGap := getAzimuthGap(p.azimuth, p.nextAzimuth)
	angleTimeOffset := getAngleTimeOffset(p.productID, p.LaserID, azimuthGap)
	precisionAzimuth := p.azimuth + angleTimeOffset

	return float64(precisionAzimuth)/100 + azimuthOffset
}