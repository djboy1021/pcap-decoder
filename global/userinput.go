package global

import (
	"github.com/urfave/cli"
)

type CLInput struct {
	PcapFile     string
	OutputPath   string
	StartFrame   int
	EndFrame     int
	Channels     cli.StringSlice
	Mkdirp       bool
	IsSaveAsJSON bool
	IsSaveAsPNG  bool
}

// CreateApp returns a CLI app
func (ui *CLInput) CreateApp(action func(c *cli.Context) error) *cli.App {
	return &cli.App{
		Name:                 "PCAP Decoder",
		Usage:                "Decodes the lidar and camera information of a PCAP file",
		EnableBashCompletion: true,
		Action:               action,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "pcapFile",
				Aliases:     []string{"p"},
				Value:       ui.PcapFile,
				Usage:       "Full path of the PCAP file",
				Destination: &(ui.PcapFile),
			},
			&cli.StringFlag{
				Name:        "outputPath",
				Aliases:     []string{"o"},
				Value:       ui.OutputPath,
				Usage:       "location of the output files",
				Destination: &(ui.OutputPath),
			},
			&cli.IntFlag{
				Name:        "startFrame",
				Aliases:     []string{"s"},
				Value:       ui.StartFrame,
				Usage:       "index of the beginning frame",
				Destination: &(ui.StartFrame),
			},
			&cli.IntFlag{
				Name:        "endFrame",
				Aliases:     []string{"e"},
				Value:       ui.EndFrame,
				Usage:       "index of the end frame",
				Destination: &(ui.EndFrame),
			},
			&cli.BoolFlag{
				Name:        "mkdirp",
				Aliases:     []string{"m"},
				Usage:       "recursively creates the output folder ",
				Required:    false,
				Hidden:      false,
				Value:       ui.Mkdirp,
				Destination: &(ui.Mkdirp),
			},
			&cli.BoolFlag{
				Name:        "JSON",
				Aliases:     []string{"j"},
				Usage:       "Save pointcloud in JSON format",
				Required:    false,
				Hidden:      false,
				Value:       ui.IsSaveAsJSON,
				Destination: &(ui.IsSaveAsJSON),
			},
			&cli.BoolFlag{
				Name:        "PNG",
				Aliases:     []string{"png"},
				Usage:       "Save pointcloud's bird's eye view in PNG format",
				Hidden:      false,
				Value:       ui.IsSaveAsJSON,
				Destination: &(ui.IsSaveAsJSON),
			},
			&cli.StringSliceFlag{
				Name:     "channels",
				Aliases:  []string{"c"},
				Usage:    "IP addresses to whitelist",
				Required: true,
				Hidden:   false,
				Value:    &(ui.Channels),
			},
		},
	}
}
