package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// Packet structure
type Packet struct {
	points  []Point
	azimuth uint16
}

var isDualMode bool
var timingOffsetTable [32][12]uint32

// Point structure
type Point struct {
	LaserID   uint8  `json:"laserID" bson:"laserID" form:"laserID" query:"laserID"`         // aka channel ID [0:32]
	Distance  uint32 `json:"distance" bson:"distance" form:"distance" query:"distance"`     // mm
	X         int16  `json:"x" bson:"x" form:"x" query:"x"`                                 // mm
	Y         int16  `json:"y" bson:"y" form:"y" query:"y"`                                 // mm
	Z         int16  `json:"z" bson:"z" form:"z" query:"z"`                                 // mm
	Azimuth   uint16 `json:"azimuth" bson:"azimuth" form:"azimuth" query:"azimuth"`         // in degree * 100
	Intensity uint8  `json:"intensity" bson:"intensity" form:"intensity" query:"intensity"` // 0 to 255
	Timestamp uint32 `json:"timestamp" bson:"timestamp" form:"timestamp" query:"timestamp"` // µs
}

func getElevationAngle(model byte, laserID uint8) float32 {
	if model == 0x28 {
		vlp32 := []float32{-25, -1, -1.667, -15.639, -11.31, 0, -0.667, -8.843, -7.254, 0.333, -0.333, -6.148,
			-5.333, 1.333, 0.667, -4, -4.667, 1.667, 1, -3.667, -3.333, 3.333, 2.333, -2.667, -3, -7, 4.667, -2.333, -2, 15, 10.333, -1.333}
		return vlp32[laserID]
	} else if model == 0x22 {
		vlp16 := []float32{-15, 1, -13, -3, -11, 5, -9, 7, -7, 9, -5, 11, -3, 13, -1, 15}
		return vlp16[laserID]
	}
	return 0
}

func getAzimuthOffset(model byte, laserID uint8) float32 {
	if model == 0x28 {
		vlp32 := []float32{1.4, -4.2, 1.4, -1.4, 1.4, -1.4, 4.2,
			-1.4, 1.4, -4.2, 1.4, -1.4, 4.2, -1.4, 4.2, -1.4, 1.4,
			4.2, 1.4, -4.2, 4.2, -1.4, 1.4, -1.4, 1.4, -1.4, 1.4,
			-4.2, 4.2, -1.4, 1.4, -1.4}
		return vlp32[laserID]
	}
	return 0
}

func getIsDualMode(factory *[]byte) bool {
	if (*factory)[0] == 57 {
		return true
	}
	return false
}

func appendFile(filename *string, contents string) {
	f, _ := os.OpenFile(*filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if _, err := f.WriteString(contents); err != nil {
		log.Println(err)
	}
	f.Close()
}

func getTime(time []byte) uint32 {
	return binary.LittleEndian.Uint32(time)
}

func getAzimuth(block []byte) uint16 {
	angle := binary.LittleEndian.Uint16(block)
	if angle > 36000 {
		return angle - 36000
	}
	return angle
}

func getDistance(block []byte) uint32 {
	return uint32(block[1])<<8 + uint32(block[0])
}

func rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func makeTimingOffsetTable(isDualMode bool) [32][12]uint32 {
	var timingOffsets [32][12]uint32

	// unit is µs (microsec)
	fullFiringCycle := float32(55.296)
	singleFiring := float32(2.304)

	dataBlockIndex := 0
	dataPointIndex := 0
	for x := 0; x < 12; x++ {
		for y := 0; y < 32; y++ {
			if isDualMode {
				dataBlockIndex = x / 2
			} else {
				dataBlockIndex = x
			}
			dataPointIndex = y / 2
			offset := fullFiringCycle*float32(dataBlockIndex) + singleFiring*float32(dataPointIndex)

			timingOffsets[y][x] = uint32(math.Round(float64(offset)))
		}
	}

	return timingOffsets
}

func check(e error) {
	if e != nil {
		panic(e)
	}
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

func generatePoint(distance *uint32, intensity byte, adjustedTimeStamp *uint32, nextAzimuth *uint16, azimuth *uint16, laserID uint8, productID *byte) Point {
	// Calculate Precision Azimuth
	var azimuthGap uint16
	if *nextAzimuth < *azimuth {
		azimuthGap = (*nextAzimuth + 36000 - *azimuth)
	} else {
		azimuthGap = (*nextAzimuth - *azimuth)
	}

	K := float64((1 + laserID) / 2)
	precisionAzimuth := *azimuth + uint16(math.Round(float64(azimuthGap)*K*0.04166667)) // 2.304/55.296 = 0.0416667
	if precisionAzimuth >= 36000 {
		precisionAzimuth -= 36000
	}

	// Get elevation Angle
	elevAngle := getElevationAngle(*productID, laserID)
	azimuthOffset := getAzimuthOffset(*productID, laserID)

	// Intermmediary calculations
	cosEl := math.Cos(rad(float64(elevAngle)))
	sinEl := math.Sin(rad(float64(elevAngle)))
	sinAzimuth := math.Sin(rad(float64(azimuthOffset) + float64(precisionAzimuth)/100))
	cosAzimuth := math.Cos(rad(float64(azimuthOffset) + float64(precisionAzimuth)/100))

	return Point{Distance: *distance,
		X:         int16(float64(*distance) * cosEl * sinAzimuth),
		Y:         int16(float64(*distance) * cosEl * cosAzimuth),
		Z:         int16(float64(*distance) * sinEl),
		Intensity: intensity,
		Timestamp: *adjustedTimeStamp,
		Azimuth:   precisionAzimuth,
		LaserID:   laserID}
}

func decodeChannel(data *[]byte, productID *byte, blkIndex *int, azimuth *uint16, nextAzimuth *uint16, firingTime *uint32, prevAzimuth *uint16, nextFramePoints *[]Point, currFramePoints *[]Point) {
	block := (*data)[*blkIndex : *blkIndex+100]
	blkID := uint8(0)

	blockPointer := 4
	for blkID < 32 {
		distance := getDistance(block[blockPointer:blockPointer+2]) << 1

		if distance > 0 {
			blockID := (*blkIndex - 42) / 100
			adjustedTimeStamp := timingOffsetTable[blkID][blockID] + *firingTime
			intensity := block[blockPointer+2]
			// New Point
			newPoint := generatePoint(&distance, intensity, &adjustedTimeStamp, nextAzimuth, azimuth, blkID, productID)

			if *azimuth <= *prevAzimuth {
				*nextFramePoints = append(*currFramePoints, newPoint)
			} else {
				*currFramePoints = append(*currFramePoints, newPoint)
			}
		}

		blockPointer += 3
		blkID++
	}
}

func decodeVLP32Blocks(data *[]byte, isFinished *bool, frameCount *int,
	totalWorkers *uint8, workerIndex *uint8,
	startFrame *int, endFrame *int,
	azimuth *uint16, prevAzimuth *uint16,
	nextFramePoints *[]Point, currFramePoints *[]Point,
	outputFolder *string) {

	firingTime := getTime((*data)[1242:1246])
	productID := (*data)[1247]
	nextAzimuth := uint16(36000)

	// Decode block
	for blkIndex := 42; blkIndex < 1242; blkIndex += 100 {
		*isFinished = (*frameCount >= *endFrame) && (*endFrame > 0)
		*azimuth = getAzimuth((*data)[blkIndex+2 : blkIndex+4])
		isDecode := *frameCount%int(*totalWorkers) == int(*workerIndex)
		isDecode = isDecode && *frameCount >= *startFrame

		if blkIndex < 1142 {
			nextAzimuth = getAzimuth((*data)[blkIndex+102 : blkIndex+104])
		} else {
			nextAzimuth = *azimuth + 23 // get the azimuth of next packet
		}

		if *isFinished {
			break
		}

		if isDecode {
			decodeChannel(data, &productID, &blkIndex, azimuth, &nextAzimuth, &firingTime, prevAzimuth, nextFramePoints, currFramePoints)
		}

		// on new frame
		if *prevAzimuth > *azimuth {
			// Post process points
			if isDecode {
				basename := fmt.Sprintf("vlp32_frame" + strconv.Itoa(*frameCount))
				filename := path.Join(*outputFolder, basename+".json")
				savePointsToJSON(currFramePoints, filename)
				go fmt.Println(*frameCount, len(*currFramePoints), *azimuth, nextAzimuth)
			}

			// Reset frame's points arrays
			*currFramePoints = *nextFramePoints
			*nextFramePoints = make([]Point, 0)
			*frameCount++
		}
		*prevAzimuth = *azimuth
	}
}

func assignWorker(pcapFile string, lidarModel byte, workerIndex uint8, totalWorkers uint8, isSaveAsJSON bool, outputFolder string, startFrame int, endFrame int, wg *sync.WaitGroup) {
	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		panic(err)
	}
	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()

	isDualMode = false
	timingOffsetTable = makeTimingOffsetTable(isDualMode)

	totalPackets := 0
	lidarPackets := 0
	frameCount := 0
	currFramePoints := make([]Point, 0)
	nextFramePoints := make([]Point, 0)
	prevAzimuth := uint16(0)
	azimuth := uint16(0)

	isFinished := false

	for packet := range packets {
		data := packet.Data()
		if len(data) == 1248 {
			if data[1247] == 0x28 {
				decodeVLP32Blocks(&data, &isFinished, &frameCount,
					&totalWorkers, &workerIndex,
					&startFrame, &endFrame,
					&azimuth, &prevAzimuth,
					&nextFramePoints, &currFramePoints, &outputFolder)
			} else if data[1247] == 0x22 {
				fmt.Println("Will decode VLP16")
			}

			if isFinished {
				break
			}

			lidarPackets++
		}
		totalPackets++
	}
	wg.Done()
}

func parsePcap(pcapFile string, outputPath *string, totalWorkers uint8, startFrame int, endFrame int, lidarModel byte) {
	var wg sync.WaitGroup
	wg.Add(int(totalWorkers))
	// fmt.Println("frameCount", "framePointCount", "totalPackets", "azimuth", "prevAzimuth")
	for workerIndex := uint8(0); workerIndex < totalWorkers; workerIndex++ {
		go assignWorker(pcapFile, lidarModel, workerIndex, totalWorkers, true, *outputPath, startFrame, endFrame, &wg)
	}
	wg.Wait()
}

func main() {
	pcapFile := "C:/Users/brendon.dulam/Desktop/Magic Hat/city.pcap"
	outputPath := "V:/JP01/DataLake/Common_Write/CLARITY_OUPUT/Magic_Hat/json/test"
	totalWorkers := uint8(runtime.NumCPU() / 2)
	// totalWorkers := uint8(1)

	startTime := time.Now()
	parsePcap(pcapFile, &outputPath, totalWorkers, 0, -1, 0x28)
	endTime := time.Now()

	fmt.Println(totalWorkers, "Workers Execution Time", endTime.Sub(startTime))
}
