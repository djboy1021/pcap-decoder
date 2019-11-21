package pcapparser

import (
	"fmt"
	"pcap-decoder/cli"
	"pcap-decoder/lib"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

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
			ipAddress := lib.GetIPv4(packet.String())
			startTime := time.Now()
			currentLP, err := NewLidarPacket(&nextPacketData)
			check(err)

			fmt.Println(ipAddress, currentLP.ProductID, time.Now().Sub(startTime))
		}
	}
}

func getPackets() (chan gopacket.Packet, error) {
	handle, err := pcap.OpenOffline(cli.UserInput.PcapFile)
	return gopacket.NewPacketSource(handle, handle.LinkType()).Packets(), err
}
