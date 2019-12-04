package pcapparser

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"pcap-decoder/calibration"
)

// LidarSource contains the iteration info of an IP address
type LidarSource struct {
	address           string
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

// LocalizeCurrentFrame locates the translation and rotation offset of the current frame relative to the previous frame
func (ls *LidarSource) LocalizeCurrentFrame(limits *[3][2]float64) {

	pixels := uint16(512)
	index := ls.CurrentFrame.Index

	Xr := limits[0]
	Yr := limits[1]
	unit := getUnit(limits, pixels)

	cfm := ls.CurrentFrame.GetMatrix(limits, pixels, RotationAngles{}, Translation{})

	trans := Translation{}

	yOffset := ls.getBestFit(limits, pixels, "y", trans)
	trans.y = yOffset
	xOffset := ls.getBestFit(limits, pixels, "x", trans)

	fmt.Println(xOffset, yOffset)

	pfm := ls.PreviousFrame.GetMatrix(limits, pixels, RotationAngles{}, Translation{y: yOffset})
	m := image.NewRGBA64(image.Rect(int(Xr[0]/unit), int(Yr[0]/unit), int(Xr[1]/unit), int(Yr[1]/unit)))
	for xInd := range cfm {
		for yInd := range cfm[xInd] {
			m.Set(xInd, yInd, color.RGBA{
				0,
				255,
				0,
				255})
		}
	}
	for xInd := range pfm {
		for yInd := range pfm[xInd] {
			m.Set(xInd, yInd, color.RGBA{
				255,
				0,
				0,
				255})
		}
	}

	filename := fmt.Sprintf("./frame%d.png", index)
	f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, m)

	// panic("Temp")
}

func (ls *LidarSource) getBestFit(limits *[3][2]float64, pixels uint16, dir string, defTrans Translation) float32 {
	maxCount := uint(0)
	bestFit := float32(0)

	isForward := true
	forwardCount := uint(0)
	isBackward := true
	backwardCount := uint(0)

	var cfm map[int]map[int]uint8
	pfm := ls.PreviousFrame.GetMatrix(limits, pixels, RotationAngles{}, Translation{})

	for offset := float32(0); isForward || isBackward; offset++ {

		if isForward {
			switch dir {
			case "x":
				defTrans.x = offset
			case "y":
				defTrans.y = offset
			case "z":
				defTrans.z = offset
			}
			cfm = ls.CurrentFrame.GetMatrix(limits, pixels, RotationAngles{}, defTrans)

			forwardCount = getTotalMatch(pfm, cfm)

			if forwardCount > maxCount {
				maxCount = forwardCount
				bestFit = offset
			}

		}

		if isBackward {
			switch dir {
			case "x":
				defTrans.x = -offset
			case "y":
				defTrans.y = -offset
			case "z":
				defTrans.z = -offset
			}
			cfm = ls.CurrentFrame.GetMatrix(limits, pixels, RotationAngles{}, defTrans)

			backwardCount = getTotalMatch(pfm, cfm)

			if backwardCount > maxCount {
				maxCount = backwardCount
				bestFit = -offset
			}
		}

		if maxCount > forwardCount<<1 || backwardCount > forwardCount<<1 {
			isForward = false
		}
		if maxCount > backwardCount<<1 || forwardCount > backwardCount<<1 {
			isBackward = false
		}

		// fmt.Println(offset, maxCount, "forward", isForward, forwardCount, "backward", isBackward, backwardCount)

	}

	return bestFit
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

func (ls *LidarSource) elevationView() {
	camera := calibration.Cameras["front"]

	Hr := []float64{-1500, 2500}
	Ar := []int{camera.Direction - camera.FOV/2, camera.Direction + camera.FOV/2}
	Dr := []int{0, 20000}

	arLen := Ar[1] - Ar[0]
	if arLen < 0 {
		arLen = 36000 + arLen
	}

	imgWidth := 4096 / 8
	imgHeight := 2176 / 8

	unitH := (Hr[1] - Hr[0]) / float64(imgHeight)

	m := image.NewRGBA64(image.Rect(0, int(Hr[0]/unitH), imgWidth, int(Hr[1]/unitH)))

	for _, point := range ls.CurrentFrame.Points {
		distance := point.Distance()
		bearing := point.Bearing()

		// Check if azimuth is within range
		azimuth100 := point.Azimuth() * 100
		if int(azimuth100) < Ar[0] && int(azimuth100) > Ar[1] {
			continue
		}

		azimuth := (int(math.Round(azimuth100)) + 36000 - Ar[0]) % 36000
		xInd := int(float32(imgWidth) * float32(azimuth) / float32(arLen))

		height := distance * math.Sin(rad(bearing))
		depth := int(math.Round(distance * math.Cos(rad(bearing))))

		if height < Hr[1] && height > Hr[0] && depth < Dr[1] {
			c := uint32(float32(Dr[1]-depth) * 0xFFFFFF / float32(Dr[1]))
			r := uint8((c & 0xFF0000) >> 16)
			g := uint8((c & 0x00FF00) >> 8)
			b := uint8(c & 0x0000FF)

			invHeight := Hr[1] + Hr[0] - height

			m.Set(
				xInd,
				int(invHeight/unitH),
				color.RGBA{r, g, b, 255})
		} else {
			// fmt.Println(height)
		}
	}
	filename := fmt.Sprintf("./%s-elev%d.png", ls.address, ls.CurrentFrame.Index)
	fmt.Println(filename)

	f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, m)

}
