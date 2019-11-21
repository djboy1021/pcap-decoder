package cli

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// CLInput contains the command line inputs of the user
type CLInput struct {
	PcapFile     string
	OutputPath   string
	StartFrame   int
	EndFrame     int
	Mkdirp       bool
	IsStatsOnly  bool
	IsSaveAsJSON bool
	IsSaveAsPNG  bool
	TotalWorkers uint8
	StartTime    time.Time
	Channels     []string
}

// UserInput is the global variable containing the user inputs
var UserInput = CLInput{EndFrame: -1}

// SetUserInput gets the command line input
func SetUserInput() {
	// Set start time
	UserInput.StartTime = time.Now()

	correctedArgs := make([]string, 0)

	// resolves any strings that are cut due to some space
	for i := 1; i < len(os.Args); i++ {
		if strings.Contains(os.Args[i], "--") {
			correctedArgs = append(correctedArgs, os.Args[i])
		} else {
			correctedArgs[len(correctedArgs)-1] += " " + os.Args[i]
		}
	}

	// resolves any strings that are cut due to some space
	var err error
	for _, arg := range correctedArgs {
		if strings.Contains(arg, "--") {
			splitStr := strings.Split(arg, " ")
			key := strings.Replace(splitStr[0], "--", "", 1)
			value := ""
			if len(splitStr) > 0 {
				value = strings.Join(splitStr[1:], " ")
				value = strings.Replace(value, "'", "", -1)
				value = strings.Replace(value, "\"", "", -1)
			}

			switch key {
			case "pcapFile":
				UserInput.PcapFile = value
			case "outputPath":
				UserInput.OutputPath = value
			case "startFrame":
				UserInput.StartFrame, err = strconv.Atoi(value)
			case "endFrame":
				UserInput.EndFrame, err = strconv.Atoi(value)
			case "mkdirp":
				setArgValue(&(UserInput.Mkdirp), key, value)
			case "statsOnly":
				setArgValue(&(UserInput.IsStatsOnly), key, value)
			case "json":
				setArgValue(&(UserInput.IsSaveAsJSON), key, value)
			case "png":
				setArgValue(&(UserInput.IsSaveAsPNG), key, value)
			default:
				panic(key + " is not recognized as a valid input")
			}

			if err != nil {
				panic(err)
			}
		} else {
			panic("Syntax Error")
		}
	}

	// Set Total Workers
	setTotalWorkers()

	// Set Channels
	// setChannels()
}

func setTotalWorkers() {
	// Allocate number of workers
	totalWorkers := uint8(runtime.NumCPU() / 2)
	startFrame := UserInput.StartFrame
	endFrame := UserInput.EndFrame
	if endFrame < 0 || int(totalWorkers) < (endFrame-startFrame) { //if all frames
		totalWorkers = uint8(runtime.NumCPU() / 2)
	} else if endFrame > startFrame { //if number of frames is less than cpu cores
		totalWorkers = uint8(endFrame - startFrame)
	} else { // if only one frame
		totalWorkers = uint8(1)
	}

	UserInput.TotalWorkers = totalWorkers
}

func setArgValue(parameter *bool, key string, value string) {
	if len(value) == 0 {
		*parameter = true
	} else {
		*parameter, _ = strconv.ParseBool(value)
	}
}

// OutputFiles returns the types of the output files
func (cli CLInput) OutputFiles() []string {
	var outputFiles []string
	if cli.IsSaveAsJSON && !cli.IsStatsOnly {
		outputFiles = append(outputFiles, "json")
	}
	if cli.IsSaveAsPNG && !cli.IsStatsOnly {
		outputFiles = append(outputFiles, "png")
	}

	return outputFiles
}
