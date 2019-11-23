package pcapparser

import "fmt"

// LidarSource contains the iteration info of an IP address
type LidarSource struct {
	FrameIndex        uint
	InitialAzimuth    uint16
	nextPacketAzimuth uint16
	CurrentPacket     LidarPacket
	CurrentFrame      []LidarPoint
	Buffer            []LidarPoint
}

// SetCurrentFrame sets the point cloud of a LidarSource
func (ls *LidarSource) SetCurrentFrame() {

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
				LaserID:     rowIndex,
				productID:   ls.CurrentPacket.ProductID,
				Intensity:   ls.CurrentPacket.Blocks[colIndex].Channels[rowIndex].Reflectivity,
				distance:    distance,
				azimuth:     currAzimuth,
				nextAzimuth: nextAzimuth}

			// currPointAz := uint16(point.Azimuth() * 100)
			// fmt.Println(currPointAz, prevAzimuth, currAzimuth)
			// C := (currPointAz - ls.InitialAzimuth) % 360
			// P := (prevAzimuth - ls.InitialAzimuth) % 360
			// Check for new frame
			// if P > C+100 {
			// 	// fmt.Println(C, P, ls.FrameIndex, currPointAz, ls.InitialAzimuth, prevAzimuth)

			// 	ls.Buffer = append(ls.Buffer, point)
			// 	ls.FrameIndex++
			// } else {
			ls.CurrentFrame = append(ls.CurrentFrame, point)
			// }

			// fmt.Println(point.Distance(), point.ElevationAngle(), point.Azimuth())
			x, y, z := point.GetXYZ()
			fmt.Println(point.LaserID, point.Distance(), point.Azimuth(), point.Bearing(), x, y, z)
			// fmt.Println(" ", point.Azimuth())

			if len(ls.CurrentFrame) > 32 {
				panic("err")
			}

			// prevAzimuth = currPointAz
		}
	}

}
