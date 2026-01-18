package main

import (
	"log"

	"github.com/vkngwrapper/core/v3/core1_0"
)

var CVarRLodBias = &CVar{
	Name:      "r_lodbias",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarGLLodBias = &CVar{
	Name:      "gl_lodbias",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}

const (
	NumStagingBuffers int = 2
)

type StagingBuffer struct {
	Buffer        core1_0.Buffer
	CommandBuffer core1_0.CommandBuffer
	Fence         core1_0.Fence
	CurrentOffset int
	Submitted     bool
	Data          []byte
}

type DynamicBuffer struct {
	Buffer        core1_0.Buffer
	CurrentOffset int
	Data          []byte
	DeviceAddress uint64
}

type RenderResourceData struct {
	stagingBuffers [NumStagingBuffers]StagingBuffer
}

func (r *RenderResourceData) CreateStagingBuffers() {
	bufferCreate := core1_0.BufferCreateInfo{
		Size:  Vulkan.StagingBufferSize,
		Usage: core1_0.BufferUsageTransferSrc,
	}

	var err error
	for i := 0; i < NumStagingBuffers; i++ {
		r.stagingBuffers[i].CurrentOffset = 0
		r.stagingBuffers[i].Submitted = false

		r.stagingBuffers[i].Buffer, _, err = Vulkan.Driver.CreateBuffer(nil, bufferCreate)
		if err != nil {
			log.Fatalln("CreateBuffer failed")
		}

	}
}

var RenderResources = &RenderResourceData{}

func (v *VulkanGlobals) SetClearColor() {
	if CVarRFastClear.Value != 0 {
		Vulkan.ColorClearValue = core1_0.ClearValueFloat{0, 0, 0, 0}
	} else {
		s := int(CVarRClearColor.Value) & 0xff
		rgb := D8To24Table[s]
		r := float32(rgb&0xff) / 255.0
		g := float32((rgb>>8)&0xff) / 255.0
		b := float32((rgb>>16)&0xff) / 255.0
		Vulkan.ColorClearValue = core1_0.ClearValueFloat{r, g, b, 0}
	}
}

func (v *VulkanGlobals) CVarSetClearColor(cvar *CVar) {
	if CVarRFastClear.Value != 0.0 {
		log.Println("Black clear color forced by r_fastclear")
	}

	v.SetClearColor()
}

func (v *VulkanGlobals) CVarSetFastClear(cvar *CVar) {
	v.SetClearColor()
}

// TODO: SIMD
