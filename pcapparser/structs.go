package pcapparser

import (
	"fmt"
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

// SetPointCloud sets the point cloud of a LidarSource
func (lp *LidarPacket) SetPointCloud(nextPacketAzimuth uint16, ci *LidarSource) {

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

			point := LidarPoint{
				LaserID:     rowIndex,
				productID:   lp.ProductID,
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
