package pcapparser

import (
	"pcap-decoder/cli"
	"pcap-decoder/lib"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// ChannelInfo contains the iteration info of an IP address
type ChannelInfo struct {
	frameIndex     int
	InitialAzimuth uint16
	CurrentPacket  LidarPacket
}

// ParsePCAP creates several go routines to start decoding the PCAP file.
func ParsePCAP() {
	// get packets
	packets, err := getPackets()
	check(err)

	indexLookup := make(map[string]uint8, 0)
	addresses := make([]ChannelInfo, 0)

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

func decodePacket(p *gopacket.Packet, indexLookup map[string]uint8, addresses *[]ChannelInfo, nextPacketData *[]byte) {
	ipAddress := lib.GetIPv4((*p).String())

	if indexLookup[ipAddress] == 0 {
		*addresses = append(*addresses, ChannelInfo{})
		indexLookup[ipAddress] = uint8(len(*addresses))
	}

	// Simplify the address of the current IP Address
	// Subtract 1 to the indexLookup value to correct the actual number
	ipadd := &((*addresses)[indexLookup[ipAddress]-1])

	// Parse packet in advance
	nextPacket, err := NewLidarPacket(nextPacketData)
	check(err)

	// Wait for nonempty timestamp
	if ipadd.CurrentPacket.TimeStamp > 0 {
		ipadd.CurrentPacket.GetPointCloud(nextPacket.Blocks[0].Azimuth)

		// fmt.Println(nextPacket.TimeStamp, ipadd.CurrentPacket.TimeStamp)
	}

	// Update current packet
	ipadd.CurrentPacket = nextPacket
}
