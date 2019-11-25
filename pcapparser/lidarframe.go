package pcapparser

import (
	"fmt"
	"math"
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

// CartesianPoint contains X, Y, Z
type CartesianPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// XYZ returns the cartesian coordinates of all points
func (lf *LidarFrame) XYZ(r RotationAngles, t Translation) []CartesianPoint {
	A := getRotationMultipliers(&r)

	points := make([]CartesianPoint, len(lf.Points))

	for i := range lf.Points {
		points[i] = lf.Points[i].GetXYZ().Rotate(&A).Translate(t)
	}

	return points
}

func getRotationMultipliers(r *RotationAngles) [3][3]float64 {
	yaw := rad(float64(r.yaw) / 100)
	pitch := rad(float64(r.pitch) / 100)
	roll := rad(float64(r.roll) / 100)

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

// Rotate returns the rotated cartesian point
func (cp CartesianPoint) Rotate(A *[3][3]float64) CartesianPoint {
	return CartesianPoint{
		X: A[0][0]*cp.X + A[0][1]*cp.Y + A[0][2]*cp.Z,
		Y: A[1][0]*cp.X + A[1][1]*cp.Y + A[1][2]*cp.Z,
		Z: A[2][0]*cp.X + A[2][1]*cp.Y + A[2][2]*cp.Z}
}

// Translate returns the translated CartesianPoint position
func (cp CartesianPoint) Translate(t Translation) CartesianPoint {
	return CartesianPoint{
		X: cp.X + float64(t.x),
		Y: cp.Y + float64(t.y),
		Z: cp.Z + float64(t.z)}
}

func getUnit(limits *[3][2]float64) float64 {
	Xr := limits[0]
	Yr := limits[1]

	var unit float64
	xDiff := Xr[1] - Xr[0]
	yDiff := Yr[1] - Yr[0]
	if xDiff > yDiff {
		unit = xDiff / 512
	} else {
		unit = yDiff / 512
	}

	return unit
}

// GetMatrix returns an array of all XYZ points with granularity
func (lf *LidarFrame) GetMatrix(limits *[3][2]float64, rotAngles RotationAngles, trans Translation) map[int]map[int]uint8 {
	Xr := limits[0]
	Yr := limits[1]
	Zr := limits[2]

	var unit float64
	xDiff := Xr[1] - Xr[0]
	yDiff := Yr[1] - Yr[0]
	if xDiff > yDiff {
		unit = xDiff / 512
	} else {
		unit = yDiff / 512
	}

	// filename := fmt.Sprintf("./frame%d.png", lf.Index)
	// m := image.NewGray16(image.Rect(int(Xr[0]/unit), int(Yr[0]/unit), int(Xr[1]/unit), int(Yr[1]/unit)))

	// allocate space for matrix
	currPNGMatrix := make(map[int]map[int]uint8)

	points := lf.XYZ(rotAngles, trans)

	for _, cp := range points {
		isWithinX := cp.X >= Xr[0] && cp.X < Xr[1]
		isWithinY := cp.Y >= Yr[0] && cp.Y < Yr[1]
		isWithinZ := cp.Z >= Zr[0] && cp.Z < Zr[1]

		if isWithinX && isWithinY && isWithinZ {
			xInd := int(cp.X / unit)
			yInd := int(cp.Y / unit)

			colorIntensity := uint16(0xFF * (cp.Z - Zr[0]) / (Zr[1] - Zr[0]))

			if currPNGMatrix[xInd] != nil && currPNGMatrix[xInd][yInd] < uint8(colorIntensity) {
				currPNGMatrix[xInd][yInd] = uint8(colorIntensity)
			} else {
				currPNGMatrix[xInd] = map[int]uint8{}
				currPNGMatrix[xInd][yInd] = uint8(colorIntensity)
			}

			// fmt.Println(xInd, yInd, currPNGMatrix[xInd][yInd])

			// m.Set(int(cp.X/unit), int(cp.Y/unit), color.Gray16{uint16(colorIntensity)})
		}
	}

	// f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer f.Close()
	// png.Encode(f, m)

	return currPNGMatrix

}

func visualizeMap(frameMap map[int]map[int]uint8, index uint) {
	// filename := fmt.Sprintf("./frame%d.png", index)
	fmt.Println(frameMap)
}
