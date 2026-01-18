package main

import (
	"log"
	"unsafe"

	"github.com/vkngwrapper/core/v3/core1_0"
	"github.com/vkngwrapper/core/v3/core1_1"
	"github.com/vkngwrapper/core/v3/loader"
	"github.com/vkngwrapper/extensions/v3/ext_debug_utils"
	"github.com/vkngwrapper/extensions/v3/ext_full_screen_exclusive"
	"github.com/vkngwrapper/extensions/v3/khr_acceleration_structure"
	"github.com/vkngwrapper/extensions/v3/khr_buffer_device_address"
	"github.com/vkngwrapper/extensions/v3/khr_get_physical_device_properties2"
	"github.com/vkngwrapper/extensions/v3/khr_get_surface_capabilities2"
	"github.com/vkngwrapper/math"
)

const (
	MaxBatchSize             int = 65
	NumColorBuffers          int = 2
	WorldPipelineCount       int = 16
	ModelPipelineCount       int = 6
	FTEParticlePipelineCount int = 16
	MaxPushConstantSize      int = 128
)

type CanvasType int

const (
	CanvasNone CanvasType = iota
	CanvasDefault
	CanvasConsole
	CanvasMenu
	CanvasSBar
	CanvasWarpImage
	CanvasCrosshair
	CanvasBottomLeft
	CanvasTopLeft
	CanvasBottomRight
	CanvasTopRight
	CanvasCSQC
	CanvasInvalid CanvasType = -1
)

type PrimaryCommandBufferContexts int

const (
	PCBXBuildAccelerationStructures PrimaryCommandBufferContexts = iota
	PCBXUpdateLightmaps
	PCBXUpdateWarp
	PCBXRenderPasses
	PCBXCount
)

type SecondaryCommandBufferContexts int

const (
	// Main render pass:
	SCBXWorld SecondaryCommandBufferContexts = iota
	SCBXEntities
	SCBXSky
	SCBXAlphaEntitiesAcrossWater
	SXCBXWater
	SCBXAlphaEntities
	SCBXParticles
	SCBXViewModel

	// UI Renderpass
	SCBXGui
	SCBXPostProcess
	SCBXCount
)

type PipelineLayout struct {
	Layout            core1_0.PipelineLayout
	PushConstantRange core1_0.PushConstantRange
}

type Pipeline struct {
	Pipeline core1_0.Pipeline
	Layout   PipelineLayout
}

type DescriptorSetLayout struct {
	Layout                    core1_0.DescriptorSetLayout
	NumCombinedImageSamplers  int
	NumUBOs                   int
	NumUBOsDynamic            int
	NumStorageBuffers         int
	NumInputAttachments       int
	NumStorageImages          int
	NumSampledImages          int
	NumAccelerationStructures int
}

type CommandBufferContext struct {
	CommandBuffer   core1_0.CommandBuffer
	CurrentCanvas   CanvasType
	RenderPass      core1_0.RenderPass
	RenderPassIndex int
	Subpass         int
	CurrentPipeline Pipeline
	VBOIndices      [MaxBatchSize]uint32
	VBOIndexCount   int
}

type BufferCreateInfo struct {
	Buffer    *core1_0.Buffer
	Size      int
	Alignment int
	Usage     core1_0.BufferUsageFlags
	Mapped    *unsafe.Pointer
	Address   *uint64
	Name      string
}

type VulkanGlobals struct {
	Driver                         core1_0.CoreDeviceDriver
	DeviceIdle                     bool
	Validation                     bool
	Queue                          core1_0.Queue
	PrimaryCommandBufferContexts   [int(PCBXCount)]CommandBufferContext
	SecondaryCommandBufferContexts [int(SCBXCount)]*CommandBufferContext
	ColorClearValue                core1_0.ClearValue
	SwapChainFormat                core1_0.Format
	WantFullScreenExclusive        bool
	SwapChainFullScreenExclusive   bool
	SwapChainFullScreenAcquired    bool
	DeviceProperties               core1_0.PhysicalDeviceProperties
	DeviceFeatures                 core1_0.PhysicalDeviceFeatures
	MemoryProperties               core1_0.PhysicalDeviceMemoryProperties
	GraphicsQueueFamilyIndex       int
	ColorFormat                    core1_0.Format
	DepthFormat                    core1_0.Format
	SampleCount                    core1_0.SampleCountFlags
	Supersampling                  bool
	NonSolidFill                   bool
	MultiDrawIndirect              bool
	ScreenEffectsSops              bool

	// Instance Extensions
	Vulkan1_1                    core1_1.CoreDeviceDriver
	GetSurfaceCapabilities2      khr_get_surface_capabilities2.ExtensionDriver
	GetPhysicalDeviceProperties2 khr_get_physical_device_properties2.ExtensionDriver

	// Device Extensions
	DedicatedAllocation             bool
	FullScreenExclusive             ext_full_screen_exclusive.ExtensionDriver
	RayQuery                        bool
	BufferDeviceAddress             khr_buffer_device_address.ExtensionDriver
	AccelerationStructure           khr_acceleration_structure.ExtensionDriver
	AccelerationStructureProperties khr_acceleration_structure.PhysicalDeviceAccelerationStructureProperties
	DebugUtils                      ext_debug_utils.ExtensionDriver

	// Buffers
	ColorBuffers [NumColorBuffers]core1_0.Image

	// Index buffers
	FanIndexBuffer core1_0.Buffer

	// Staging buffers
	StagingBufferSize int

	// Render passes
	MainRenderPass [2]core1_0.RenderPass
	WarpRenderPass core1_0.RenderPass

	// Pipelines
	BasicAlphatestPipeline         [2]Pipeline
	BasicBlendPipeline             [2]Pipeline
	BasicNotexBlendPipeline        [2]Pipeline
	BasicPipelineLayout            PipelineLayout
	WorldPipelines                 [WorldPipelineCount]Pipeline
	WorldPipelineLayout            PipelineLayout
	RasterTexWarpPipeline          Pipeline
	ParticlePipeline               Pipeline
	SpritePipeline                 Pipeline
	SkyPipelineLayout              [2]PipelineLayout
	SkyStencilPipeline             [2]Pipeline
	SkyColorPipeline               [2]Pipeline
	SkyBoxPipeline                 Pipeline
	SkyCubePipeline                [2]Pipeline
	SkyLayerPipeline               [2]Pipeline
	AliasPipelines                 [ModelPipelineCount]Pipeline
	MD5Pipelines                   [ModelPipelineCount]Pipeline
	PostprocessPipeline            Pipeline
	ScreenEffectsPipeline          Pipeline
	ScreenEffectsScalePipeline     Pipeline
	ScreenEffectsScaleSopsPipeline Pipeline
	CSTexWarpPipeline              Pipeline
	ShowtrisPipeline               Pipeline
	ShowtrisIndirectPipeline       Pipeline
	ShowtrisDepthTestPipeline      Pipeline
	ShowbboxesPipeline             Pipeline
	UpdateLightmapPipeline         Pipeline
	UpdateLightmapRtPipeline       Pipeline
	IndirectDrawPipeline           Pipeline
	IndirectClearPipeline          Pipeline
	RayDebugPipeline               Pipeline
	FTEParticlePipelines           [FTEParticlePipelineCount]Pipeline

	// Descriptors
	DescriptorPool                core1_0.DescriptorPool
	UBOSetLayout                  DescriptorSetLayout
	SingleTextureSetLayout        DescriptorSetLayout
	InputAttachmentSetLayout      DescriptorSetLayout
	ScreenEffectsDescSet          core1_0.DescriptorSet
	ScreenEffectsSetLayout        DescriptorSetLayout
	SingleTextureCSWriteSetLayout DescriptorSetLayout
	LightmapComputeSetLayout      DescriptorSetLayout
	IndirectComputeDescSet        core1_0.DescriptorSet
	IndirectComputeSetLayout      DescriptorSetLayout
	LightmapComputeRTSetLayout    DescriptorSetLayout
	RayDebugDescSet               core1_0.DescriptorSet
	RayDebugSetLayout             DescriptorSetLayout
	JointsBufferSetLayout         DescriptorSetLayout

	// Samplers
	PointSampler              core1_0.Sampler
	LinearSampler             core1_0.Sampler
	PointAnisoSampler         core1_0.Sampler
	LinearAnisoSampler        core1_0.Sampler
	PointSamplerLodBias       core1_0.Sampler
	LinearSamplerLodBias      core1_0.Sampler
	PointAnisoSamplerLodBias  core1_0.Sampler
	LinearAnisoSamplerLodBias core1_0.Sampler

	// Matrices
	ProjectionMatrix     math.Mat4x4[float32]
	ViewMatrix           math.Mat4x4[float32]
	ViewProjectionMatrix math.Mat4x4[float32]
}

var Vulkan = &VulkanGlobals{}

func RenderBindPipeline(context *CommandBufferContext, bindPoint core1_0.PipelineBindPoint, pipeline Pipeline) {
	var zeroes [MaxPushConstantSize]byte
	if !pipeline.Pipeline.Initialized() {
		log.Fatalln("Pipeline cannot be uninitialized")
	}
	if !pipeline.Layout.Layout.Initialized() {
		log.Fatalln("PipelineLayout cannot be uninitialized")
	}
	if context.CurrentPipeline.Layout.PushConstantRange.Size > MaxPushConstantSize {
		log.Fatalf("Pipeline PushConstantRange size must be less than or equal to %d but is %d", MaxPushConstantSize, context.CurrentPipeline.Layout.PushConstantRange.Size)
	}

	if context.CurrentPipeline.Pipeline.Handle() != pipeline.Pipeline.Handle() {
		Vulkan.Driver.CmdBindPipeline(context.CommandBuffer, bindPoint, pipeline.Pipeline)
		if pipeline.Layout.PushConstantRange.Size > 0 &&
			(context.CurrentPipeline.Layout.PushConstantRange.StageFlags != pipeline.Layout.PushConstantRange.StageFlags ||
				context.CurrentPipeline.Layout.PushConstantRange.Size != pipeline.Layout.PushConstantRange.Size) {
			Vulkan.Driver.CmdPushConstants(context.CommandBuffer, pipeline.Layout.Layout, pipeline.Layout.PushConstantRange.StageFlags, 0, zeroes[:pipeline.Layout.PushConstantRange.Size])
		}

		context.CurrentPipeline = pipeline
	}
}

func RenderPushConstants(context *CommandBufferContext, stageFlags core1_0.ShaderStageFlags, offset int, data []byte) {
	Vulkan.Driver.CmdPushConstants(context.CommandBuffer, context.CurrentPipeline.Layout.Layout, stageFlags, offset, data)
}

func RenderBeginDebugUtilsLabel(context *CommandBufferContext, name string) {
	if Vulkan.DebugUtils != nil {
		Vulkan.DebugUtils.CmdBeginDebugUtilsLabel(context.CommandBuffer, ext_debug_utils.DebugUtilsLabel{
			LabelName: name,
		})
	}
}

func RenderEndDebugUtilsLabel(context *CommandBufferContext) {
	if Vulkan.DebugUtils != nil {
		Vulkan.DebugUtils.CmdEndDebugUtilsLabel(context.CommandBuffer)
	}
}

func SetObjectName(object loader.VulkanHandle, objectType core1_0.ObjectType, name string) {
	if Vulkan.DebugUtils != nil {
		Vulkan.DebugUtils.SetDebugUtilsObjectName(
			Vulkan.Driver.Device(),
			ext_debug_utils.DebugUtilsObjectNameInfo{
				ObjectName:   name,
				ObjectType:   objectType,
				ObjectHandle: object,
			})
	}
}
