package main

import "github.com/vkngwrapper/math"

type Particle struct {
	Next *Particle

	Die float32

	// driver-usable fields
	Org            math.Vec3[float32]
	RGBA           math.Vec4[float32]
	Scale          float32
	S1, T1, S2, T2 float32

	OldOrg   math.Vec3[float32] // to throttle traces
	Velocity math.Vec3[float32] // renderer uses for sparks
	Angle    float32

	NextEmit   float32
	TrailState *TrailState

	// Drivers never touch the following fields
	RotationSpeed float32
}

type SkyTris struct {
	Next *SkyTris

	Org      math.Vec3[float32]
	X        math.Vec3[float32]
	Y        math.Vec3[float32]
	Area     float32
	NextTime float64
	PType    int
	Face     *MSurface
}

type SkyTriBlock struct {
	Next *SkyTriBlock

	Count uint
	Tris  [1024]SkyTris
}

type BeamSegment struct {
	Next *BeamSegment

	Particle *Particle
	Flags    int32
	Dir      math.Vec3[float32]

	Texture float32
}

type TrailState struct {
	Key      **TrailState // key to change if trailstate has been overwritten
	Assoc    *TrailState  // associated linked trail
	LastBeam *BeamSegment

	LastDist  float32 // last distance used with particle effect
	StateTime float32 // time to emit effect again (used by spawntime field)
	LastStop  float32 // last stopping point for particle effect
	EmitTime  float32 // used by render_effect emitters
}
