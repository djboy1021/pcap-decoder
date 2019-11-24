package pcapparser

// LidarSource contains the iteration info of an IP address
type LidarSource struct {
	FrameIndex        uint
	InitialAzimuth    uint16
	nextPacketAzimuth uint16
	CurrentPacket     LidarPacket
	CurrenPoints      []LidarPoint
	Buffer            []LidarPoint
}

// SetCurrentFrame sets the point cloud of a LidarSource
func (ls *LidarSource) SetCurrentFrame() {
	prevAzimuth := float64(ls.InitialAzimuth)

	for colIndex := uint8(0); colIndex < 12; colIndex++ {
		currAzimuth := ls.CurrentPacket.Blocks[colIndex].Azimuth
		var nextAzimuth uint16
		if colIndex < 10 {
			nextAzimuth = ls.CurrentPacket.Blocks[colIndex+1].Azimuth
		} else {
			nextAzimuth = ls.nextPacketAzimuth
		}

		for rowIndex := uint8(0); rowIndex < 32; rowIndex++ {
			distance := ls.CurrentPacket.Blocks[colIndex].Channels[rowIndex].Distance
			if distance == 0 {
				continue
			}

			point := LidarPoint{
				rowIndex:    rowIndex,
				productID:   ls.CurrentPacket.ProductID,
				Intensity:   ls.CurrentPacket.Blocks[colIndex].Channels[rowIndex].Reflectivity,
				distance:    distance,
				azimuth:     currAzimuth,
				nextAzimuth: nextAzimuth}

			precisionAzimuth := point.Azimuth()

			isNewFrame := isNewFrame(prevAzimuth, precisionAzimuth, currAzimuth, nextAzimuth, ls)

			if isNewFrame {
				ls.Buffer = append(ls.Buffer, point)
			} else {
				ls.CurrenPoints = append(ls.CurrenPoints, point)
			}

			prevAzimuth = precisionAzimuth
		}
	}

}

func isNewFrame(previous float64, current float64, currAzimuth uint16, nextAzimuth uint16, ls *LidarSource) bool {
	// offset := int(36000 - ls.InitialAzimuth)

	isNewFrame := nextAzimuth < currAzimuth

	// isNewFrame := len(ls.Buffer) > 0 || pPrevious-pCurrent > 10000

	return isNewFrame
}
