package main

import (
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/vkngwrapper/core/v3/core1_0"
)

const (
	DefaultRefreshRate         int = 60
	InitialStagingBufferSizeKB int = 16384
)

var CVarVidFullScreen = CVar{
	Name:      "vid_fullscreen",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidWidth = CVar{
	Name:      "vid_width",
	StringVal: "1280",
	Flags:     CVarFlagArchive,
}
var CVarVidHeight = CVar{
	Name:      "vid_height",
	StringVal: "720",
	Flags:     CVarFlagArchive,
}
var CVarVidRefreshRate = CVar{
	Name:      "vid_refreshrate",
	StringVal: "60",
	Flags:     CVarFlagArchive,
}
var CVarVidVSync = CVar{
	Name:      "vid_vsync",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidDesktopFullscreen = CVar{
	Name:      "vid_desktopfullscreen",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidBorderless = CVar{
	Name:      "vid_borderless",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidPalettize = CVar{
	Name:      "vid_palettize",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidFilter = CVar{
	Name:      "vid_filter",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidAnisotropic = CVar{
	Name:      "vid_anisotropic",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidFSAA = CVar{
	Name:      "vid_fsaa",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidFSAAMode = CVar{
	Name:      "vid_fsaamode",
	StringVal: "0",
	Flags:     CVarFlagArchive,
}
var CVarVidGamma = CVar{
	Name:      "gamma",
	StringVal: "0.9",
	Flags:     CVarFlagArchive,
}
var CVarVidContrast = CVar{
	Name:      "contrast",
	StringVal: "1.4",
	Flags:     CVarFlagArchive,
}
var CVarRenderUsesOps = CVar{
	Name:      "r_usesops",
	StringVal: "1",
	Flags:     CVarFlagArchive,
}

type VideoMode struct {
	Width       int
	Height      int
	RefreshRate int
}

type MenuMode struct {
	Width  int
	Height int
}

type SDLVideo struct {
	drawContext     *sdl.Window
	previousDisplay int
	modeState       ModeState
	modes           []VideoMode
	menuModes       []MenuMode
	locked          bool
	changed         bool
	initialized     bool

	PaletteColorsBuffer core1_0.Buffer
	PaletteOctreeBuffer core1_0.Buffer
}

var Video = &SDLVideo{}

func (v *SDLVideo) gammaInit() {
	CVars.Register(&CVarVidGamma)
	CVars.Register(&CVarVidContrast)
}

func (v *SDLVideo) GetCurrentWidth() int {
	w, _ := v.drawContext.VulkanGetDrawableSize()
	return int(w)
}

func (v *SDLVideo) GetCurrentHeight() int {
	_, h := v.drawContext.VulkanGetDrawableSize()
	return int(h)
}

func (v *SDLVideo) GetCurrentRefreshRate() int {
	display, err := v.drawContext.GetDisplayIndex()
	if display < 0 || err != nil {
		display = 0
	}

	mode, err := sdl.GetCurrentDisplayMode(display)
	if err != nil {
		return DefaultRefreshRate
	}

	return int(mode.RefreshRate)
}

func (v *SDLVideo) GetCurrentBPP() int {
	pixelFormat, _ := v.drawContext.GetPixelFormat()
	return int((pixelFormat >> 8) & 0xff)
}

func (v *SDLVideo) GetFullscreen() bool {
	return v.drawContext.GetFlags()&sdl.WINDOW_FULLSCREEN != 0
}

func (v *SDLVideo) GetDesktopFullscreen() bool {
	return v.drawContext.GetFlags()&sdl.WINDOW_FULLSCREEN_DESKTOP != 0
}

func (v *SDLVideo) GetWindow() *sdl.Window {
	return v.drawContext
}

func (v *SDLVideo) HasMouseOrInputFocus() bool {
	return v.drawContext.GetFlags()&(sdl.WINDOW_MOUSE_FOCUS|sdl.WINDOW_INPUT_FOCUS) != 0
}

func (v *SDLVideo) IsMinimized() bool {
	return v.drawContext.GetFlags()&sdl.WINDOW_SHOWN == 0
}

func (v *SDLVideo) GetDisplayMode(width int, height int, refreshRate int) *sdl.DisplayMode {
	modeCount, _ := sdl.GetNumDisplayModes(0)

	for i := 0; i < modeCount; i++ {
		mode, err := sdl.GetDisplayMode(0, i)
		if err != nil {
			continue
		}

		bpp := (mode.Format >> 8) & 0xff
		if int(mode.W) == width && int(mode.H) == height && bpp >= 24 && int(mode.RefreshRate) == refreshRate {
			return &mode
		}
	}

	return nil
}

func (v *SDLVideo) ValidMode(width, height, refreshRate int, fullScreen bool) bool {
	if fullScreen && CVarVidDesktopFullscreen.Value != 0 {
		return true
	}

	if width < 320 || height < 200 {
		return false
	}

	if fullScreen && v.GetDisplayMode(width, height, refreshRate) == nil {
		return false
	}

	return true
}

func (v *SDLVideo) centeredDisplay(displayIndex int) int32 {
	var mask uint32 = 0x2fff0000
	return int32(mask | uint32(displayIndex))
}

func (v *SDLVideo) SetMode(width, height, refreshRate int, fullScreen bool) {
	// TODO: Temporarily disable screen for loading
	// TODO: Pause all audio

	caption := EngineNameAndVersion

	var err error
	if v.drawContext == nil {
		flags := uint32(sdl.WINDOW_HIDDEN | sdl.WINDOW_VULKAN)

		if CVarVidBorderless.Value != 0 {
			flags |= sdl.WINDOW_BORDERLESS
		} else if !fullScreen {
			flags |= sdl.WINDOW_RESIZABLE
		}

		v.drawContext, err = sdl.CreateWindow(caption, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(width), int32(height), flags)
		if err != nil {
			log.Fatalf("Couldn't create window: %s\n", err)
		}

		v.previousDisplay = -1
	} else {
		v.previousDisplay, _ = v.drawContext.GetDisplayIndex()
	}

	if v.GetFullscreen() {
		err = v.drawContext.SetFullscreen(0)
		if err != nil {
			log.Fatalf("Couldn't set fullscreen state mode: %s", err)
		}
	}

	v.drawContext.SetSize(int32(width), int32(height))
	if v.previousDisplay >= 0 {
		v.drawContext.SetPosition(v.centeredDisplay(v.previousDisplay), v.centeredDisplay(v.previousDisplay))
	} else {
		v.drawContext.SetPosition(v.centeredDisplay(0), v.centeredDisplay(0))
	}

	_ = v.drawContext.SetDisplayMode(v.GetDisplayMode(width, height, refreshRate))

	var borderless bool
	if CVarVidBorderless.Value != 0 {
		borderless = true
	}
	v.drawContext.SetBordered(borderless)

	if fullScreen {
		fullScreenFlag := uint32(sdl.WINDOW_FULLSCREEN)
		if CVarVidDesktopFullscreen.Value != 0 {
			fullScreenFlag = sdl.WINDOW_FULLSCREEN_DESKTOP
		}
		if err = v.drawContext.SetFullscreen(fullScreenFlag); err != nil {
			log.Fatalf("Couldn't set fullscreen state mode: %s", err)
		}
	}

	v.drawContext.Show()
	v.drawContext.Raise()

	VideoData.Width = v.GetCurrentWidth()
	VideoData.Height = v.GetCurrentHeight()
	VideoData.ConWidth = VideoData.Width & 0xfffffff8
	VideoData.ConHeight = VideoData.ConWidth * VideoData.Height / VideoData.Width

	modeState := ModeStateWindowed
	if v.GetFullscreen() {
		modeState = ModeStateFullscreen
	}
	v.modeState = modeState

	// TODO: Resume all audio
	// TODO: Restore disable screen for loading

	// fix the leftover alt from any alt-tab or the lilke that switched us away
	v.ClearAllStates()

	VideoData.RecalcRefDef = true

	// No pending changes
	v.changed = false

	// TODO: SCR_UpdateRelativeScale()
}

func (v *SDLVideo) CVarChanged(cvar *CVar) {
	v.changed = true
}

func (v *SDLVideo) CVarFilterChanged(cvar *CVar) {
	// TODO: R_InitSamplers()
}

//func (v *SDLVideo) CVarFSAAChanged(cvar *CVar) {
//	v.Restart(false)
//}

//func (v *SDLVideo) CmdTest() {
//	if v.locked || !v.changed {
//		return
//	}
//
//	// now try the switch
//	oldWidth := v.GetCurrentWidth()
//	oldHeight := v.GetCurrentHeight()
//	oldRefreshRate := v.GetCurrentRefreshRate()
//	oldFullscreen := 0
//
//	if v.GetFullscreen() {
//		if Vulkan.SwapChainFullScreenExclusive {
//			oldFullscreen = 2
//		} else {
//			oldFullscreen = 1
//		}
//	}
//	v.Restart(true)
//
//	// TODO: Pop up confirmaation dialogue SCR_ModalMessage
//}

func (v *SDLVideo) CmdUnlock() {
	v.locked = false
	v.SyncCVars()
}

// Lock - Subsequent changes to vid cvars and Restart() commands will be ignored until
// CmdUnlock() is run
//
// Used when changing gamedirs so the current settings override what was saved in the config.cfg
func (v *SDLVideo) Lock() {
	v.locked = true
}

func (v *SDLVideo) ClearAllStates() {
	// TODO: Clear key states
	// TODO: Clear input states
}

func (v *SDLVideo) CmdDescribeCurrentMode() {
	if v.drawContext != nil {
		fullscreen := "windowed"
		if v.GetFullscreen() {
			fullscreen = "fullscreen"
		}
		log.Printf("%dx%dx%d %dHz %s\n", v.GetCurrentWidth(), v.GetCurrentHeight(), v.GetCurrentBPP(), v.GetCurrentRefreshRate(), fullscreen)
	}
}

func (v *SDLVideo) CmdDescribeModes() {
	var lastWidth, lastHeight, count int

	for _, mode := range v.modes {
		if lastWidth != mode.Width || lastHeight != mode.Height {
			log.Printf("   %d x %d : %d\n", mode.Width, mode.Height, mode.RefreshRate)
			lastWidth = mode.Width
			lastHeight = mode.Height
			count++
		}
	}

	log.Printf("%d modes\n", count)
}

//func (v *SDLVideo) initModeList() {
//	modeCount, _ := sdl.GetNumDisplayModes(0)
//	for i := 0; i < modeCount; i++ {
//		mode, err := sdl.GetDisplayMode(0, i)
//		if err == nil {
//			v.modes = append(v.modes, VideoMode{
//				Width:       int(mode.W),
//				Height:      int(mode.H),
//				RefreshRate: int(mode.RefreshRate),
//			})
//		}
//	}
//}
//
//func (v *SDLVideo) Init() {
//	readVars := []string{"vid_fullscreen", "vid_width", "vid_height", "vid_refreshrate",
//		"vid_vsync", "vid_desktopfullscreen", "vid_fsaa", "vid_borderless"}
//	CVars.Register(&CVarVidFullScreen)
//	CVars.Register(&CVarVidWidth)
//	CVars.Register(&CVarVidHeight)
//	CVars.Register(&CVarVidRefreshRate)
//	CVars.Register(&CVarVidVSync)
//	CVars.Register(&CVarVidFilter)
//	CVars.Register(&CVarVidAnisotropic)
//	CVars.Register(&CVarVidFSAAMode)
//	CVars.Register(&CVarVidFSAA)
//	CVars.Register(&CVarVidDesktopFullscreen)
//	CVars.Register(&CVarVidBorderless)
//	CVars.Register(&CVarVidPalettize)
//	InitDebug()
//
//	CVars.SetCallback(&CVarVidFullScreen, v.CVarChanged)
//	CVars.SetCallback(&CVarVidWidth, v.CVarChanged)
//	CVars.SetCallback(&CVarVidHeight, v.CVarChanged)
//	CVars.SetCallback(&CVarVidRefreshRate, v.CVarChanged)
//	CVars.SetCallback(&CVarVidFilter, v.CVarFilterChanged)
//	CVars.SetCallback(&CVarVidAnisotropic, v.CVarFilterChanged)
//	CVars.SetCallback(&CVarVidFSAAMode, v.CVarFSAAChanged)
//	CVars.SetCallback(&CVarVidFSAA, v.CVarFSAAChanged)
//	CVars.SetCallback(&CVarVidVSync, v.CVarChanged)
//	CVars.SetCallback(&CVarVidDesktopFullscreen, v.CVarChanged)
//	CVars.SetCallback(&CVarVidBorderless, v.CVarChanged)
//
//	Cmds.Add("vid_unlock", v.CmdUnlock, CmdSourceCommand)
//	Cmds.Add("vid_restart", v.CmdRestart, CmdSourceCommand)
//	Cmds.Add("vid_test", v.CmdTest, CmdSourceCommand)
//	Cmds.Add("vid_describecurrentmode", v.CmdDescribeCurrentMode, CmdSourceCommand)
//	Cmds.Add("vid_describemodes", v.CmdDescribeModes, CmdSourceCommand)
//
//	// TODO: CreatePaletteOctree
//
//	_ = os.Setenv("SDL_VIDEO_CENTERED", "center")
//
//	if err := sdl.InitSubSystem(sdl.INIT_VIDEO); err != nil {
//		log.Fatalf("Couldn't init SDL video: %s\n", err)
//	}
//
//	mode, err := sdl.GetDesktopDisplayMode(0)
//	if err != nil {
//		log.Fatalf("Could not get desktop display mode: %s\n", err)
//	}
//	displayWidth := int(mode.W)
//	displayHeight := int(mode.H)
//	displayRefreshRate := int(mode.RefreshRate)
//
//	videoDriver, _ := sdl.GetCurrentVideoDriver()
//	log.Printf("SDL Video Driver: %s\n", videoDriver)
//
//	if Config.OpenConfig("config.cfg") {
//		_ = Config.ReadCVars(readVars)
//		Config.Close()
//	}
//	Config.ReadCVarOverrides(readVars)
//
//	v.initModeList()
//
//	width := int(CVarVidWidth.Value)
//	height := int(CVarVidHeight.Value)
//	refreshRate := int(CVarVidRefreshRate.Value)
//	fullScreen := CVarVidFullScreen.Value != 0
//	Vulkan.WantFullScreenExclusive = CVarVidFullScreen.Value >= 2
//
//	if CmdLine.CheckParam("-current") > 0 {
//		width = displayWidth
//		height = displayHeight
//		refreshRate = displayRefreshRate
//		fullScreen = true
//	} else {
//		p := CmdLine.CheckParam("-width")
//		if p > 0 && p < CmdLine.ArgCount()-1 {
//			width, _ = strconv.Atoi(CmdLine.Arg(p + 1))
//
//			if CmdLine.CheckParam("-height") == 0 {
//				height = width * 3 / 4
//			}
//		}
//
//		p = CmdLine.CheckParam("-height")
//		if p > 0 && p < CmdLine.ArgCount()-1 {
//			height, _ = strconv.Atoi(CmdLine.Arg(p + 1))
//
//			if CmdLine.CheckParam("-width") == 0 {
//				width = height * 4 / 3
//			}
//		}
//
//		p = CmdLine.CheckParam("-refreshrate")
//		if p > 0 && p < CmdLine.ArgCount()-1 {
//			refreshRate, _ = strconv.Atoi(CmdLine.Arg(p + 1))
//		}
//
//		if CmdLine.CheckParam("-window") > 0 || CmdLine.CheckParam("-w") > 0 {
//			fullScreen = false
//		} else if CmdLine.CheckParam("-fullscreen") > 0 || CmdLine.CheckParam("-f") > 0 {
//			fullScreen = true
//		}
//	}
//
//	if !v.ValidMode(width, height, refreshRate, fullScreen) {
//		width = int(CVarVidWidth.Value)
//		height = int(CVarVidHeight.Value)
//		refreshRate = int(CVarVidRefreshRate.Value)
//		fullScreen = CVarVidFullScreen.Value != 0
//	}
//
//	if !v.ValidMode(width, height, refreshRate, fullScreen) {
//		width = 640
//		height = 480
//		refreshRate = displayRefreshRate
//		fullScreen = false
//	}
//
//	v.initialized = true
//
//	// TODO: Initialize colormap and fullbright from host_colormap
//
//	v.SetMode(width, height, refreshRate, fullScreen)
//
//	// TODO: Set window Icon
//
//	log.Println("\nVulkan Initialization")
//	_ = sdl.VulkanLoadLibrary("")
//	v.initInstance()
//	v.initDevice()
//	v.initCommandBuffers()
//
//	Vulkan.StagingBufferSize = InitialStagingBufferSizeKB * 1024
//	// TODO: R_InitStagingBuffers()
//	// TODO: R_CreateDescriptorSetLayouts()
//	// TODO: R_CreateDescriptorPool()
//	// TODO: R_InitGPUBuffers()
//	// TODO: R_InitMeshHeap()
//	TexMgr.InitHeap()
//	// TODO: R_InitSamplers()
//	// TODO: R_CreatePipelineLayouts()
//	v.CreatePaletteOctreeBuffers()
//
//	v.gammaInit()
//	v.MenuInit()
//
//	// Current vid settings should override config file settings,
//	// so we have to lock the vid mode from now until after all config files are read
//	v.locked = true
//}

//func (v *SDLVideo) Restart(setMode bool) {
//	if !v.initialized {
//		return
//	}
//
//	v.SynchronizedEndRenderingTask()
//
//	width := int(CVarVidWidth.Value)
//	height := int(CVarVidHeight.Value)
//	refreshRate := int(CVarVidRefreshRate.Value)
//	fullScreen := CVarVidFullScreen.Value != 0
//	Vulkan.WantFullScreenExclusive = CVarVidFullScreen.Value >= 2
//
//	// Validate new mode
//	if setMode && !v.ValidMode(width, height, refreshRate, fullScreen) {
//		fullScreenText := "windowed"
//		if fullScreen {
//			fullScreenText = "fullscreen"
//		}
//		log.Printf("%dx%d %dHz %s is not a valid mode\n", width, height, refreshRate, fullScreenText)
//		return
//	}
//
//	// TODO: Set screen initialized false
//	v.WaitForDeviceIdle()
//	v.DestroyRenderResources()
//
//	if setMode {
//		v.SetMode(width, height, refreshRate, fullScreen)
//	}
//
//	v.CreateRenderResources()
//
//	// conwidth and conheight recalc
//	// TODO: need scr_conwidth
//
//	v.SyncCVars()
//
//	// update mouse grab
//	// TODO: input mouse grab
//
//	// TODO: R_InitSamplers()
//
//	// TODO: SCR_UpdateRelativeScale
//
//	// TODO: set screen initialized true
//}
//
//func (v *SDLVideo) CmdRestart() {
//	if v.locked || !v.changed {
//		return
//	}
//
//	v.Restart(true)
//}

func (v *SDLVideo) Toggle() {
	// TODO: Clear sound buffer

	var flags uint32

	if !v.GetFullscreen() {
		flags = sdl.WINDOW_FULLSCREEN
		if CVarVidDesktopFullscreen.Value != 0 {
			flags = sdl.WINDOW_FULLSCREEN_DESKTOP
		}
	}

	err := v.drawContext.SetFullscreen(flags)
	if err == nil {
		modeState := ModeStateWindowed
		if v.GetFullscreen() {
			modeState = ModeStateFullscreen
		}
		v.modeState = modeState

		v.SyncCVars()

		// update mouse grab
		// TODO: input mouse grab
	}
}

func (v *SDLVideo) SyncCVars() {
	if v.drawContext != nil {
		if !v.GetDesktopFullscreen() {
			CVars.SetValueQuick(&CVarVidWidth, float64(v.GetCurrentWidth()))
			CVars.SetValueQuick(&CVarVidHeight, float64(v.GetCurrentHeight()))
		}
		CVars.SetValueQuick(&CVarVidRefreshRate, float64(v.GetCurrentRefreshRate()))

		fullScreenVal := "0"
		if v.GetFullscreen() && Vulkan.WantFullScreenExclusive {
			fullScreenVal = "2"
		} else if v.GetFullscreen() {
			fullScreenVal = "1"
		}
		CVars.SetQuick(&CVarVidFullScreen, fullScreenVal)
	}

	v.changed = false
}

// Menus

func (v *SDLVideo) menuInit() {
	uniqueMenuModes := make(map[MenuMode]struct{})

	for _, mode := range v.modes {
		w := mode.Width
		h := mode.Height

		uniqueMenuModes[MenuMode{
			Width:  w,
			Height: h,
		}] = struct{}{}
	}

	v.menuModes = make([]MenuMode, 0, len(uniqueMenuModes))

	for menuMode := range uniqueMenuModes {
		v.menuModes = append(v.menuModes, menuMode)
	}
}

// Vulkan
//
//func (v *SDLVideo) createPaletteOctreeBuffers(colors []uint32, nodes []PaletteOctreeNode) {
//	colorsSize := len(colors) * int(unsafe.Sizeof(uint32(0)))
//	nodesSize := len(nodes) * int(unsafe.Sizeof(PaletteOctreeNode{}))
//
//	bufferCreateInfos := [2]BufferCreateInfo{
//		{
//			Buffer: &v.PaletteColorsBuffer,
//			Size: colorsSize,
//			Usage: core1_0.BufferUsageUniformTexelBuffer|core1_0.BufferUsageTransferDst
//			Name: "Palette colors",
//		},
//		{
//			Buffer: &v.PaletteOctreeBuffer,
//			Size: nodesSize,
//			Usage: core1_0.BufferUsageUniformBuffer|core1_0.BufferUsageTransferDst,
//			Name: "Palette octree",
//		},
//	}
//	//
//}
//
//func (v *SDLVideo) InitInstance() {
//	global, _ := core.CreateDriverFromProcAddr(sdl.VulkanGetVkGetInstanceProcAddr())
//	instanceExtensions, _, err := global.AvailableExtensions()
//	if err != nil {
//		log.Fatalf("AvailableExtensiosn failed: %s", err)
//	}
//
//	Vulkan.GetSurfaceCapabilities2 = nil
//	Vulkan.GetPhysicalDeviceProperties2 = nil
//	Vulkan.DebugUtils = nil
//	Vulkan.Vulkan1_1 = nil
//
//	sdlExtensions := v.drawContext.VulkanGetInstanceExtensions()
//
//	desiredVersion := common.Vulkan1_0
//	if global.Loader().Version().IsAtLeast(common.Vulkan1_1) {
//		desiredVersion = common.Vulkan1_1
//	}
//
//	instanceCreateInfo := core1_0.InstanceCreateInfo{
//		ApplicationName: "vkngQuake",
//		ApplicationVersion: 1,
//		EngineName: "vkngQuake",
//		EngineVersion: 1,
//		APIVersion: desiredVersion,
//		EnabledExtensionNames: extensions,
//	}
//}
//
