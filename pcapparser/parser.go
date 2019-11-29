package pcapparser

import (
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
	lidarSources := make([]LidarSource, 0)

PACKETS:
	for packet := range packets {
		nextPacketData := packet.Data()

		switch len(nextPacketData) {
		case 1248:
			decodePacket(&packet, indexLookup, &lidarSources, &nextPacketData)
		case 554:
			// fmt.Println("GPRMC packet")
		default:
			continue PACKETS
		}

		// if len(lidarSources) > 2 {
		// 	break
		// }
	}
}

func getPackets() (chan gopacket.Packet, error) {
	handle, err := pcap.OpenOffline(cli.UserInput.PcapFile)
	return gopacket.NewPacketSource(handle, handle.LinkType()).Packets(), err
}

func decodePacket(p *gopacket.Packet, indexLookup map[string]uint8, lidarSources *[]LidarSource, nextPacketData *[]byte) {
	ipAddress := lib.GetIPv4((*p).String())

	// Parse packet in advance
	nextPacket, err := NewLidarPacket(nextPacketData)
	check(err)

	if indexLookup[ipAddress] == 0 {
		initialAzimuth := nextPacket.blocks[0].Azimuth
		if initialAzimuth == 0 {
			initialAzimuth = 360
		}

		*lidarSources = append(*lidarSources, LidarSource{
			InitialAzimuth: initialAzimuth})
		indexLookup[ipAddress] = uint8(len(*lidarSources))
	}

	// Simplify the address of the current IP Address
	// Subtract 1 to the indexLookup value to correct the actual number
	lidarSource := &((*lidarSources)[indexLookup[ipAddress]-1])

	// Wait for nonempty timestamp
	if lidarSource.CurrentPacket.TimeStamp > 0 {
		// Set next packet's azimuth
		lidarSource.nextPacketAzimuth = nextPacket.blocks[0].Azimuth
		lidarSource.SetCurrentFrame()

		if len(lidarSource.Buffer) > 0 {

			// lidarSource.CurrentFrame.XYZ(
			// 	RotationAngles{},
			// 	Translation{x: 50, y: 20})

			if len(lidarSource.PreviousFrame.Points) > 0 {
				lidarSource.elevationView()

				// Variables for localization
				limits := [3][2]float64{
					{-20000, 20000},
					{-20000, 20000},
					{-10000, 10000}}
				// lidarSource.LocalizeCurrentFrame(&limits)
				lidarSource.CurrentFrame.visualizeFrame(&limits, 1024)
			}

			lidarSource.PreviousFrame = lidarSource.CurrentFrame
			lidarSource.CurrentFrame.Points = lidarSource.Buffer
			lidarSource.Buffer = nil

			lidarSource.CurrentFrame.Index++

			// if lidarSource.CurrentFrame.Index > 60 {
			// 	panic("Temp stop")
			// }
		}

		// fmt.Println(nextPacket.TimeStamp, lidarSource.CurrentPacket.TimeStamp)
	}

	// Update current packet
	lidarSource.CurrentPacket = nextPacket
}
