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

// SphericalPoint contains the point information in spherical system.
type SphericalPoint struct {
	LaserID     uint8
	ProductID   byte
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
	// prevAzimuth := uint16(360)

	for colIndex := uint8(0); colIndex < 12; colIndex++ {
		currAzimuth := lp.Blocks[colIndex].Azimuth

		for rowIndex := uint8(0); rowIndex < 32; rowIndex++ {
			distance := lp.Blocks[colIndex].Channels[rowIndex].Distance
			if distance == 0 {
				continue
			}

			point := SphericalPoint{
				LaserID:     rowIndex,
				ProductID:   lp.ProductID,
				Intensity:   lp.Blocks[colIndex].Channels[rowIndex].Reflectivity,
				distance:    distance,
				azimuth:     currAzimuth,
				nextAzimuth: nextPacketAzimuth}

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
			fmt.Println(point.GetXYZ())
			// fmt.Println(" ", point.Azimuth())

			if len(ci.CurrentFrame) > 10 {
				panic("err")
			}

			// prevAzimuth = currPointAz
		}
	}

}

// GetXYZ returns the XYZ Coordinates
func (p SphericalPoint) GetXYZ() (int, int, int) {
	var azimuthOffset float64
	if p.ProductID == 0x28 {
		azimuthOffset = float64(dictionary.VLP32AzimuthOffset[p.LaserID]) / 1000
	}

	azimuth := rad(p.Azimuth() + azimuthOffset)
	fmt.Println(p.Azimuth() + azimuthOffset)
	elevAngle := rad(p.Bearing())

	cosEl := math.Cos(elevAngle)
	sinEl := math.Sin(elevAngle)
	sinAzimuth := math.Sin(azimuth)
	cosAzimuth := math.Cos(azimuth)

	distance := float64(p.Distance())
	X := math.Round(distance * cosEl * sinAzimuth)
	Y := math.Round(distance * cosEl * cosAzimuth)
	Z := math.Round(distance * sinEl)

	return int(X), int(Y), int(Z)
}

// Distance returns the distance in mm
func (p SphericalPoint) Distance() uint32 {
	d := uint32(p.distance)
	return d << 1
}

// Bearing returns the elevation angle in radians
func (p SphericalPoint) Bearing() float64 {
	var elevAngle float64

	switch p.ProductID {
	case 0x22:
		elevAngle = float64(dictionary.VLP16ElevationAngles[p.LaserID%16]) / 1000
	case 0x28:
		elevAngle = float64(dictionary.VLP32ElevationAngles[p.LaserID]) / 1000
	}

	return elevAngle
}

// Azimuth returns the azimuth angle in radians
func (p SphericalPoint) Azimuth() float64 {
	pAzimuth := float64(getPrecisionAzimuth(p.azimuth, p.nextAzimuth, p.LaserID, p.ProductID))
	fmt.Println(pAzimuth, p.azimuth, p.nextAzimuth, p.LaserID, p.ProductID)

	return pAzimuth
}
