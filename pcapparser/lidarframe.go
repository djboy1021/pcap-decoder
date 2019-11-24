package pcapparser

import (
	"math"
)

// LidarFrame contains one revolution of the lidar
type LidarFrame struct {
	Points      []LidarPoint
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
