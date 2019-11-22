package pcapparser

import (
	"fmt"
	"math"
	"pcap-decoder/dictionary"
)

/*
General Structure
	LidarPacket
		Time
		IsDualMode
		ProductID
		Blocks[12]
			Azimuth
			Channels[32]
				Distance
				Reflectivity
*/

// LidarPacket is the raw decoded info of a lidar packet
type LidarPacket struct {
	TimeStamp  uint32       `json:"timestamp"`
	ProductID  byte         `json:"productID"`
	IsDualMode bool         `json:"isDualMode"`
	Blocks     []LidarBlock `json:"blocks"`
}

// LidarBlock contains the channels data
type LidarBlock struct {
	Azimuth  uint16         `json:"azimuth"`
	Channels []LidarChannel `json:"channels"`
}

// LidarChannel contains the basic point data
type LidarChannel struct {
	Distance     uint16 `json:"distance"`
	Reflectivity uint8  `json:"reflectivity"`
}

/*
SphericalPoint contains the point information in spherical system.
Distance is raw
Azimuth is 100x in degrees
Bearing is 1000x in degrees
Intensity is raw
*/
type SphericalPoint struct {
	LaserID   uint8
	ProductID byte
	distance  uint16
	azimuth   uint16
	Bearing   int16
	Intensity byte
}

// NewLidarPacket creates a new LidarPacket Object
func NewLidarPacket(data *[]byte) (LidarPacket, error) {
	var lp LidarPacket
	var err error

	if len(*data) == 1248 {
		blocks := make([]LidarBlock, 12)
		setBlocks(data, &blocks)

		lp = LidarPacket{
			IsDualMode: isDualMode(data),
			ProductID:  getProductID(data),
			TimeStamp:  getTime(data),
			Blocks:     blocks}
	} else {
		err = fmt.Errorf("not a lidar packet")
	}

	return lp, err
}

// SetPointCloud sets the point cloud of a ChannelInfo
func (lp *LidarPacket) SetPointCloud(nextPacketAzimuth uint16, ci *ChannelInfo) {
	prevAzimuth := ci.InitialAzimuth

	for colIndex := uint8(0); colIndex < 12; colIndex++ {
		currAzimuth := lp.Blocks[colIndex].Azimuth

		for rowIndex := uint8(0); rowIndex < 32; rowIndex++ {
			distance := lp.Blocks[colIndex].Channels[rowIndex].Distance
			if distance == 0 {
				continue
			}

			point := SphericalPoint{
				LaserID:   rowIndex,
				ProductID: lp.ProductID,
				Intensity: lp.Blocks[colIndex].Channels[rowIndex].Reflectivity,
				distance:  distance,
				azimuth:   currAzimuth,
				Bearing:   getRawElevationAngle(lp.ProductID, rowIndex)}

			prevAzimuthUint := point.azimuth * 100
			if ci.InitialAzimuth <= prevAzimuthUint && ci.InitialAzimuth > prevAzimuth {
				// New Frame
				ci.FrameIndex++
				ci.Buffer = append(ci.Buffer, point)
			} else {
				ci.CurrentFrame = append(ci.CurrentFrame, point)
			}
			fmt.Println(point.Distance(), point.ElevationAngle(), point.Azimuth())
			fmt.Println(point.GetXYZ())

			prevAzimuth = prevAzimuthUint
		}
	}

}

// GetXYZ returns the XYZ Coordinates
func (p SphericalPoint) GetXYZ() (int, int, int) {
	azimuth := p.Azimuth()

	elevAngle := p.ElevationAngle()

	cosEl := math.Cos(elevAngle)
	sinEl := math.Sin(elevAngle)
	sinAzimuth := math.Sin(azimuth)
	cosAzimuth := math.Cos(azimuth)

	distance := p.Distance()
	X := math.Round(float64(distance) * cosEl * sinAzimuth)
	Y := math.Round(float64(distance) * cosEl * cosAzimuth)
	Z := math.Round(float64(distance) * sinEl)

	return int(X), int(Y), int(Z)
}

// Distance returns the distance in mm
func (p SphericalPoint) Distance() uint32 {
	d := uint32(p.distance)
	return d << 1
}

// ElevationAngle returns the elevation angle in radians
func (p SphericalPoint) ElevationAngle() float64 {
	elevAngle := getElevationAngle(p.ProductID, p.LaserID)
	return rad(elevAngle)
}

// Azimuth returns the azimuth angle in radians
func (p SphericalPoint) Azimuth() float64 {
	var azimuthOffset float64
	if p.ProductID == 0x28 {
		azimuthOffset = float64(dictionary.VLP32AzimuthOffset[p.LaserID]) / 1000
	}
	adjustedAzimuth := azimuthOffset + float64(p.azimuth)/100

	return rad(adjustedAzimuth)
}
