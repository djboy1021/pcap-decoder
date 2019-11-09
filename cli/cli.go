package cli

import (
	"os"
	"strconv"
	"strings"
)

// CLInput contains the command line inputs of the user
type CLInput struct {
	pcapFile     string
	outputPath   string
	startFrame   int
	endFrame     int
	mkdirp       bool
	isStatsOnly  bool
	isSaveAsJSON bool
	isSaveAsPNG  bool
}

// UserInput is the global variable containing the user inputs
var UserInput = CLInput{endFrame: -1}

// SetUserInput gets the command line input
func SetUserInput() {
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
				UserInput.pcapFile = value
			case "outputPath":
				UserInput.outputPath = value
			case "startFrame":
				UserInput.startFrame, err = strconv.Atoi(value)
			case "endFrame":
				UserInput.endFrame, err = strconv.Atoi(value)
			case "mkdirp":
				setArgValue(&(UserInput.mkdirp), key, value)
			case "statsOnly":
				setArgValue(&(UserInput.isStatsOnly), key, value)
			case "json":
				setArgValue(&(UserInput.isSaveAsJSON), key, value)
			case "png":
				setArgValue(&(UserInput.isSaveAsPNG), key, value)
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
}

func setArgValue(parameter *bool, key string, value string) {
	if len(value) == 0 {
		*parameter = true
	} else {
		*parameter, _ = strconv.ParseBool(value)
	}
}
