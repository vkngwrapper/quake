package main

type VideoRect struct {
	X, Y, Width, Height int

	Next *VideoRect
}

type ModeState int

const (
	ModeStateUninit ModeState = iota
	ModeStateWindowed
	ModeStateFullscreen
)

type VideoDef struct {
	Buffer           []byte // invisible buffer
	ColorMap         []byte
	ColorMap16       []uint16
	FullBright       int // index of first fullbright color
	RowBytes         int // Maybe >width if displayed in a window
	Width            int
	Height           int
	Aspect           float32 // width / height -- <0 is taller than wide
	RecalcRefDef     bool    // If true, recalc vid-based stuff
	ConBuffer        []byte
	ConRowBytes      int
	ConWidth         int
	ConHeight        int
	RestartNextFrame bool
}

var VideoData = &VideoDef{}
