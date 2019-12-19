package pcapdecoder

import (
	"clarity/lib"
	"fmt"
	"github.com/bldulam1/pcap-decoder/global"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"pcap-decoder/calibration"
	"strings"
)

var lidarSources = make(map[string]LidarSource)

// ParsePCAP creates several go routines to start decoding the PCAP file.
func ParsePCAP() {
	// get packets handler
	handle, err := pcap.OpenOffline(global.UserInput.PcapFile)
	lib.DisplayError(err)
	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()

PACKETS:
	for packet := range packets {
		nextPacketData := packet.Data()

		switch len(nextPacketData) {
		case 1248:
			address := getIPv4(packet.String())
			decodeLidarPacket(address, &nextPacketData)

			//decodeLidarPacket(&packet, indexLookup, &lidarSources, &nextPacketData)
		case 554:
			// fmt.Println("GPRMC packet")
		case 1358:
			// decodeCameraPacket(&packet)
		default:
			continue PACKETS
		}

	}
}

// getIPv4 returns the IP address of the lidar packet
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

func decodeLidarPacket(address string, nextPacketData *[]byte) {
	lidarSource := lidarSources[address]

	// Parse packet in advance
	nextPacket, err := NewLidarPacket(nextPacketData)
	check(err)

	if len(lidarSource.Address) == 0 {
		lidarSource = LidarSource{
			Address:        address,
			InitialAzimuth: nextPacket.Blocks[0].Azimuth,
			Calibration:    calibration.Lidars[address],
		}
		fmt.Println(address, len(*nextPacketData))
	}

	// Wait for nonempty timestamp
	if lidarSource.CurrentPacket.TimeStamp > 0 {
		// Set next packet's azimuth
		lidarSource.NextPacketAzimuth = nextPacket.Blocks[0].Azimuth
		prevFrameIndex := lidarSource.CurrentFrame.Index
		lidarSource.SetCurrentFrame(lidarSource.CurrentFrame.Index)

		if prevFrameIndex < lidarSource.CurrentFrame.Index {

			if len(lidarSource.PreviousFrame.Points) > 0 {
				fmt.Println(lidarSource.CurrentFrame.Index, address)
			}

			lidarSource.PreviousFrame = lidarSource.CurrentFrame
			lidarSource.CurrentFrame.Points = lidarSource.Buffer
			lidarSource.Buffer = nil
		}

	}

	// Update current packet
	lidarSource.CurrentPacket = nextPacket
	lidarSources[address] = lidarSource
}
