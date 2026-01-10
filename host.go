package main

type QuakeParams struct {
	baseDir string
	userDir string

	errState int
}

var HostParams *QuakeParams

var HostInitialized bool
var MultiUser bool
var IsDedicated bool
var FitzMode bool
