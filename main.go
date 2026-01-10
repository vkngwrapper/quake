package main

import (
	"os"
)

func main() {
	//var time, oldTime, newTime float64

	HostParams = &QuakeParams{
		baseDir:  ".",
		errState: 0,
	}

	CmdLine.Init(os.Args)
	IsDedicated = CmdLine.CheckParam("-dedicated") > 0

}
