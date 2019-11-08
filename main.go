package main

import (
	"encoding/json"
	"fmt"
	"pcap-decoder/cli"
	"pcap-decoder/parsepcap"
	"pcap-decoder/path"
	"runtime"
	"time"
)

func printSummary(pcapFile *string, outputPath *string, startFrame *int, endFrame *int, startTime *time.Time, endTime *time.Time, mkdirp *bool, isStatsOnly *bool) {
	fmt.Println("SUMMARY")
	fmt.Println("pcapFile:", *pcapFile)
	fmt.Println("isStatsOnly:", *isStatsOnly)
	fmt.Println("outputPath:", *outputPath)
	fmt.Println("mkdirp:", *mkdirp)
	fmt.Println("startFrame:", *startFrame)
	fmt.Println("endFrame:", *endFrame)
	fmt.Println("Execution Time:", (*endTime).Sub(*startTime))
}

func validatePaths(pcapFile *string, outputPath *string, mkdirp bool) {
	// Check if PCAP file exists
	if !path.Exists(*pcapFile) {
		panic(*pcapFile + " does not exist")
	}

	// Check if Output path exists
	if !path.Exists(*outputPath) {
		if mkdirp {
			fmt.Println("Output Path will be generated")
		} else {
			panic(*outputPath + " does not exist")
		}
	}

}

func printHelp() {
	fmt.Println("Usage: ./pcap-decoder.exe <arguments>")
	fmt.Println("")
	fmt.Println("where <arguments> is one of:")
	fmt.Println("    --pcapFile    <string, full path of the pcap file>")
	fmt.Println("    --outputPath  <string, path of the output files>")
	fmt.Println("    --mkdirp      <boolean, true/false>")
	fmt.Println("    --startFrame  <integer, 0 for first frame>")
	fmt.Println("    --endFrame    <integer, -1 for last frame>")
	fmt.Println("    --mkdirp      <integer>")
	fmt.Println("    --isStatsOnly <boolean>")
	fmt.Println("    --json        <boolean>")
	fmt.Println("")
	fmt.Println("Sample:")
	fmt.Println("./pcap-decoder.exe --pcapFile V:/JP01/DataLake/Common_Write/KEEP_Magic_Hat_GT/RTB_DC/External_Shared_Data/city.pcap --outputPath V:/JP01/DataLake/Common_Write/KEEP_Magic_Hat_GT/RTB_DC/External_Shared_Data --mkdirp false --startFrame 0 --endFrame 12 --json true")
	fmt.Println("")
}

func main() {
	// Default values
	pcapFile := ""
	outputPath := ""
	startFrame := 0
	endFrame := -1
	mkdirp := false
	isStatsOnly := false
	isSaveAsJSON := false

	// Get Commandline arguments
	cli.GetArgs(&pcapFile, &outputPath, &startFrame, &endFrame, &mkdirp, &isStatsOnly, &isSaveAsJSON)

	// Allocate number of workers
	totalWorkers := uint8(runtime.NumCPU() / 2)
	if endFrame < 0 || int(totalWorkers) < (endFrame-startFrame) { //if all frames
		totalWorkers = uint8(runtime.NumCPU() / 2)
	} else if endFrame > startFrame { //if number of frames is less than cpu cores
		totalWorkers = uint8(endFrame - startFrame)
	} else { // if only one frame
		totalWorkers = uint8(1)
	}

	// Start decoding the PCAP file
	startTime := time.Now()
	if isStatsOnly {
		pcapStats := parsepcap.GetStats(&pcapFile)
		jsonBin, _ := json.Marshal(pcapStats)
		fmt.Println("Stats:", string(jsonBin))
	} else if len(pcapFile) > 0 && len(outputPath) > 0 && isSaveAsJSON {
		// Check if paths are valid
		validatePaths(&pcapFile, &outputPath, mkdirp)

		// Check IP addresses of the packets
		channels := parsepcap.GetIP4Channels(&pcapFile)
		parsepcap.ParsePCAP(&pcapFile, &outputPath, totalWorkers, startFrame, endFrame, channels, isSaveAsJSON)

		// Display Summary
		endTime := time.Now()
		printSummary(&pcapFile, &outputPath, &startFrame, &endFrame, &startTime, &endTime, &mkdirp, &isStatsOnly)
	} else {
		printHelp()
	}
}
