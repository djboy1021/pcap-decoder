package pcapdecoder

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
	Address           string
	InitialAzimuth    uint16
	NextPacketAzimuth uint16
	CurrentPacket     LidarPacket
	CurrentFrame      LidarFrame
	PreviousFrame     LidarFrame
	Buffer            []LidarPoint
	Calibration       calibration.LidarCalib
}

// SetCurrentFrame sets the point cloud of a LidarSource
func (ls *LidarSource) SetCurrentFrame(frameIndex uint) {

	for colIndex := uint8(0); colIndex < 12; colIndex++ {
		currAzimuth := ls.CurrentPacket.Blocks[colIndex].Azimuth
		nextAzimuth := getNextAzimuth(colIndex, ls)

		isNewFrame := isNewFrame(currAzimuth, nextAzimuth, ls)
		if isNewFrame {
			ls.CurrentFrame.Index++
		}

		//if ls.CurrentFrame.Index != frameIndex {
		//	continue
		//}

		//curAzimuthPoints := make([]float64, 0)
		//weights := make([]float64, 0)

		for rowIndex := uint8(0); rowIndex < 32; rowIndex++ {
			distance := ls.CurrentPacket.Blocks[colIndex].Channels[rowIndex].Distance
			//weights = append(weights, 1)
			//curAzimuthPoints = append(curAzimuthPoints, float64(distance)*2)
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

			if isNewFrame {
				ls.Buffer = append(ls.Buffer, point)
			} else {
				ls.CurrentFrame.Points = append(ls.CurrentFrame.Points, point)
			}
			//fmt.Println(ls.CurrentFrame.Index, len(ls.Buffer), len(ls.CurrentFrame.Points))
		}

		//mean, stdDev := stat.MeanStdDev(curAzimuthPoints, weights)
		//fmt.Println(mean, stdDev, curAzimuthPoints)
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
		nextAzimuth = ls.CurrentPacket.Blocks[colIndex+1].Azimuth
	} else {
		nextAzimuth = ls.NextPacketAzimuth
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

func (ls *LidarSource) elevationView(cameraName string, imgWidth int, imgHeight int) {

	//fmt.Println(ls.CurrentFrame.Index)

	camera := calibration.Cameras[cameraName]

	Ar := camera.AzimuthRange()
	Hr := []float64{-1500, 2500}
	Dr := []int{0, 10000}

	arLen := Ar[1] - Ar[0]
	if arLen < 0 {
		arLen = 36000 + arLen
	}

	unitH := (Hr[1] - Hr[0]) / float64(imgHeight)

	m := image.NewRGBA64(image.Rect(0, int(Hr[0]/unitH), imgWidth, int(Hr[1]/unitH)))

	// rotation := RotationAngles{
	// 	pitch: ls.Calibration.Rotation.Yaw,
	// 	roll:  ls.Calibration.Rotation.Pitch,
	// 	yaw:   ls.Calibration.Rotation.Roll}
	rotation := RotationAngles{}

	A := getRotationMultipliers(&rotation)

	// p := CartesianPoint{X: 1, Y: 1, Z: 1}
	// fmt.Println(p, p.Rotate(&A), A)
	// translation := Translation{x: 0, y: 0, z: 0}

	var distance, bearing, azimuth100 float64

	for _, point := range ls.CurrentFrame.Points {
		if rotation.pitch == 0 && rotation.roll == 0 && rotation.yaw == 0 {
			distance = point.Distance()
			bearing = point.Bearing()
			azimuth100 = point.Azimuth() * 100
		} else {
			sp := point.GetXYZ().Rotate(&A).ToSpherical()

			distance = sp.Radius
			bearing = sp.Bearing
			azimuth100 = sp.Azimuth * 100
		}

		// fmt.Println(distance, azimuth100, bearing, point.GetXYZ().Rotate(&A).ToSpherical())

		// Check if azimuth is within range
		if int(azimuth100) < Ar[0] && int(azimuth100) > Ar[1] {
			continue
		}

		azimuth := (int(math.Round(azimuth100)) + 36000 - Ar[0]) % 36000
		xInd := int(float32(imgWidth) * float32(azimuth) / float32(arLen))

		height := distance * math.Sin(radians(bearing))
		depth := int(math.Round(distance * math.Cos(radians(bearing))))

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

		}
	}

	filename := fmt.Sprintf("./%s-elev%d.png", ls.Address, ls.CurrentFrame.Index)
	fmt.Println(filename)

	f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, m)
}
