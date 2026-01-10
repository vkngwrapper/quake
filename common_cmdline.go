package main

import (
	"log"
	"strings"
)

const MaxCmdLineArgs = 50
const MaxCmdLineLength = 256

var CmdLine = &CmdLineArgs{}

type CmdLineArgs struct {
	cmdLine       string
	args          []string
	safeMode      bool
	rogue         bool
	standardQuake bool
	hipnotic      bool
}

func (a *CmdLineArgs) SafeMode() bool {
	return a.safeMode
}

func (a *CmdLineArgs) SetRogue() {
	a.rogue = true
	a.standardQuake = false
}

func (a *CmdLineArgs) SetHipnotic() {
	a.hipnotic = true
	a.standardQuake = false
}

func (a *CmdLineArgs) SetStandardQuake() {
	a.hipnotic = false
	a.rogue = false
	a.standardQuake = true
}

func (a *CmdLineArgs) Rogue() bool {
	return a.rogue
}

func (a *CmdLineArgs) StandardQuake() bool {
	return a.standardQuake
}

func (a *CmdLineArgs) Hipnotic() bool {
	return a.hipnotic
}

func (a *CmdLineArgs) Args() string {
	return a.cmdLine
}

func (a *CmdLineArgs) ArgCount() int {
	return len(a.args)
}

func (a *CmdLineArgs) Arg(index int) string {
	if index < 0 || index >= len(a.args) {
		return ""
	}
	return a.args[index]
}

func (a *CmdLineArgs) CmdLine() string {
	return a.cmdLine
}

func (a *CmdLineArgs) Init(args []string) {
	argCount := len(args)
	if argCount > MaxCmdLineArgs {
		argCount = MaxCmdLineArgs
	}

	var cmdLineStr strings.Builder
	remainingLength := MaxCmdLineLength
	a.args = a.args[:0]
	a.safeMode = false
	a.standardQuake = true

	for i := 0; i < argCount; i++ {
		arg := args[i]
		if len(arg) > remainingLength {
			cmdLineStr.WriteString(arg[:remainingLength])
			a.args = append(a.args, arg[:remainingLength])
			break
		} else {
			cmdLineStr.WriteString(arg)
			a.args = append(a.args, arg)
		}

		if arg == "-safe" {
			a.safeMode = true
		}

		remainingLength -= len(arg)

		if remainingLength > 0 {
			cmdLineStr.WriteRune(' ')
			remainingLength--
		} else {
			break
		}
	}

	a.cmdLine = strings.TrimRight(cmdLineStr.String(), " ")
	log.Printf("Command line: %s\n", a.cmdLine)

	if a.CheckParam("-rogue") > 0 {
		a.SetRogue()
	}

	if a.CheckParam("-hipnotic") > 0 || a.CheckParam("-quoth") > 0 {
		a.SetHipnotic()
	}
}

func (a *CmdLineArgs) CheckParam(param string) int {
	return a.CheckParamNext(0, param)
}

func (a *CmdLineArgs) CheckParamNext(last int, param string) int {
	for i := last + 1; i < len(a.args); i++ {
		if a.args[i] == param {
			return i
		}
	}

	return 0
}
