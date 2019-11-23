package pcapparser

// LidarSource contains the iteration info of an IP address
type LidarSource struct {
	FrameIndex     uint
	InitialAzimuth uint16
	CurrentPacket  LidarPacket
	CurrentFrame   []LidarPoint
	Buffer         []LidarPoint
}
