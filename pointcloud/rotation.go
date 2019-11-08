package pointcloud

import "math"

// CartesianPoint struct
type CartesianPoint struct {
	X float64
	Y float64
	Z float64
}

// Rotate gives the coordinates of a point when rotated
// The input angles of rotation, pitch, roll, yaw should be in radians
func (point CartesianPoint) Rotate(pitch float64, roll float64, yaw float64) CartesianPoint {
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

	var px = point.X
	var py = point.Y
	var pz = point.Z

	return CartesianPoint{Axx*px + Axy*py + Axz*pz,
		Ayx*px + Ayy*py + Ayz*pz,
		Azx*px + Azy*py + Azz*pz}
}
