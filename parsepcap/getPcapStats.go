package parsepcap

import (
	"pcap-decoder/dictionary"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// LidarStats contains the statistics of a particular Lidar model
type LidarStats struct {
	ID           byte   `json:"id"`
	Model        string `json:"model"`
	PacketsCount int    `json:"packetsCount"`
}

// PCAPStats contains the statistics of a PCAP file
type PCAPStats struct {
	TotalPackets    int          `json:"totalPackets"`
	PositionPackets int          `json:"positionPackets"`
	Lidars          []LidarStats `json:"lidars"`
}

// GetStats provides the number of packets, lidarpackets, and position packets in a PCAP file
func GetStats(pcapFile *string) PCAPStats {
	var pcapStats PCAPStats

	handle, err := pcap.OpenOffline(*pcapFile)
	if err != nil {
		panic(err)
	}

	models := make(map[byte]int)

	for packet := range gopacket.NewPacketSource(handle, handle.LinkType()).Packets() {
		data := packet.Data()

		if len(data) == 1248 {
			models[data[1247]]++
		} else if len(data) == 554 {
			pcapStats.PositionPackets++
		}
		pcapStats.TotalPackets++
	}

	for id, count := range models {
		lidar := LidarStats{id, dictionary.GetProductID(id), count}
		pcapStats.Lidars = append(pcapStats.Lidars, lidar)
	}

	return pcapStats
}
