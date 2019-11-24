package pcapparser

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

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

// LocalizeFrame creates an array of all XYZ points with granularity
func (ls *LidarSource) LocalizeFrame(limits *[3][2]float64, unit float64) {
	Xr := limits[0]
	Yr := limits[1]
	Zr := limits[2]

	filename := fmt.Sprintf("./frame%d.png", ls.FrameIndex)

	m := image.NewGray16(image.Rect(int(Xr[0]/unit), int(Yr[0]/unit), int(Xr[1]/unit), int(Yr[1]/unit)))
	for _, point := range ls.CurrentFrame.Points {
		cp := point.GetXYZ()
		isWithinX := cp.X >= Xr[0] && cp.X < Xr[1]
		isWithinY := cp.Y >= Yr[0] && cp.Y < Yr[1]
		isWithinZ := cp.Z >= Zr[0] && cp.Z < Zr[1]

		if isWithinX && isWithinY && isWithinZ {
			colorIntensity := 0xFFFF * (cp.Z - Zr[0]) / (Zr[1] - Zr[0])
			m.Set(int(cp.X/unit), int(cp.Y/unit), color.Gray16{uint16(colorIntensity)})
		}
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	png.Encode(f, m)
}
