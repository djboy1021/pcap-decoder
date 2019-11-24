package pcapparser

import "fmt"

// LidarFrame contains one revolution of the lidar
type LidarFrame struct {
	Points      []LidarPoint
	rotation    Rotation
	translation Translation
}

// Rotation is in euler rotation form, including pitch, roll, yaw. Units in 100x degrees
type Rotation struct {
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
	X float64
	Y float64
	Z float64
}

// XYZ returns the cartesian coordinates of all points
func (lf *LidarFrame) XYZ(r Rotation, t Translation) {
	// cartesianPoints := make()

	fmt.Println(r, t)

	// for _, point := range lf.Points {
	// 	fmt.Println(point.GetXYZ())
	// }
}
