package main

const (
	NumPaletteOctreeNodes  int = 184
	NumPaletteOctreeColors int = 5844
)

type PaletteOctreeNode struct {
	ChildOffsets [8]uint32
}
