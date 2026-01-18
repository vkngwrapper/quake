package main

import (
	"encoding/binary"
	"unsafe"

	"github.com/vkngwrapper/core/v3/core1_0"
	"github.com/vkngwrapper/math"
)
import stdmath "math"

var CVarRDrawEntities = &CVar{
	Name:      "r_drawentities",
	StringVal: "1",
}
var CVarRDrawViewModel = &CVar{
	Name:      "r_drawviewmodel",
	StringVal: "1",
}
var CVarRSpeeds = &CVar{
	Name:      "r_speeds",
	StringVal: "0",
}
var CVarRPos = &CVar{
	Name:      "r_pos",
	StringVal: "0",
}
var CVarRFullBright = &CVar{
	Name:      "r_fullbright",
	StringVal: "0",
}
var CVarRLightMap = &CVar{
	Name:      "r_lightmap",
	StringVal: "0",
}
var CVarRWaterAlpha = &CVar{
	Name:      "r_wateralpha",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRDynamic = &CVar{
	Name:      "r_dynamic",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRNovis = &CVar{
	Name:      "r_novis",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarRSIMD = &CVar{
	Name:      "r_simd",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRAlphaSort = &CVar{
	Name:      "r_alphasort",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}

var CVarGLFinish = &CVar{
	Name:      "gl_finish",
	StringVal: "0",
}
var CVarGLPolyBlend = &CVar{
	Name:      "gl_polyblend",
	StringVal: "1",
}
var CVarGLNoColors = &CVar{
	Name:      "gl_nocolors",
	StringVal: "0",
}

var CVarRClearColor = &CVar{
	Name:      "r_clearcolor",
	StringVal: "2",
	Flags:     CVarFlagArchive,
}
var CVarRFastClear = &CVar{
	Name:      "r_fastclear",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRFlatLightStyles = &CVar{
	Name:      "r_flatlightstyles",
	StringVal: "0",
}
var CVarRLerpLightStyles = &CVar{
	Name:      "r_lerplightstyles",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarGLFullBrights = &CVar{
	Name:      "gl_fullbrights",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarGLFarClip = &CVar{
	Name:      "gl_farclip",
	StringVal: "16384",
	Flags:     CVarFlagArchive,
}
var CVarROldSkyLeaf = &CVar{
	Name:      "r_oldskyleaf",
	StringVal: "0",
}
var CVarRDrawWorld = &CVar{
	Name:      "r_drawworld",
	StringVal: "1",
}
var CVarRShowTris = &CVar{
	Name:      "r_showtris",
	StringVal: "0",
}
var CVarRShowBBoxes = &CVar{
	Name:      "r_showbboxes",
	StringVal: "0",
}
var CVarRShowBBoxesFilter = &CVar{
	Name:      "r_showbboxes_filter",
	StringVal: "",
}
var CVarRLerpModels = &CVar{
	Name:      "r_lerpmodels",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRLerpMove = &CVar{
	Name:      "r_lerpmove",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRLerpTurn = &CVar{
	Name:      "r_lerpturn",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRNoLerpList = &CVar{
	Name: "r_nolerp_list",
	StringVal: "progs/flame.mdl,progs/flame2.mdl,progs/braztall.mdl,progs/brazshrt.mdl,progs/longtrch.mdl," +
		"progs/flame_pyre.mdl,progs/v_saw.mdl,progs/v_xfist.mdl,progs/h2stuff/newfire.mdl",
}
var CVarGLZFix = &CVar{
	Name:      "gl_zfix",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRLavaAlpha = &CVar{
	Name:      "r_lavaalpha",
	StringVal: "0",
}
var CVarRTeleAlpha = &CVar{
	Name:      "r_telealpha",
	StringVal: "0",
}
var CVarRSlimeAlpha = &CVar{
	Name:      "r_slimealpha",
	StringVal: "0",
}
var CVarRScale = &CVar{
	Name:      "r_scale",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRGPULightMapUpdate = &CVar{
	Name:      "r_gpulightmapupdate",
	StringVal: "1",
}
var CVarRRTShadows = &CVar{
	Name:      "r_rtshadows",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}
var CVarRTasks = &CVar{
	Name:      "r_tasks",
	StringVal: "1",
}
var CVarRIndirect = &CVar{
	Name:      "r_indirect",
	StringVal: "1",
}

const (
	NearClip float32 = 4
)

type CameraData struct {
	planes [4]MPlane

	viewNormal math.Vec3[float32]
	viewUp     math.Vec3[float32]
	viewRight  math.Vec3[float32]
	viewOrigin math.Vec3[float32]

	fovX float32
	fovY float32

	viewLeaf    *MLeaf
	oldViewLeaf *MLeaf

	renderWarp  bool
	renderScale int

	drawWorldCheatSafe  bool
	fullBrightCheatSafe bool
	lightMapCheatSafe   bool
}

var Camera = &CameraData{}

func (f *CameraData) CullBox(mins *math.Vec3[float32], maxs *math.Vec3[float32]) bool {
	for i := 0; i < 4; i++ {
		plane := f.planes[i]
		signBits := plane.SignBits
		vec := maxs
		if signBits&1 != 0 {
			vec.X = mins.X
		}
		if signBits&2 != 0 {
			vec.Y = mins.Y
		}
		if signBits&4 != 0 {
			vec.Z = mins.Z
		}
		if vec.DotProduct(&plane.Normal) < plane.Dist {
			return true
		}
	}

	return false
}

func (f *CameraData) CullModelForEntity(e *Entity) bool {
	var mins, maxs math.Vec3[float32]
	var minBounds, maxBounds *math.Vec3[float32]

	if e.Angles.X != 0 || e.Angles.Z != 0 {
		minBounds = &e.Model.RMins
		maxBounds = &e.Model.RMaxs
	} else if e.Angles.Y != 0 {
		minBounds = &e.Model.YMins
		maxBounds = &e.Model.YMaxs
	} else {
		minBounds = &e.Model.Mins
		maxBounds = &e.Model.Maxs
	}

	scaleFactor := float32(e.NetState.Scale) / 16.0
	if scaleFactor != 1.0 {
		mins.SetScale(minBounds, scaleFactor)
		mins.AddVec3(&e.Origin)
		maxs.SetScale(maxBounds, scaleFactor)
		maxs.AddVec3(&e.Origin)
	} else {
		mins.SetAddVec3(&e.Origin, minBounds)
		maxs.SetAddVec3(&e.Origin, maxBounds)
	}

	return f.CullBox(&mins, &maxs)
}

func (f *CameraData) SetFrustum(fovX, fovY float32) {
	f.planes[0].Normal.SetRotateTowards(&f.viewNormal, math.ToRadians(fovX/2.0-90), &f.viewRight)
	f.planes[1].Normal.SetRotateTowards(&f.viewNormal, math.ToRadians(90-fovX/2.0), &f.viewRight)
	f.planes[2].Normal.SetRotateTowards(&f.viewNormal, math.ToRadians(90-fovY/2.0), &f.viewUp)
	f.planes[3].Normal.SetRotateTowards(&f.viewNormal, math.ToRadians(fovY/2.0-90), &f.viewUp)

	for i := 0; i < 4; i++ {
		f.planes[i].Type = byte(PlaneAnyZ)
		f.planes[i].Dist = f.viewOrigin.DotProduct(&f.planes[i].Normal)
		f.planes[i].SignBits = SignBitsForPlane(&f.planes[i])
	}
}

func (f *CameraData) SetupMatrices() {
	// Projection matrix
	Vulkan.ProjectionMatrix.SetAdaptiveNearPlanePerspective(
		math.ToRadians(f.fovY), f.fovX/f.fovY, float32(CVarGLFarClip.Value), 0.5, NearClip,
	)

	// View matrix
	Vulkan.ViewMatrix.SetRotationX(-stdmath.Pi / 2)
	Vulkan.ViewMatrix.RotateZ(stdmath.Pi / 2)
	Vulkan.ViewMatrix.RotateX(math.ToRadians(-RenderData.ViewAngles.Z))
	Vulkan.ViewMatrix.RotateY(math.ToRadians(-RenderData.ViewAngles.X))
	Vulkan.ViewMatrix.RotateZ(math.ToRadians(-RenderData.ViewAngles.Y))
	Vulkan.ViewMatrix.Translate(-RenderData.ViewOrigin.X, -RenderData.ViewOrigin.Y, -RenderData.ViewOrigin.Z)

	// View projection matrix
	Vulkan.ViewProjectionMatrix.SetMultMat4x4(&Vulkan.ProjectionMatrix, &Vulkan.ViewMatrix)
}

func (c *CameraData) RenderSetupViewBeforeMark() {
	// Need to do this early because we now update dynamic light maps during R_MarkSurfaces
	if CVarRGPULightMapUpdate.Value != 0 {
		// TODO: R_PushDLights
	}
	// TODO: R_AnimateLight

	// Build the transformation matrix for the given view angles
	RenderData.ViewOrigin.SetVec3(&c.viewOrigin)
	math.EulersToVectors(
		math.ToRadians(RenderData.ViewAngles.X),
		math.ToRadians(RenderData.ViewAngles.Y),
		math.ToRadians(RenderData.ViewAngles.Z),
		&c.viewNormal,
		&c.viewRight,
		&c.viewUp,
	)

	// Current viewleaf
	c.oldViewLeaf = c.viewLeaf
	// TODO: Mod_PointINLeaf

	c.fovX = RenderData.FovX
	c.fovY = RenderData.FovY
	c.renderWarp = false
	c.renderScale = int(CVarRScale.Value)

	// TODO: Warp

	c.SetFrustum(c.fovX, c.fovY)
	c.SetupMatrices()

	c.fullBrightCheatSafe = false
	c.lightMapCheatSafe = false
	c.drawWorldCheatSafe = true
	// TODO: Client
}

func RotateForEntity(matrix *math.Mat4x4[float32], origin *math.Vec3[float32], angles *math.Vec3[float32], scale uint8) {
	matrix.Translate(origin.X, origin.Y, origin.Z)

	matrix.RotateZ(math.ToRadians(float64(angles.Y)))
	matrix.RotateY(math.ToRadians(float64(-angles.X)))
	matrix.RotateX(math.ToRadians(float64(angles.Z)))

	scaleFactor := float32(scale) / 16.0
	if scaleFactor != 1.0 {
		matrix.Scale(scaleFactor, scaleFactor, scaleFactor)
	}
}

func SignBitsForPlane(plane *MPlane) byte {
	var bits byte
	if plane.Normal.X < 0 {
		bits |= 1
	}
	if plane.Normal.Y < 0 {
		bits |= 2
	}
	if plane.Normal.Z < 0 {
		bits |= 4
	}

	return bits
}

func RenderSetupCommandBufferContext(context *CommandBufferContext) {
	// TODO: GL_Viewport
	RenderBindPipeline(context, core1_0.PipelineBindPointGraphics, Vulkan.BasicBlendPipeline[context.RenderPassIndex])

	matrixData := make([]byte, 0, 16*int(unsafe.Sizeof(float32(0))))
	_, _ = binary.Append(matrixData, binary.LittleEndian, Vulkan.ViewProjectionMatrix)
	RenderPushConstants(context, core1_0.StageAllGraphics, 0, matrixData)
}
