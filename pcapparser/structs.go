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
		productID
		Blocks[12]
			Azimuth
			Channels[32]
				Distance
				Reflectivity
*/

// LidarPacket is the raw decoded info of a lidar packet
type LidarPacket struct {
	TimeStamp  uint32       `json:"timestamp"`
	productID  byte         `json:"productID"`
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

// SphericalPoint contains the point information in spherical system.
type SphericalPoint struct {
	LaserID     uint8
	productID   byte
	distance    uint16
	azimuth     uint16
	nextAzimuth uint16
	Intensity   byte
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
			productID:  getProductID(data),
			TimeStamp:  getTime(data),
			Blocks:     blocks}
	} else {
		err = fmt.Errorf("not a lidar packet")
	}

	return lp, err
}

// SetPointCloud sets the point cloud of a ChannelInfo
func (lp *LidarPacket) SetPointCloud(nextPacketAzimuth uint16, ci *ChannelInfo) {
	// prevAzimuth := uint16(360)

	for colIndex := uint8(0); colIndex < 12; colIndex++ {
		currAzimuth := lp.Blocks[colIndex].Azimuth
		var nextAzimuth uint16
		if colIndex < 10 {
			nextAzimuth = lp.Blocks[colIndex+1].Azimuth
		} else {
			nextAzimuth = nextPacketAzimuth
		}

		for rowIndex := uint8(0); rowIndex < 32; rowIndex++ {
			distance := lp.Blocks[colIndex].Channels[rowIndex].Distance
			if distance == 0 {
				continue
			}

			point := SphericalPoint{
				LaserID:     rowIndex,
				productID:   lp.productID,
				Intensity:   lp.Blocks[colIndex].Channels[rowIndex].Reflectivity,
				distance:    distance,
				azimuth:     currAzimuth,
				nextAzimuth: nextAzimuth}

			// currPointAz := uint16(point.Azimuth() * 100)
			// fmt.Println(currPointAz, prevAzimuth, currAzimuth)
			// C := (currPointAz - ci.InitialAzimuth) % 360
			// P := (prevAzimuth - ci.InitialAzimuth) % 360
			// Check for new frame
			// if P > C+100 {
			// 	// fmt.Println(C, P, ci.FrameIndex, currPointAz, ci.InitialAzimuth, prevAzimuth)

			// 	ci.Buffer = append(ci.Buffer, point)
			// 	ci.FrameIndex++
			// } else {
			ci.CurrentFrame = append(ci.CurrentFrame, point)
			// }

			// fmt.Println(point.Distance(), point.ElevationAngle(), point.Azimuth())
			x, y, z := point.GetXYZ()
			fmt.Println(point.LaserID, point.Distance(), point.Azimuth(), point.Bearing(), x, y, z)
			// fmt.Println(" ", point.Azimuth())

			if len(ci.CurrentFrame) > 32 {
				panic("err")
			}

			// prevAzimuth = currPointAz
		}
	}

}

// GetXYZ returns the XYZ Coordinates
func (p SphericalPoint) GetXYZ() (float64, float64, float64) {
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
func (p SphericalPoint) Distance() float64 {
	return 2 * float64(p.distance)
}

// Bearing returns the elevation angle in radians
func (p SphericalPoint) Bearing() float64 {
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
func (p SphericalPoint) Azimuth() float64 {
	var azimuthOffset float64

	if p.productID == 0x28 {
		azimuthOffset = float64(dictionary.VLP32AzimuthOffset[p.LaserID]) / 1000
	}

	azimuthGap := getAzimuthGap(p.azimuth, p.nextAzimuth)
	angleTimeOffset := getAngleTimeOffset(p.productID, p.LaserID, azimuthGap)
	precisionAzimuth := p.azimuth + angleTimeOffset

	return float64(precisionAzimuth)/100 + azimuthOffset
}
