package global

// UserInput contains the users commandline input
var UserInput = CLInput{
	PcapFile:     "",
	OutputPath:   "",
	StartFrame:   0,
	EndFrame:     0,
	Mkdirp:       false,
	IsSaveAsJSON: false,
	IsSaveAsPNG:  false,
}
