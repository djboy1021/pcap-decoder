package main

import (
	"clarity/lib"
	"github.com/bldulam1/pcap-decoder/global"
	"github.com/bldulam1/pcap-decoder/pcapdecoder"
	"github.com/urfave/cli"
	"os"
	"pcap-decoder/path"
)

func validatePaths() {
	// Check if PCAP file exists
	if !path.Exists(global.UserInput.PcapFile) {
		panic(global.UserInput.PcapFile + " does not exist")
	}

	// Check if Output path exists
	if !path.Exists(global.UserInput.OutputPath) {
		if global.UserInput.Mkdirp {
			os.MkdirAll(global.UserInput.OutputPath, os.ModePerm)
		} else {
			panic(global.UserInput.OutputPath + " does not exist")
		}
	}
}

func main() {
	app := global.UserInput.CreateApp(runApp)
	lib.DisplayError(app.Run(os.Args))
}

func runApp(c *cli.Context) error {
	validatePaths()
	pcapdecoder.ParsePCAP()
	return nil
}
