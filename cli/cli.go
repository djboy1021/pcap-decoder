package cli

import (
	"os"
	"strconv"
	"strings"
)

// GetArgs gets the command line input
func GetArgs(pcapFile *string, outputPath *string, startFrame *int, endFrame *int, mkdirp *bool, isStatsOnly *bool) {
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
			key := splitStr[0]
			value := ""
			if len(splitStr) > 0 {
				value = strings.Join(splitStr[1:], " ")
				value = strings.Replace(value, "'", "", -1)
				value = strings.Replace(value, "\"", "", -1)
			}

			if strings.Contains(key, "pcapFile") {
				*pcapFile = value
			} else if strings.Contains(key, "outputPath") {
				*outputPath = value
			} else if strings.Contains(key, "startFrame") {
				*startFrame, err = strconv.Atoi(value)
			} else if strings.Contains(key, "endFrame") {
				*endFrame, err = strconv.Atoi(value)
			} else if strings.Contains(key, "mkdirp") {
				*mkdirp, err = strconv.ParseBool(value)
			} else if strings.Contains(key, "statsOnly") {
				if len(value) == 0 {
					*isStatsOnly = true
				} else {
					*isStatsOnly, _ = strconv.ParseBool(value)
				}
			} else {
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
