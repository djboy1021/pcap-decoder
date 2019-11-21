package lib

import (
	"pcap-decoder/cli"
	"strings"

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
		ip4 := GetIPv4(packet.String())

		channelMap[ip4]++

		if channelMap[ip4] > 4 {
			break
		} else if channelMap[ip4] == 1 {
			channels = append(channels, ip4)
		}
	}

	return channels
}

// GetIPv4 returns the IP address of the lidar packet
func GetIPv4(pcapString string) string {
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
