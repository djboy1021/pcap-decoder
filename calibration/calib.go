package calibration

// PointXYZ contains the X, Y, Z values of a point
type PointXYZ struct {
	X float64
	Y float64
	Z float64
}

// RotationPRY contains the pitch, roll, and yaw values of a rotation
type RotationPRY struct {
	Pitch float64
	Roll  float64
	Yaw   float64
}

// LidarCalib contains the rotation and translation of a lidar
type LidarCalib struct {
	Rotation    RotationPRY
	Translation PointXYZ
}

// Camera contains the rotation and translation of a lidar
type Camera struct {
	Name      string
	Direction int
	FOV       int
}

// Lidars contain the lidar calibration
var Lidars = map[string]LidarCalib{
	"192.168.1.201": LidarCalib{
		Rotation:    RotationPRY{0.0, 0.0, 0.0},
		Translation: PointXYZ{0.0, 0.0, 0.0}},

	"192.168.1.202": LidarCalib{
		Rotation:    RotationPRY{-0.110758, 35.8838, 0.963915},
		Translation: PointXYZ{0.698224, -0.409333, 0.0245121}},

	"192.168.1.203": LidarCalib{
		Rotation:    RotationPRY{0.452015, -35.5394, 0.0694017},
		Translation: PointXYZ{-0.405472, -0.363273, -0.0448637}}}

// Cameras contain the Camera calibration
var Cameras = map[string]Camera{
	"front": Camera{
		Name:      "front",
		Direction: 0,
		FOV:       8500}}

// lidars:
// - condition: "src 192.168.1.201"
// 	calib: "VLP_32C.xml"
// 	crop_mode: "Spherical"
// 	crop_outside: True
// 	crop_region: [0.0, 360.0, -90.0, 90.0, 0.0, 1.8]
// 	rotation: [0.0, 0.0, 0.0]
// 	translation: [0.0, 0.0, 0.0]
// - condition: "src 192.168.1.202"
// 	calib: "VLP-16.xml"
// 	rotation: [-0.110758, 35.8838, 0.963915]
// 	translation: [0.698224, -0.409333, 0.0245121]
// - condition: "src 192.168.1.203"
// 	calib: "VLP-16.xml"
// 	rotation: [0.452015, -35.5394, 0.0694017]
// 	translation: [-0.405472, -0.363273, -0.0448637]
