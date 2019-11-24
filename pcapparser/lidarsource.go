package pcapparser

// LidarSource contains the iteration info of an IP address
type LidarSource struct {
	FrameIndex        uint
	InitialAzimuth    uint16
	nextPacketAzimuth uint16
	CurrentPacket     LidarPacket
	CurrentFrame      LidarFrame
	Buffer            []LidarPoint
}

// SetCurrentFrame sets the point cloud of a LidarSource
func (ls *LidarSource) SetCurrentFrame() {
	for colIndex := uint8(0); colIndex < 12; colIndex++ {
		currAzimuth := ls.CurrentPacket.blocks[colIndex].Azimuth
		nextAzimuth := getNextAzimuth(colIndex, ls)

		isNewFrame := isNewFrame(currAzimuth, nextAzimuth, ls)

		for rowIndex := uint8(0); rowIndex < 32; rowIndex++ {
			distance := ls.CurrentPacket.blocks[colIndex].Channels[rowIndex].Distance
			if distance > 0 {
				point := LidarPoint{
					rowIndex:    rowIndex,
					productID:   ls.CurrentPacket.ProductID,
					Intensity:   ls.CurrentPacket.blocks[colIndex].Channels[rowIndex].Reflectivity,
					distance:    distance,
					azimuth:     currAzimuth,
					nextAzimuth: nextAzimuth}

				if isNewFrame {
					ls.Buffer = append(ls.Buffer, point)
				} else {
					ls.CurrentFrame.Points = append(ls.CurrentFrame.Points, point)
				}
			}
		}
	}
}

func isNewFrame(currAzimuth uint16, nextAzimuth uint16, ls *LidarSource) bool {
	offset := int(36000 - ls.InitialAzimuth)

	pCurrAzimuth := (int(currAzimuth) + offset) % 36000
	pNextAzimuth := (int(nextAzimuth) + offset) % 36000

	return pNextAzimuth < pCurrAzimuth
}

func getNextAzimuth(colIndex uint8, ls *LidarSource) uint16 {
	var nextAzimuth uint16
	if colIndex < 10 {
		nextAzimuth = ls.CurrentPacket.blocks[colIndex+1].Azimuth
	} else {
		nextAzimuth = ls.nextPacketAzimuth
	}
	return nextAzimuth
}
