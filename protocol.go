package main

import "github.com/vkngwrapper/math"

type EntityState struct {
	Origin         math.Vec3[float32]
	Angles         math.Vec3[float32]
	ModelIndex     uint16
	Frame          uint16
	Effects        uint
	ColorMap       uint8
	Skin           uint8
	Scale          uint8
	PMoveType      uint8
	TrailEffectNum uint16   // For qc-defined particle trails. used for things that are not trails
	EmitEffectNum  uint16   // For qc-defined particle trails. used for things that are not trails
	Velocity       [3]int16 // player's velocity
	EFlags         uint8
	TagIndex       uint8
	TagEntity      uint16
	Pad            uint16
	ColorMod       [3]uint8 // entity tints
	Alpha          uint8
	SolidSize      uint32 // For CSQC prediction logic
	Lerp           uint16
}
