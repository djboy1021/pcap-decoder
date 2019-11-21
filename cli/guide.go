package cli

import (
	"fmt"
	"time"
)

// PrintHelp displays the usage of this application
func PrintHelp() {
	fmt.Println("Usage: ./pcap-decoder.exe <arguments>")
	fmt.Println("")
	fmt.Println("where <arguments> is one of:")
	fmt.Println("    --pcapFile    <string, full path of the pcap file>")
	fmt.Println("    --outputPath  <string, path of the output files>")
	fmt.Println("    --mkdirp      <boolean, true/false>")
	fmt.Println("    --startFrame  <integer, 0 for first frame>")
	fmt.Println("    --endFrame    <integer, -1 for last frame>")

	fmt.Println("    --mkdirp")
	fmt.Println("    --isStatsOnly")
	fmt.Println("    --json")
	fmt.Println("    --png")

	fmt.Println("")
	fmt.Println("Sample:")
	fmt.Println("./pcap-decoder.exe --pcapFile V:/JP01/DataLake/Common_Write/KEEP_Magic_Hat_GT/RTB_DC/External_Shared_Data/city.pcap --outputPath V:/JP01/DataLake/Common_Write/KEEP_Magic_Hat_GT/RTB_DC/External_Shared_Data --mkdirp false --startFrame 0 --endFrame 12 --json true")
	fmt.Println("")
}

// PrintSummary displays the user input summary
func PrintSummary() {
	fmt.Println("     PcapFile:", UserInput.PcapFile)
	fmt.Println("   OutputPath:", UserInput.OutputPath)
	fmt.Println("  OutputFiles:", UserInput.OutputFiles())
	fmt.Println("       Mkdirp:", UserInput.Mkdirp)
	fmt.Println("   StartFrame:", UserInput.StartFrame)
	fmt.Println("     EndFrame:", UserInput.EndFrame)
	fmt.Println("     Channels:", UserInput.Channels)
	fmt.Println("ExecutionTime:", time.Now().Sub(UserInput.StartTime))
}
