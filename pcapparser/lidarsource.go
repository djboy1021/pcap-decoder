package pcapparser

import (
	"fmt"
	"math"
)

// LidarSource contains the iteration info of an IP address
type LidarSource struct {
	InitialAzimuth    uint16
	nextPacketAzimuth uint16
	CurrentPacket     LidarPacket
	CurrentFrame      LidarFrame
	PreviousFrame     LidarFrame
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

// GetCurrentFramePosition locates the position of the current frame relative to the previous frame
func (ls *LidarSource) GetCurrentFramePosition(xyzRange *[3][2]float64) {
	pfM := ls.PreviousFrame.GetMatrix(xyzRange, RotationAngles{}, Translation{})
	// cfM := ls.CurrentFrame.GetMatrix(xyzRange, RotationAngles{}, Translation{})
	unit := getUnit(xyzRange)

	offsets := []float64{-20 * unit, 0, 20 * unit}
	matches := make(map[float64]uint16)

	var cfM map[int]map[int]uint8
	maxAccuracy := uint16(0)
	var bestOffset float64

	for i := 0; i < 8; i++ {
		for _, offset := range offsets {
			if matches[offset] == 0 {
				cfM = ls.CurrentFrame.GetMatrix(xyzRange, RotationAngles{}, Translation{y: float32(offset)})
				matches[offset] = getTotalMatch(pfM, cfM)
				if maxAccuracy < matches[offset] {
					maxAccuracy = matches[offset]
					bestOffset = offset
				}
			}
		}

		if matches[offsets[0]] < matches[offsets[1]] {
			offsets[2] = offsets[1]
		} else {
			offsets[0] = offsets[1]
		}
		offsets[1] = (offsets[0] + offsets[2]) / 2
		// fmt.Println(matches)
	}

	fmt.Println(maxAccuracy, bestOffset)
}

func getTotalMatch(previousFrame map[int]map[int]uint8, currentFrame map[int]map[int]uint8) uint16 {
	sum := float64(0)
	count := 0

	for row := range currentFrame {
		for col := range currentFrame[row] {
			if previousFrame[row][col] != 0 {
				count++
				sum += math.Abs(float64(previousFrame[row][col])-float64(currentFrame[row][col])) / float64(previousFrame[row][col])
			}
		}
	}

	return uint16(10000 * sum / float64(count))
}
