package parsepcap

import (
	"pcap-decoder/cli"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// GetIP4Channels return unique IPV4 Addresses from a PCAP file
func GetIP4Channels() []string {
	channels := make([]string, 0)
	channelMap := make(map[string]uint8)

	handle, err := pcap.OpenOffline(cli.UserInput.PcapFile)
	if err != nil {
		panic(err)
	}
	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()

	for packet := range packets {
		if len(packet.Data()) != 1248 {
			continue
		}
		ip4 := getIPv4(packet.String())

		channelMap[ip4]++

		if channelMap[ip4] > 4 {
			break
		} else if channelMap[ip4] == 1 {
			channels = append(channels, ip4)
		}
	}

	return channels
}
