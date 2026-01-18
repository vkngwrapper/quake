package main

const (
	MaxLightMaps int = 4
	MaxMapHulls  int = 4
	MipLevels    int = 4
	NumAmbients  int = 4
)

type DModel struct {
	Mins      [3]float32
	Maxs      [3]float32
	HeadNode  [MaxMapHulls]int
	VisLeafs  int // not including the solid leaf 0
	FirstFace int
	NumFaces  int
}

const (
	PlaneX    int = 0
	PlaneY    int = 1
	PlaneZ    int = 2
	PlaneAnyX int = 3
	PlaneAnyY int = 4
	PlaneAnyZ int = 5
)
