package parsepcap

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"pcap-decoder/cli"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

/*
LidarPacket
	Time
	IsDualMode
	ProductID
	Blocks[12]
		Azimuth
		Channels[32]
			Distance
			Reflectivity
*/

// LidarPacket is the raw decoded info of a lidar packet
type LidarPacket struct {
	TimeStamp  uint32       `json:"timestamp"`
	ProductID  byte         `json:"productID"`
	IsDualMode bool         `json:"isDualMode"`
	Blocks     []LidarBlock `json:"blocks"`
}

// LidarBlock contains the channels data
type LidarBlock struct {
	Azimuth  uint16         `json:"azimuth"`
	Channels []LidarChannel `json:"channels"`
}

// LidarChannel contains the basic point data
type LidarChannel struct {
	Distance     uint16 `json:"distance"`
	Reflectivity uint8  `json:"reflectivity"`
}

// NewLidarPacket creates a new LidarPacket Object
func NewLidarPacket(data *[]byte) (LidarPacket, error) {
	var lp LidarPacket
	var err error

	if len(*data) == 1248 {
		blocks := make([]LidarBlock, 12)
		setBlocks(data, &blocks)

		lp = LidarPacket{
			IsDualMode: isDualMode(data),
			ProductID:  getProductID(data),
			TimeStamp:  getTime(data),
			Blocks:     blocks}
	} else {
		err = fmt.Errorf("not a lidar packet")
	}

	return lp, err
}

func setChannel(data *[]byte, channels *[]LidarChannel, chIndex uint8, blkIndex uint16, wg *sync.WaitGroup) {
	(*channels)[chIndex] = LidarChannel{
		Distance:     uint16((*data)[blkIndex+1])<<8 + uint16((*data)[blkIndex]),
		Reflectivity: (*data)[blkIndex+2]}

	wg.Done()
}

func setChannels(data *[]byte, index uint16, channels *[]LidarChannel, blocksWG *sync.WaitGroup) {
	var channelsWaitGroup sync.WaitGroup
	chIndex := uint8(0)
	channelsWaitGroup.Add(32)
	for blkIndex := index + 4; blkIndex < index+100; blkIndex += 3 {
		setChannel(data, channels, chIndex, blkIndex, &channelsWaitGroup)
		chIndex++
	}
	channelsWaitGroup.Wait()

	blocksWG.Done()
}

func setBlocks(data *[]byte, blocks *[]LidarBlock) {
	blkIndex := uint8(0)
	var blocksWG sync.WaitGroup
	blocksWG.Add(12)
	for index := uint16(42); index < 1242; index += 100 {
		// Set Azimuth
		(*blocks)[blkIndex].Azimuth = binary.LittleEndian.Uint16((*data)[index+2 : index+4])

		// Initialize channels
		(*blocks)[blkIndex].Channels = make([]LidarChannel, 32)

		setChannels(data, index, &((*blocks)[blkIndex].Channels), &blocksWG)
		blkIndex++
	}
	blocksWG.Wait()
}

// ParsePCAP creates several go routines to start decoding the PCAP file.
func ParsePCAP() {
	packets, err := getPackets()
	check(err)

	// lidarIndex := make(map[string]uint8)
	// lidarSlices := make([]iterationInfo, 0)

	for packet := range packets {
		nextPacketData := packet.Data()

		switch len(nextPacketData) {
		case 1248:
			// ipAddress := getIPv4(packet.String())
			startTime := time.Now()
			currentSet, err := NewLidarPacket(&nextPacketData)
			check(err)

			jsonObject, _ := json.Marshal(currentSet)

			fmt.Println(string(jsonObject), time.Now().Sub(startTime))

			// // load location of current lidar iteration info
			// if lidarIndex[ipAddress] == 0 {
			// 	lidarSlices = append(lidarSlices, iterationInfo{})
			// 	lidarIndex[ipAddress] = uint8(len(lidarSlices))
			// }
			// lidar := &(lidarSlices[lidarIndex[ipAddress]-1])

			// if len(lidar.currPacketData) == 1248 {
			// 	lp, err := NewLidarPacket(&(lidar.currPacketData))
			// 	check(err)
			// 	fmt.Println(lp, ipAddress)
			// }

			// // Update current packet data
			// lidar.currPacketData = nextPacketData
		}

	}
}

func getPackets() (chan gopacket.Packet, error) {
	handle, err := pcap.OpenOffline(cli.UserInput.PcapFile)
	return gopacket.NewPacketSource(handle, handle.LinkType()).Packets(), err
}
