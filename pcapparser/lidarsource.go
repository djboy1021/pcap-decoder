package pcapparser

import (
	"fmt"
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
func (ls *LidarSource) GetCurrentFramePosition(limits *[3][2]float64) {
	var pixels uint16
	var start, end float64

	prevOffset := float64(ls.PreviousFrame.translation.y)
	pixels = 2048
	unit := getUnit(limits, pixels)
	if ls.PreviousFrame.Index > 0 && prevOffset > 100 {
		start = prevOffset * 0.8
		end = prevOffset * 1.2
	} else if ls.PreviousFrame.Index > 0 && prevOffset > 10 {
		start = prevOffset - 5*unit
		end = prevOffset + 5*unit
	} else {
		start = prevOffset - 10*unit
		end = prevOffset + 10*unit
	}

	pfM := ls.PreviousFrame.GetMatrix(limits, pixels, RotationAngles{}, Translation{})
	var cfM map[int]map[int]uint8

	maxAccuracy := uint(0)
	for offset := start; offset <= end; offset += unit {
		cfM = ls.CurrentFrame.GetMatrix(limits, pixels, RotationAngles{}, Translation{y: float32(offset)})
		match := getTotalMatch(pfM, cfM)

		if maxAccuracy < match {
			maxAccuracy = match
			ls.CurrentFrame.translation.y = float32(offset)
		}
	}

	fmt.Println(pixels, unit, start, prevOffset, end)
	ls.PreviousFrame.visualizeFrame(limits, pixels)
	// visualizeMap(pfM, limits, pixels, ls.PreviousFrame.Index)
	fmt.Println("\nbest offset", ls.CurrentFrame.translation.y, maxAccuracy)
}

func getTotalMatch(previousFrame map[int]map[int]uint8, currentFrame map[int]map[int]uint8) uint {
	count := uint(0)

	for row := range currentFrame {
		for col := range currentFrame[row] {
			if previousFrame[row][col] != 0 {
				count++
			}
		}
	}

	return count
}
