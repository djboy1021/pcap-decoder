package pcapparser

import (
	"fmt"
	"pcap-decoder/cli"
	"pcap-decoder/lib"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

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
	lidarSource := &((*addresses)[indexLookup[ipAddress]-1])

	// Wait for nonempty timestamp
	if lidarSource.CurrentPacket.TimeStamp > 0 {
		lidarSource.CurrentPacket.SetPointCloud(
			nextPacket.Blocks[0].Azimuth,
			lidarSource)

		if len(lidarSource.Buffer) > 0 {

			fmt.Println("New Frame", lidarSource.FrameIndex, len(lidarSource.Buffer), len(lidarSource.CurrentFrame))
			// fmt.Println(lidarSource.CurrentFrame[0])
			// fmt.Println(lidarSource.CurrentFrame[0].GetXYZ())
			lidarSource.CurrentFrame = lidarSource.Buffer
			lidarSource.Buffer = nil
			// panic("End")
		}

		// fmt.Println(nextPacket.TimeStamp, lidarSource.CurrentPacket.TimeStamp)
	}

	// Update current packet
	lidarSource.CurrentPacket = nextPacket
}
