package main

import "github.com/vkngwrapper/math"

type LightCache struct {
	SurfaceIndex int32 // <0 black surface, ==0 no cache, >0 1+index of surface
	Pos          math.Vec3[float32]
	DS           uint16
	DT           uint16
}

type EntityFragment struct {
	LeafNext *EntityFragment

	Entity *Entity
}

type Entity struct {
	ForceLink bool // model changed

	UpdateType int32

	Baseline EntityState // to fill in defaults in updates
	NetState EntityState // the latest network state

	MsgTime    float64               // time of last update
	MsgOrigins [2]math.Vec3[float32] // Last two updates - 0 is newest
	Origin     math.Vec3[float32]
	MsgAngles  [2]math.Vec3[float32] // last two updates - 0 is newest
	Angles     math.Vec3[float32]
	Model      *QModel         // nil = no model
	EFrag      *EntityFragment // linked list of entity fragments
	Frame      int32
	SyncBase   float32 // for client-side animations
	ColorMap   []byte
	Effects    int32 // light, particles, etc.
	SkinNum    int32 // for alias models
	VisFrame   int32 // last frame this entity was found in an active leaf

	DynamicLightFrame int32
	DynamicLightBits  int32

	TopNode *MNode // for bmodels, first world node that splits bmodel, or NULL if not split

	EFlags       byte // Mostly a mirror of netstate, but handles tag inheritance
	Alpha        byte
	LerpFlags    byte
	LerpStart    float32 // animation lerping
	LerpTime     float32
	LerpFinish   float32 // server sent us a more accurate interval, use it instead of 0.1
	PreviousPose uint16  // animation lerping
	CurrentPose  uint16

	MoveLerpStart  float32 // transform lerping
	PreviousOrigin math.Vec3[float32]
	CurrentOrigin  math.Vec3[float32]
	PreviousAngles math.Vec3[float32]
	CurrentAngles  math.Vec3[float32]

	TrailState *TrailState // managed by the particle system, so we don't lose our position and spawn the wrong
	// number of particles, and we can track beams etc.
	EmitState  *TrailState        // for effects which are not so static
	TrailDelay float32            // time left until next particle trail update
	TrailOrg   math.Vec3[float32] // previous particle trail point

	LightCache LightCache

	ContentsCache       int32
	ContentsCacheOrigin math.Vec3[float32]
}

type RenderDefinition struct {
	VideoRect                                 VideoRect // subwindow in video for refresh
	AliasVideoRect                            VideoRect // scaled alias version
	VideoRectRight, VideoRectBottom           int
	AliasVideoRectRight, AliasVideoRectBottom int
	VideoRectRightEdge                        float32
	VideoRectX, VideoRectY                    float32
	VideoRectXAdjusted, VideoRectYAdjusted    float32
	VideoRectXAdjustedShift20                 int
	VideoRectRightAdjustedShift20             int
	VideoRectRightAdjusted                    float32
	VideoRectBottomAdjusted                   float32

	VideoRectRightFloat   float32
	VideoRectBottomFloat  float32
	HorizontalFieldOfView float32
	XOrigin               float32
	YOrigin               float32

	ViewOrigin math.Vec3[float32]
	ViewAngles math.Vec3[float32]

	BaseFOV    float32
	FovX, FovY float32

	AmbientLight int
}

var RenderData = &RenderDefinition{}
