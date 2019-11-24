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
	blocks     []LidarBlock `json:"blocks"`
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
			blocks:     blocks}
	} else {
		err = fmt.Errorf("not a lidar packet")
	}

	return lp, err
}
