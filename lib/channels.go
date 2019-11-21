package lib

import (
	"pcap-decoder/cli"
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// IPAddress is an IPV4 address
type IPAddress struct {
	ByteAddress []uint8
}

// NewIPv4Address creates a new IPv4 address based from the string input
func NewIPv4Address(address string) IPAddress {
	ipadd := make([]uint8, 4)

	for i, add := range strings.Split(address, ".") {
		val, _ := strconv.ParseUint(add, 10, 8)

		ipadd[i] = uint8(val)
	}

	return IPAddress{ByteAddress: ipadd}
}

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
