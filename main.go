package main

import (
	"encoding/json"
	"fmt"
	"os"
	"pcap-decoder/cli"
	"pcap-decoder/lib"
	"pcap-decoder/parsepcap"
	"pcap-decoder/path"
	"pcap-decoder/pcapparser"
)

func validatePaths() {
	// Check if PCAP file exists
	if !path.Exists(cli.UserInput.PcapFile) {
		panic(cli.UserInput.PcapFile + " does not exist")
	}

	// Check if Output path exists
	if !path.Exists(cli.UserInput.OutputPath) {
		if cli.UserInput.Mkdirp {
			os.MkdirAll(cli.UserInput.OutputPath, os.ModePerm)
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
			cli.UserInput.Channels = lib.GetIP4Channels()
		}

		pcapparser.ParsePCAP()

		// parsepcap.ParsePCAP()

		// Display Summary
		cli.PrintSummary()
	} else {
		cli.PrintHelp()
	}
}
