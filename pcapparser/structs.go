package pcapparser

import "fmt"

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

// SphericalPoint contains the point information in spherical system
type SphericalPoint struct {
	Distance  uint16
	Azimuth   uint16
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
		var nextAzimuth uint16
		if colIndex < 11 {
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
				Intensity: lp.Blocks[colIndex].Channels[rowIndex].Reflectivity,
				Distance:  distance,
				Azimuth:   getPrecisionAzimuth(currAzimuth, nextAzimuth, rowIndex, lp.ProductID),
				Bearing:   getRawElevationAngle(lp.ProductID, rowIndex)}

			if ci.InitialAzimuth <= point.Azimuth && ci.InitialAzimuth > prevAzimuth {
				// New Frame
				ci.FrameIndex++
				ci.Buffer = append(ci.Buffer, point)
			} else {
				ci.CurrentFrame = append(ci.CurrentFrame, point)
			}

			prevAzimuth = point.Azimuth
		}
	}

}
