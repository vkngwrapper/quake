package main

import (
	"sync/atomic"

	"github.com/vkngwrapper/core/v3/core1_0"
	"github.com/vkngwrapper/extensions/v3/khr_acceleration_structure"
	"github.com/vkngwrapper/math"
)

const (
	VertexSize       int = 7
	MaxDynamicLights int = 64
)

type ModelType int

const (
	ModelBrush ModelType = iota
	ModelSprite
	ModelAlias
)

type SyncType int

const (
	SyncTypeSync SyncType = iota
	SyncTypeRand
	SyncTypeFrameTime
)

type TextureChain int

const (
	TexChainWorld TextureChain = iota
	TexChainModel0
	TexChainModel1
	TexChainModel2
	TexChainModel3
	TexChainModel4
	TexChainModel5
	TexChainAlphaModelAcrossWater
	TexChainAlphaModel
	TexChainNum
)

type PoseVertType int

const (
	PoseVertTypeQuake1 PoseVertType = iota
	PoseVertTypeMD5
	PoseVertTypeQuake3
	PoseVertTypeSize
)

type AABBStructureOfArrays [2 * 3 * 8]float32
type PlaneStructureOfArrays [4 * 8]float32

type MPlane struct {
	Normal   math.Vec3[float32]
	Dist     float32
	Type     byte // For texture axis selection and fast side tests
	SignBits byte // SinX + SignY<<1 + SinZ<<1
	Pad      [2]byte
}

type GLPoly struct {
	Next *GLPoly

	NumVerts int
	Verts    [4][VertexSize]float32
}

type Texture struct {
	Name           string
	Width          uint
	Height         uint
	Shift          uint
	SourceFile     string // Relative filepath
	SourceOffset   int    // Offset from start of BSP file for BSP textures
	GLTexture      *GLTexture
	FullBright     *GLTexture    // Fullbright mask
	WarpImage      *GLTexture    // water animation
	UpdateWarp     atomic.Uint32 // should update warp this frame
	TextureChains  [TexChainNum]*MSurface
	ChainSize      [TexChainNum]int
	AnimTotal      int // Total tenths in a sequence ( 0 = no )
	AnimMin        int // Time for this frame is between min and max
	AnimMax        int
	AnimNext       *Texture // Next in the animation sequence
	AlternateAnims *Texture // bmodels in frame 1 use these
	Offsets        [MipLevels]uint
	Palette        bool
}

type TexInfo struct {
	Vecs     [2][4]float32
	Texture  *Texture
	Flags    int
	TexIndex int
}

type MSurface struct {
	VisFrame int // Should be drawn when node is cross

	Plane *MPlane
	Flags int

	FirstEdge int // Lookup in model->surfedges[], negative numbers are backwards edges
	NumEdges  int

	TextureMins [2]int16
	Extents     [2]int16

	LightS int // GL lightmap coordinates
	LightT int

	Polys         *GLPoly
	TextureChains [TexChainNum]*MSurface

	TexInfo       *TexInfo
	IndirectIndex int

	VBOFirstVert int

	// lighting info
	DynamicLightFrame int
	DynamicLightBits  [(MaxDynamicLights + 31) >> 5]uint
	// int is 32 bits, need an array for MaxDynamicLights > 32

	LightMapTextureNum int
	Styles             [MaxLightMaps]byte
	StylesBitmap       uint32            // bitmap of styles used (16..64 OR-folded into bits 16..31)
	CachedLight        [MaxLightMaps]int // values currently used int lightmap
	CachedDynamicLight bool              // true if dynamic light is in cache
	Samples            []byte            // [numstyles*surfsize]
}

type MNode struct {
	// common with leaf
	Contents int        // 0, to differentiate from leafs
	MinMaxs  [6]float32 // for bounding box culling

	// node specific
	FirstSurface uint
	NumSurfaces  uint
	Plane        *MPlane
	Children     [2]*MNode
}

type MLeaf struct {
	// Common with node
	Contents int        // will be a negative contents number
	MinMaxs  [6]float32 // for bounding box culling

	// Leaf specific
	NumMarkSurfaces   int
	CombinedDeps      int // contains index into brush_deps_data[] with used warp and lightmap textures
	AmbientSoundLevel [NumAmbients]byte
	CompressedVis     []byte
	FirstMarkSurface  *int
	EntityFragments   *EntityFragment
}

type MVertex struct {
	Position math.Vec3[float32]
}

type MEdge struct {
	V                [2]uint
	CachedEdgeOffset uint
}

type MClipNode struct {
	PlaneNum int
	Children [2]int // negative numbers are contents
}

type Hull struct {
	ClipNodes     *MClipNode
	Planes        *MPlane
	FirstClipNode int
	LastClipNode  int
	ClipMins      math.Vec3[float32]
	ClipMaxs      math.Vec3[float32]
}

type QModel struct {
	Name     string
	PathId   int
	NeedLoad bool

	Type      ModelType
	NumFrames int
	SyncType  SyncType

	Flags int

	EmitEffect  int
	TrailEffect int
	SkyTris     *SkyTris
	SkyTriMem   *SkyTriBlock
	SkyTime     float64

	// volume occupied by the model graphics
	Mins  math.Vec3[float32]
	Maxs  math.Vec3[float32]
	YMins math.Vec3[float32] // Entities with nonzero yaw
	YMaxs math.Vec3[float32]
	RMins math.Vec3[float32] // Entities with nonzero pitch or roll
	RMaxs math.Vec3[float32]

	// Solid volume for clipping
	Clipbox  bool
	ClipMins math.Vec3[float32]
	ClipMaxs math.Vec3[float32]

	// Bush model
	FirstModelSurface int
	NumModelSurfaces  int

	Submodels    []DModel
	Planes       []MPlane
	Leafs        []MLeaf
	Vertices     []MVertex
	Edges        []MEdge
	Nodes        []MNode
	TexInfo      []TexInfo
	Surfaces     []MSurface
	SurfEdges    []int
	ClipNodes    []MClipNode
	MarkSurfaces []int

	LeafBounds *AABBStructureOfArrays
	SurfVis    *byte
	SurfPlanes *PlaneStructureOfArrays

	Hulls [MaxMapHulls]Hull

	Textures []Texture

	VisData   []byte
	LightData []byte
	Entities  string

	VisWarn   bool // For Mod_DecompressVis()
	BogusTree bool // BSP node tree doesn't visit nummodelsurfaces

	BSPVersion          int
	ContentsTransparent int // added this so we can disable glitchy wsateralpha where it's not supported

	CombinedDeps int // contains index into brush_deps_data[] with used warp and lightmap textures
	UsedSpecials int // contains SURF_DRAWSKY, SURF_DRAWATER, SURF_DRAWLAVA, SURF_DRAWTELE flags if used by any surf

	WaterSurfs         []int // worldmodel only: list of surface indices with SURF_DRAWTURB flag of transparent types
	UsedWaterSurfs     int
	WaterSurfsSpecials int // which surfaces are in water_surfs (SURF_DRAWATER, SURF_DRAWLAVA, SURF_DRAWESLIME, SURF_DRAWTELE) to track transparency changes

	// additional model data
	ExtraData [PoseVertTypeSize][]byte // only access through Mod_ExtraData

	// Ray tracing
	BottomLevelAccelStructure khr_acceleration_structure.AccelerationStructure
	BLASBuffer                core1_0.Buffer
	BLASAddress               uint64
}
