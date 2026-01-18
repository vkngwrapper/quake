package main

import (
	"github.com/vkngwrapper/arsenal/vam"
	"github.com/vkngwrapper/core/v3/core1_0"
)

var D8To24Table [256]uint32
var D8To24TableFullBright [256]uint32
var D8To24TableFullBrightFence [256]uint32
var D8To24TableNoBright [256]uint32
var D8To24TableNoBrightFence [256]uint32
var D8To24TableConChars [256]uint32

type TextureFlags int

const (
	TexPrefNone TextureFlags = 1 << iota
	TexPrefMipMap
	// TexPrefLinear and TexPrefNearest aren't supposed to be ORd with TexPrefMipMap
	TexPrefLinear
	TexPrefNearest
	TexPrefAlpha
	TexPrefPad
	TexPrefPersist
	TexPrefOverwrite
	TexPrefNoPicMip
	TexPrefFullBright
	TexPrefNoBright
	TexPrefConChars
	TexPrefWarpImage
	TexPrefPreMultiply
	TexPrefAlphaPixels
)

type SourceFormat int

const (
	SourceIndexed SourceFormat = iota
	SourceLightmap
	SourceRGBA
	SourceSurfIndices
	SourceRGBACubeMap
	SourceIndexedPalette
)

type GLTexture struct {
	Next *GLTexture

	Owner *QModel

	// managed by image loading
	Name         string
	PathId       int // directory owner came from if owner != nil, otherwise 0
	Width        uint
	Height       uint
	Flags        TextureFlags
	SourceFile   string
	SourceOffset int
	SourceFormat SourceFormat
	SourceWidth  uint
	SourceHeight uint
	SourceCRC    uint16
	Shirt        int8 // 0-13 shirt color, or -1 if never colormapped
	Pants        int8 // 0-13 pants color, or -1 if never colormapped

	// Used for rendering
	Image                core1_0.Image
	ImageView            core1_0.ImageView
	TargetImageView      core1_0.ImageView
	Allocation           vam.Allocation
	DescriptorSet        core1_0.DescriptorSet
	FrameBuffer          core1_0.Framebuffer
	StorageDescriptorSet core1_0.DescriptorSet
}

type TextureManager struct {
}

var TexMgr = &TextureManager{}
