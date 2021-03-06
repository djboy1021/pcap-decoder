package pcapdecoder

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"pcap-decoder/cli"
	"sync"
)

// LidarFrame contains one revolution of the lidar
type LidarFrame struct {
	Points      []LidarPoint
	Index       uint
	rotation    RotationAngles
	translation Translation
}

// RotationAngles is in euler rotation form, including pitch, roll, yaw. Units in 100x degrees
type RotationAngles struct {
	pitch int16
	roll  int16
	yaw   int16
}

// Translation is in cartesian form, units in mm
type Translation struct {
	x float32
	y float32
	z float32
}

// CartesianPoints returns the cartesian coordinates of all points
func (lf *LidarFrame) CartesianPoints(r RotationAngles, t Translation) []CartesianPoint {
	A := getRotationMultipliers(&r)

	points := make([]CartesianPoint, len(lf.Points))

	for i := range lf.Points {
		points[i] = lf.Points[i].GetXYZ().Rotate(&A).Translate(t)
	}

	return points
}

// SphericalPoints returns the spherical coordinates of all points
func (lf *LidarFrame) SphericalPoints(r RotationAngles, t Translation) []SphericalPoint {
	A := getRotationMultipliers(&r)

	points := make([]SphericalPoint, len(lf.Points))

	for i := range lf.Points {
		if lf.Points[i].distance == 0 {
			fmt.Println(lf.Points[i])
		}
		cp := lf.Points[i].GetXYZ().Rotate(&A).Translate(t)
		points[i] = cp.ToSpherical()
	}

	return points
}

func getRotationMultipliers(r *RotationAngles) [3][3]float64 {
	pitch := radians(float64(r.pitch) / 100)
	roll := radians(float64(r.roll) / 100)
	yaw := radians(float64(r.yaw) / 100)

	cosa := math.Cos(yaw)
	sina := math.Sin(yaw)

	cosb := math.Cos(pitch)
	sinb := math.Sin(pitch)

	cosc := math.Cos(roll)
	sinc := math.Sin(roll)

	var Axx = cosa * cosb
	var Axy = cosa*sinb*sinc - sina*cosc
	var Axz = cosa*sinb*cosc + sina*sinc

	var Ayx = sina * cosb
	var Ayy = sina*sinb*sinc + cosa*cosc
	var Ayz = sina*sinb*cosc - cosa*sinc

	var Azx = -sinb
	var Azy = cosb * sinc
	var Azz = cosb * cosc

	return [3][3]float64{
		{Axx, Axy, Axz},
		{Ayx, Ayy, Ayz},
		{Azx, Azy, Azz}}

}

// GetMatrix returns an array of all XYZ points with granularity
func (lf *LidarFrame) GetMatrix(limits *[3][2]float64, pixels uint16, rotAngles RotationAngles, trans Translation) map[int]map[int]uint8 {
	Xr := limits[0]
	Yr := limits[1]
	Zr := limits[2]

	unit := getUnit(limits, pixels)

	// allocate space for matrix
	frameMap := make(map[int]map[int]uint8)

	points := lf.CartesianPoints(rotAngles, trans)

	for _, cp := range points {
		isWithinX := cp.X >= Xr[0] && cp.X < Xr[1]
		isWithinY := cp.Y >= Yr[0] && cp.Y < Yr[1]
		isWithinZ := cp.Z >= Zr[0] && cp.Z < Zr[1]

		if isWithinX && isWithinY && isWithinZ {
			xInd := int(cp.X / unit)
			yInd := int(cp.Y / unit)

			colorIntensity := uint16(0xFF * (cp.Z - Zr[0]) / (Zr[1] - Zr[0]))

			if frameMap[xInd] != nil && frameMap[xInd][yInd] < uint8(colorIntensity) {
				frameMap[xInd][yInd] = uint8(colorIntensity)
			} else {
				frameMap[xInd] = map[int]uint8{}
				frameMap[xInd][yInd] = uint8(colorIntensity)
			}
		}
	}

	return frameMap
}

func (lf *LidarFrame) visualizeFrame(limits *[3][2]float64, pixels uint16) {
	index := lf.Index

	Xr := limits[0]
	Yr := limits[1]
	// Zr := limits[2]

	unit := getUnit(limits, pixels)

	filename := fmt.Sprintf("./frame%d.png", index)

	m := image.NewRGBA64(image.Rect(int(Xr[0]/unit), int(Yr[0]/unit), int(Xr[1]/unit), int(Yr[1]/unit)))

	frameMap := lf.GetMatrix(limits, pixels, RotationAngles{}, lf.translation)

	for xInd := range frameMap {
		for yInd := range frameMap[xInd] {
			m.Set(xInd, yInd, color.RGBA{
				0,
				255,
				0,
				255})
		}
	}

	f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, m)
}

// ToJSON saves the lidar points in json format
func (lf *LidarFrame) ToJSON(isSave bool) []byte {
	pointsLen := len(lf.Points)
	points := make([]CartesianPoint, pointsLen)

	// Waitgroup for storing points
	var wg sync.WaitGroup
	wg.Add(pointsLen)

	// Convert points to Cartesian points
	for index, p := range lf.Points {
		go appendToPoints(&points, index, &p, &wg)
	}
	wg.Wait()

	allPoints, _ := json.Marshal(points)

	if isSave {
		// Save to JSON
		outputFileName := fmt.Sprintf("frame%d.json", lf.Index)
		outputFileName = filepath.Join(cli.UserInput.OutputPath, outputFileName)

		os.Remove(outputFileName)
		f, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		check(err)

		_, err = f.WriteString(string(allPoints))
		check(err)
	}

	return allPoints
}

func appendToPoints(points *[]CartesianPoint, index int, point *LidarPoint, wg *sync.WaitGroup) {
	xyzi := (*point).GetXYZ()

	(*points)[index] = CartesianPoint{
		X:         math.Round(xyzi.X*100) / 100,
		Y:         math.Round(xyzi.Y*100) / 100,
		Z:         math.Round(xyzi.Z*100) / 100,
		Intensity: xyzi.Intensity}
	wg.Done()
}

func getUnit(limits *[3][2]float64, pixels uint16) float64 {
	Xr := limits[0]
	Yr := limits[1]

	var unit float64
	xDiff := Xr[1] - Xr[0]
	yDiff := Yr[1] - Yr[0]
	if xDiff > yDiff {
		unit = xDiff / float64(pixels)
	} else {
		unit = yDiff / float64(pixels)
	}

	return unit
}
