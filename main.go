package main

import (
	"encoding/json"
	"fmt"
	"pcap-decoder/cli"
	"pcap-decoder/parsepcap"
	"pcap-decoder/path"
)

func validatePaths() {
	// Check if PCAP file exists
	if !path.Exists(cli.UserInput.PcapFile) {
		panic(cli.UserInput.PcapFile + " does not exist")
	}

	// Check if Output path exists
	if !path.Exists(cli.UserInput.OutputPath) {
		if cli.UserInput.Mkdirp {
			fmt.Println("Output Path will be generated")
		} else {
			panic(cli.UserInput.OutputPath + " does not exist")
		}
	}
}

func main() {
	// Set Commandline arguments global variables
	cli.SetUserInput()
	// fmt.Println(cli.UserInput)

	// Start decoding the PCAP file
	if cli.UserInput.IsStatsOnly {
		pcapStats := parsepcap.GetStats()
		jsonBin, _ := json.Marshal(pcapStats)
		fmt.Println("Stats:", string(jsonBin))
	} else if len(cli.UserInput.PcapFile) > 0 && len(cli.UserInput.OutputPath) > 0 {
		// Check if paths are valid
		validatePaths()

		// Check available IP addresses of the packets
		if cli.UserInput.Channels == nil {
			cli.UserInput.Channels = parsepcap.GetIP4Channels()
		}

		parsepcap.ParsePCAP()

		// Display Summary
		cli.PrintSummary()
	} else {
		cli.PrintHelp()
	}
}
