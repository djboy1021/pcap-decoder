package pcapparser

import (
	"fmt"
	"pcap-decoder/cli"
	"pcap-decoder/lib"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// LidarSource contains the iteration info of an IP address
type LidarSource struct {
	FrameIndex     uint
	InitialAzimuth uint16
	CurrentPacket  LidarPacket
	CurrentFrame   []LidarPoint
	Buffer         []LidarPoint
}

// ParsePCAP creates several go routines to start decoding the PCAP file.
func ParsePCAP() {
	// get packets
	packets, err := getPackets()
	check(err)

	indexLookup := make(map[string]uint8, 0)
	addresses := make([]LidarSource, 0)

PACKETS:
	for packet := range packets {
		nextPacketData := packet.Data()

		switch len(nextPacketData) {
		case 1248:

			decodePacket(&packet, indexLookup, &addresses, &nextPacketData)

		case 554:
			// fmt.Println("GPRMC packet")
		default:
			continue PACKETS
		}

		// if len(addresses) > 2 {
		// 	break
		// }
	}
}

func getPackets() (chan gopacket.Packet, error) {
	handle, err := pcap.OpenOffline(cli.UserInput.PcapFile)
	return gopacket.NewPacketSource(handle, handle.LinkType()).Packets(), err
}

func decodePacket(p *gopacket.Packet, indexLookup map[string]uint8, addresses *[]LidarSource, nextPacketData *[]byte) {
	ipAddress := lib.GetIPv4((*p).String())

	// Parse packet in advance
	nextPacket, err := NewLidarPacket(nextPacketData)
	check(err)

	if indexLookup[ipAddress] == 0 {
		initialAzimuth := nextPacket.Blocks[0].Azimuth
		if initialAzimuth == 0 {
			initialAzimuth = 360
		}

		*addresses = append(*addresses, LidarSource{
			FrameIndex:     0,
			InitialAzimuth: initialAzimuth})
		indexLookup[ipAddress] = uint8(len(*addresses))
	}

	// Simplify the address of the current IP Address
	// Subtract 1 to the indexLookup value to correct the actual number
	channel := &((*addresses)[indexLookup[ipAddress]-1])

	// Wait for nonempty timestamp
	if channel.CurrentPacket.TimeStamp > 0 {
		channel.CurrentPacket.SetPointCloud(
			nextPacket.Blocks[0].Azimuth,
			channel)

		if len(channel.Buffer) > 0 {

			fmt.Println("New Frame", channel.FrameIndex, len(channel.Buffer), len(channel.CurrentFrame))
			// fmt.Println(channel.CurrentFrame[0])
			// fmt.Println(channel.CurrentFrame[0].GetXYZ())
			channel.CurrentFrame = channel.Buffer
			channel.Buffer = nil
			// panic("End")
		}

		// fmt.Println(nextPacket.TimeStamp, channel.CurrentPacket.TimeStamp)
	}

	// Update current packet
	channel.CurrentPacket = nextPacket
}
