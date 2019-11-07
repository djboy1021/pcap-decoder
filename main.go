package main

import (
	"encoding/json"
	"fmt"
	"lidar/cli"
	"lidar/parsepcap"
	"lidar/path"
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

	// Check if PCAP file exists
	if !path.Exists(pcapFile) {
		panic(pcapFile + " does not exist")
	}

	// Check if Output path exists
	if !path.Exists(outputPath) {
		if mkdirp {
			fmt.Println("Output Path will be generated")
		} else {
			panic(outputPath + " does not exist")
		}
	}

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
	} else {
		channels := parsepcap.GetIP4Channels(&pcapFile)
		parsepcap.ParsePCAP(&pcapFile, &outputPath, totalWorkers, startFrame, endFrame, channels, isSaveAsJSON)
	}
	endTime := time.Now()

	// Display Summary
	printSummary(&pcapFile, &outputPath, &startFrame, &endFrame, &startTime, &endTime, &mkdirp, &isStatsOnly)
}
