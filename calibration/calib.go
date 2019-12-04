package calibration

// PointXYZ contains the X, Y, Z values of a point
type PointXYZ struct {
	X float32
	Y float32
	Z float32
}

// RotationPRY contains the pitch, roll, and yaw values of a rotation
type RotationPRY struct {
	Pitch int16
	Roll  int16
	Yaw   int16
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
		Rotation:    RotationPRY{0, 0, 0},
		Translation: PointXYZ{0.0, 0.0, 0.0}},

	"192.168.1.202": LidarCalib{
		// Rotation:    RotationPRY{-0.110758, 35.8838, 0.963915},
		Rotation:    RotationPRY{-11, 3588, 96},
		Translation: PointXYZ{698.224, -409.333, 24.5121}},

	"192.168.1.203": LidarCalib{
		// Rotation:    RotationPRY{0.452015, -35.5394, 0.0694017},
		Rotation:    RotationPRY{45, -3554, 7},
		Translation: PointXYZ{-405.472, -363.273, -44.8637}}}

// Cameras contain the Camera calibration
var Cameras = map[string]Camera{
	"front": Camera{
		Name:      "front",
		Direction: 0,
		FOV:       8200},
	"right": Camera{
		Name:      "right",
		Direction: 90,
		FOV:       30000}}

// AzimuthRange returns the fov angle range of the camera
func (c Camera) AzimuthRange() []int {
	fov2 := c.FOV / 2

	return []int{
		normalizeAngle(c.Direction - fov2),
		normalizeAngle(c.Direction + fov2)}
}

func normalizeAngle(angle int) int {
	if angle < 0 {
		angle += 36000
	}
	return angle
}

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

// Camera Parameters
// Front:
// 	H: 82
// 	V: 50
// rear:
// 	H: 61
// 	V: 47
// side:
// 	H: 185
// 	V: 150 to 160
