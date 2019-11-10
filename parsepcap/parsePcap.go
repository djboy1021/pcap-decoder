package parsepcap

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path"
	"pcap-decoder/cli"
	"strconv"
	"strings"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// Point struct
type Point struct {
	LidarModel byte   `json:"lidarModel" bson:"lidarModel" form:"lidarModel" query:"lidarModel"` // product ID of the Lidar
	LaserID    uint8  `json:"laserID" bson:"laserID" form:"laserID" query:"laserID"`             // aka channel ID [0:32]
	Distance   uint32 `json:"distance" bson:"distance" form:"distance" query:"distance"`         // mm
	X          int    `json:"x" bson:"x" form:"x" query:"x"`                                     // mm
	Y          int    `json:"y" bson:"y" form:"y" query:"y"`                                     // mm
	Z          int    `json:"z" bson:"z" form:"z" query:"z"`                                     // mm
	Azimuth    uint16 `json:"azimuth" bson:"azimuth" form:"azimuth" query:"azimuth"`             // in degree * 100
	Intensity  uint8  `json:"intensity" bson:"intensity" form:"intensity" query:"intensity"`     // 0 to 255
	Timestamp  uint32 `json:"timestamp" bson:"timestamp" form:"timestamp" query:"timestamp"`     // µs
}

// ParsePCAP creates several go routines to start decoding the PCAP file.
func ParsePCAP() {
	var wg sync.WaitGroup
	wg.Add(int(cli.UserInput.TotalWorkers))

	for workerIndex := uint8(0); workerIndex < cli.UserInput.TotalWorkers; workerIndex++ {
		go assignWorker(workerIndex, &wg)
	}

	wg.Wait()
}

func getIPv4(pcapString string) string {
	srcIP := ""
	details := strings.Split(pcapString, "\n")

	for _, detail := range details {
		if strings.Contains(detail, "SrcIP") {
			subDetails := strings.Split(detail, " ")
			for _, subDetail := range subDetails {
				if strings.Contains(subDetail, "SrcIP") {
					srcIP = strings.Split(subDetail, "=")[1]
				}
			}
		}
	}
	return srcIP
}

type iterationInfo struct {
	frameCount     int
	isFinished     bool
	isReady        bool
	currPoints     []Point
	nextPoints     []Point
	currPacketData []byte
}

func assignWorker(workerIndex uint8, wg *sync.WaitGroup) {
	pcapFile := cli.UserInput.PcapFile
	channels := cli.UserInput.Channels
	outputFolder := cli.UserInput.OutputPath

	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		panic(err)
	}
	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()

	ip4channels := make(map[string]iterationInfo)
	for i := range channels {
		ip4channels[channels[i]] = iterationInfo{0, false, false, make([]Point, 0), make([]Point, 0), make([]byte, 0)}
	}

	totalPackets := 0
	lidarPackets := 0

	// isFinished := false
	mergedPoints := make([]Point, 0)

	var nextPacketData []byte

	for packet := range packets {
		nextPacketData = packet.Data()

		totalPackets++
		if len(nextPacketData) != 1248 {
			continue
		}

		ip4 := getIPv4(packet.String())
		channel := ip4channels[ip4]

		if len(ip4channels[ip4].currPacketData) == 1248 {
			// decode blocks
			decodeBlocks(&mergedPoints, &channel, &nextPacketData, workerIndex)
			ip4channels[ip4] = channel

			if ip4channels[ip4].isFinished {
				break
			}
		}
		channel.currPacketData = nextPacketData
		ip4channels[ip4] = channel

		areAllChannelsReady := true
		for i := range channels {
			if !ip4channels[channels[i]].isReady {
				areAllChannelsReady = false
			}
		}

		if areAllChannelsReady {
			for i := range channels {
				mergedPoints = append(mergedPoints, ip4channels[channels[i]].currPoints...)
				channel = ip4channels[channels[i]]
				channel.currPoints = channel.nextPoints
				channel.nextPoints = nil
				channel.isReady = false

				ip4channels[channels[i]] = channel
			}

			// Do something on the merged points
			basename := fmt.Sprintf("frame" + strconv.Itoa(ip4channels[channels[0]].frameCount-1))
			filename := path.Join(outputFolder, basename)
			if cli.UserInput.IsSaveAsJSON {
				savePointsToJSON(&mergedPoints, filename+".json")
			}

			// Save as image
			if cli.UserInput.IsSaveAsPNG {
				savePointsToPNG(&mergedPoints, filename+".png")
			}

			mergedPoints = nil
		}

		lidarPackets++
	}

	// fmt.Println(totalPackets, lidarPackets, frameCount)
	wg.Done()
}

func savePointsToJSON(points *[]Point, outputFileName string) {
	os.Remove(outputFileName)
	f, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	check(err)

	allPoints, _ := json.Marshal(*points)

	_, err = f.WriteString(string(allPoints))
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func savePointsToPNG(points *[]Point, filename string) {
	Xr := [2]int{-10000, 10000}
	Yr := [2]int{-10000, 10000}
	Zr := [2]int{-5000, 10000}

	unit := 20

	m := image.NewGray16(image.Rect(Xr[0]/unit, Yr[0]/unit, Xr[1]/unit, Yr[1]/unit))
	for _, point := range *points {
		isWithinX := point.X >= Xr[0] && point.X < Xr[1]
		isWithinY := point.Y >= Yr[0] && point.Y < Yr[1]
		isWithinZ := point.Z >= Zr[0] && point.Z < Zr[1]

		if isWithinX && isWithinY && isWithinZ {
			colorIntensity := 0xFFFF * (point.Z - Zr[0]) / (Zr[1] - Zr[0])
			m.Set(point.X/unit, point.Y/unit, color.Gray16{uint16(colorIntensity)})
		}
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	png.Encode(f, m)

}

func determineCloudSize(points *[]Point) {
	const MaxUint = ^uint(0)
	const MinUint = 0
	const MaxInt = int(MaxUint >> 1)
	const MinInt = -MaxInt - 1

	xMax := MinInt
	xMin := MaxInt
	yMax := MinInt
	yMin := MaxInt
	zMax := MinInt
	zMin := MaxInt
	for i := range *points {
		if (*points)[i].X > xMax {
			xMax = (*points)[i].X
		}
		if (*points)[i].X < xMin {
			xMin = (*points)[i].X
		}
		if (*points)[i].Y > yMax {
			yMax = (*points)[i].Y
		}
		if (*points)[i].Y < yMin {
			yMin = (*points)[i].Y
		}
		if (*points)[i].Z > zMax {
			zMax = (*points)[i].Z
		}
		if (*points)[i].Z < zMin {
			zMin = (*points)[i].Z
		}
	}

	fmt.Println(xMax, xMin, yMax, yMin, zMax, zMin)
}
