package parsepcap

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
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
func ParsePCAP(pcapFile *string, outputPath *string, totalWorkers uint8, startFrame int, endFrame int, channels []string, isSaveAsJSON bool) {
	var wg sync.WaitGroup
	wg.Add(int(totalWorkers))

	for workerIndex := uint8(0); workerIndex < totalWorkers; workerIndex++ {
		go assignWorker(*pcapFile, workerIndex, totalWorkers, isSaveAsJSON, outputPath, startFrame, endFrame, channels, &wg)
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

func assignWorker(pcapFile string, workerIndex uint8, totalWorkers uint8, isSaveAsJSON bool, outputFolder *string, startFrame int, endFrame int, ip4s []string, wg *sync.WaitGroup) {
	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		panic(err)
	}
	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()

	ip4channels := make(map[string]iterationInfo)
	for i := range ip4s {
		ip4channels[ip4s[i]] = iterationInfo{0, false, false, make([]Point, 0), make([]Point, 0), make([]byte, 0)}
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
			decodeBlocks(&mergedPoints, &channel, &nextPacketData, workerIndex, totalWorkers, &endFrame, &startFrame)
			ip4channels[ip4] = channel

			if ip4channels[ip4].isFinished {
				break
			}
		}
		channel.currPacketData = nextPacketData
		ip4channels[ip4] = channel

		areAllChannelsReady := true
		for i := range ip4s {
			if !ip4channels[ip4s[i]].isReady {
				areAllChannelsReady = false
			}
		}

		if areAllChannelsReady {
			for i := range ip4s {
				mergedPoints = append(mergedPoints, ip4channels[ip4s[i]].currPoints...)
				channel = ip4channels[ip4s[i]]
				channel.currPoints = channel.nextPoints
				channel.nextPoints = nil
				channel.isReady = false

				ip4channels[ip4s[i]] = channel
			}

			// Do something on the merged points
			if isSaveAsJSON {
				basename := fmt.Sprintf("frame" + strconv.Itoa(ip4channels[ip4s[0]].frameCount-1))
				filename := path.Join(*outputFolder, basename+".json")
				savePointsToJSON(&mergedPoints, filename)
			}
			// Save as image
			savePointsToPNG(&mergedPoints)

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

func savePointsToPNG(points *[]Point) {
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
