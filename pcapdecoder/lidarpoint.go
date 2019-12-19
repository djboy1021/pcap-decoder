package pcapdecoder

import (
	"fmt"
	"math"
	"pcap-decoder/dictionary"
)

// LidarPoint contains the point information in spherical system.
type LidarPoint struct {
	distance    uint16
	azimuth     uint16
	nextAzimuth uint16
	rowIndex    uint8
	productID   byte
	Intensity   byte
}

// CartesianPoint contains X, Y, Z
type CartesianPoint struct {
	X         float64 `json:"x" bson:"x"`
	Y         float64 `json:"y" bson:"y"`
	Z         float64 `json:"z" bson:"z"`
	Intensity uint8   `json:"i" bson:"i"`
}

// SphericalPoint contains radius, azimuth angle, and elevation angle
type SphericalPoint struct {
	Radius  float64 `json:"r"`
	Azimuth float64 `json:"azimuth"`
	Bearing float64 `json:"bearing"`
}

// ToSpherical transforms the cartesian point into spherical coordinate, angles are in degrees
func (cp CartesianPoint) ToSpherical() SphericalPoint {
	radius := math.Sqrt(cp.X*cp.X + cp.Y*cp.Y + cp.Z*cp.Z)
	return SphericalPoint{
		Radius:  radius,
		Azimuth: normalizeAngle(degrees(math.Atan(cp.X / cp.Y))),
		Bearing: normalizeAngle(degrees(math.Asin(cp.Z / radius)))}
}

// Rotate returns the rotated cartesian point
func (cp CartesianPoint) Rotate(A *[3][3]float64) CartesianPoint {
	return CartesianPoint{
		X: A[0][0]*cp.X + A[0][1]*cp.Y + A[0][2]*cp.Z,
		Y: A[1][0]*cp.X + A[1][1]*cp.Y + A[1][2]*cp.Z,
		Z: A[2][0]*cp.X + A[2][1]*cp.Y + A[2][2]*cp.Z}
}

// Translate returns the translated CartesianPoint position
func (cp CartesianPoint) Translate(t Translation) CartesianPoint {
	return CartesianPoint{
		X: cp.X + float64(t.x),
		Y: cp.Y + float64(t.y),
		Z: cp.Z + float64(t.z)}
}

// GetXYZ returns the XYZ Coordinates
func (p LidarPoint) GetXYZ() CartesianPoint {
	azimuth := radians(p.Azimuth())
	elevAngle := radians(p.Bearing())

	cosEl := math.Cos(elevAngle)
	sinEl := math.Sin(elevAngle)
	sinAzimuth := math.Sin(azimuth)
	cosAzimuth := math.Cos(azimuth)

	distance := p.Distance()

	return CartesianPoint{
		X:         distance * cosEl * sinAzimuth,
		Y:         distance * cosEl * cosAzimuth,
		Z:         distance * sinEl,
		Intensity: uint8(p.Intensity)}
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
