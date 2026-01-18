package main

import (
	"bytes"
	"compress/flate"
	"embed"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/vkngwrapper/quake/crc"
)

//go:embed build/embedded
var embeddedFiles embed.FS

const GameName = "id1"
const MaxFilesInPack = 2048
const PakFileSize = 64
const Pak0FileCount = 339
const (
	Pak0CrcV100 uint16 = 13900
	Pak0CrcV101        = 62751
	Pak0CrcV106        = 32981
)

type BytesFile struct {
	bytes.Reader
}

func (f *BytesFile) Close() error {
	return nil
}

var CVarRegistered = CVar{
	Name:      "registered",
	StringVal: "1",
	Flags:     CVarFlagROM,
}

var CVarCmdline = CVar{
	Name:  "cmdline",
	Flags: CVarFlagROM,
}

type PackFile struct {
	name    string
	filePos int
	fileLen int
}

type GamePack struct {
	fileName string
	handle   io.ReadSeekCloser
	files    []PackFile
}

type SearchPath struct {
	pathId   int
	fileName string
	pack     *GamePack
	dir      string
	next     *SearchPath
}

type FileSystem struct {
	baseDir   string
	gameDir   string
	gameNames string

	modified            bool
	vkQuakePakExtracted *BytesFile

	searchPaths     *SearchPath
	baseSearchPaths *SearchPath
}

var Files *FileSystem = &FileSystem{}

func (f *FileSystem) Init() {
	CVars.Register(&CVarRegistered)
	CVars.Register(&CVarCmdline)
	Cmds.Add("path", f.CmdPath, CmdSourceCommand)
	Cmds.Add("game", f.CmdGame, CmdSourceCommand)

	baseDirArgIndex := CmdLine.CheckParam("-basedir")
	if baseDirArgIndex > 0 && baseDirArgIndex < CmdLine.ArgCount()-1 {
		f.baseDir = CmdLine.Arg(baseDirArgIndex + 1)
	} else {
		f.baseDir = HostParams.baseDir
	}

	if f.baseDir == "" {
		log.Fatalln("Bad argument to -basedir")
	}
	if f.baseDir[len(f.baseDir)-1] == os.PathSeparator {
		f.baseDir = f.baseDir[:len(f.baseDir)-1]
	}

	baseGameArgIndex := CmdLine.CheckParamNext(baseDirArgIndex, "-basegame")
	if baseGameArgIndex > 0 {
		f.modified = true
		for ; baseGameArgIndex > 0 && baseGameArgIndex < CmdLine.ArgCount()-1; baseGameArgIndex = CmdLine.CheckParamNext(baseGameArgIndex, "-basegame") {
			dir := CmdLine.Arg(baseGameArgIndex)

			if ModForbiddenChars(dir) {
				log.Fatalln("gamedir should be a single directory name, not a path")
			}
			if dir != "" {
				f.AddGameDirectory(dir)
			}
		}
	} else {
		f.AddGameDirectory(GameName)
	}

	f.baseSearchPaths = f.searchPaths
	f.ResetGameDirectories("")

	if CmdLine.CheckParam("-rogue") > 0 {
		f.AddGameDirectory("rogue")
	} else if CmdLine.CheckParam("-hipnotic") > 0 {
		f.AddGameDirectory("hipnotic")
	} else if CmdLine.CheckParam("-quoth") > 0 {
		f.AddGameDirectory("quoth")
	}

	gameArgIndex := CmdLine.CheckParamNext(0, "-game")
	for ; gameArgIndex > 0 && gameArgIndex < CmdLine.ArgCount()-1; gameArgIndex = CmdLine.CheckParamNext(gameArgIndex, "-game") {
		dir := CmdLine.Arg(gameArgIndex + 1)
		if ModForbiddenChars(dir) {
			log.Fatalln("gamedir should be a single directory name, not a path")
		}
		f.modified = true
		if dir != "" {
			f.AddGameDirectory(dir)
		}
	}

	f.CheckRegistered()
}

func (f *FileSystem) loadEmbeddedPak() {
	file, err := embeddedFiles.Open("vkQuake.pak")
	if err != nil {
		log.Fatalf("Error extracting embedded pak: %s\n", err)
	}
	defer func() {
		_ = file.Close()
	}()

	reader := flate.NewReader(file)
	defer func() {
		_ = reader.Close()
	}()

	pakBytes, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Error extracting embedded pak: %s\n", err)
	}

	f.vkQuakePakExtracted = &BytesFile{*bytes.NewReader(pakBytes)}
}

func (f *FileSystem) addPath(pathId int, dir string) {
	search := &SearchPath{
		pathId:   pathId,
		fileName: f.gameDir,
		dir:      dir,
		next:     f.searchPaths,
	}
	f.searchPaths = search

	for pakIndex := 0; ; pakIndex++ {
		pakFile := path.Join(f.gameDir, fmt.Sprintf("pak%d.pak", pakIndex))
		file, err := os.Open(pakFile)
		if err != nil {
			return
		}
		pak := f.LoadPackFile(pakFile, file)
		if pak != nil {
			search = &SearchPath{
				pathId: pathId,
				pack:   pak,
				dir:    dir,
				next:   f.searchPaths,
			}
			f.searchPaths = search
		}

		if pakIndex == 0 && pathId == 1 && !FitzMode {
			if f.vkQuakePakExtracted == nil {
				f.loadEmbeddedPak()
			}
			wasModified := f.modified
			pak = f.LoadPackFile("vkQuake.pak", f.vkQuakePakExtracted)
			search = &SearchPath{
				pathId: pathId,
				pack:   pak,
				dir:    dir,
				next:   f.searchPaths,
			}
			f.searchPaths = search
			f.modified = wasModified
		}

		if pak == nil {
			return
		}
	}
}

func (f *FileSystem) AddGameDirectory(dir string) {
	if f.gameNames != "" {
		f.gameNames += ";"
	}
	f.gameNames += dir

	if dir == "rogue" {
		CmdLine.SetRogue()
	}
	if dir == "hipnotic" || dir == "quoth" {
		CmdLine.SetHipnotic()
	}

	f.gameDir = path.Join(f.baseDir, dir)

	var pathId int
	if f.searchPaths != nil {
		pathId = f.searchPaths.pathId << 1
	} else {
		pathId = 1
	}

	f.addPath(pathId, dir)
	if HostParams.userDir != HostParams.baseDir {
		f.gameDir = path.Join(HostParams.userDir, dir)
		err := os.MkdirAll(f.gameDir, 0777)
		if err != nil {
			log.Fatalf("Unable to create directory %s: %s\n", f.gameDir, err.Error())
		}
		f.addPath(pathId, dir)
	}
}

func (f *FileSystem) LoadPackFile(path string, file io.ReadSeekCloser) *GamePack {
	var header struct {
		id        [4]byte
		dirOffset int32
		dirSize   int32
	}
	err := binary.Read(file, binary.LittleEndian, &header)
	if err != nil || header.id[0] != 'P' || header.id[1] != 'A' || header.id[2] != 'C' || header.id[3] != 'K' {
		log.Fatalf("%s is not a packfile", path)
	}
	if header.dirOffset < 0 || header.dirSize < 0 {
		log.Fatalf("Invalid packfile %s (dirSize: %d, dirOffset: %d)", path, header.dirSize, header.dirOffset)
	}

	numFiles := header.dirSize / int32(PakFileSize)
	if numFiles < 1 {
		log.Printf("WARNING: %s has no files, ignored\n", path)
		_ = file.Close()
		return nil
	}
	if numFiles > MaxFilesInPack {
		log.Fatalf("%s has %d files", path, numFiles)
	}
	if numFiles != Pak0FileCount {
		f.modified = true
	}

	fileData := make([]PackFile, numFiles)
	fileDataBytes := make([]byte, header.dirSize)
	_, _ = file.Seek(int64(header.dirOffset), 0)
	_ = binary.Read(file, binary.LittleEndian, &fileDataBytes)

	var crcValue uint16
	crc.Init(&crcValue)
	for _, b := range fileDataBytes {
		crc.ProcessByte(&crcValue, b)
	}

	if crcValue != Pak0CrcV106 && crcValue != Pak0CrcV101 && crcValue != Pak0CrcV100 {
		f.modified = true
	}

	for fileDataIndex := range fileData {
		startByteIndex := PakFileSize * fileDataIndex
		fileData[fileDataIndex].name = string(fileDataBytes[startByteIndex : startByteIndex+56])
		fileData[fileDataIndex].filePos = int(binary.LittleEndian.Uint32(fileDataBytes[startByteIndex+56:]))
		fileData[fileDataIndex].fileLen = int(binary.LittleEndian.Uint32(fileDataBytes[startByteIndex+60:]))
	}

	return &GamePack{
		fileName: path,
		handle:   file,
		files:    fileData,
	}
}

func (f *FileSystem) ResetGameDirectories(newDirs string) {
	for f.searchPaths != f.baseSearchPaths {
		if f.searchPaths.pack != nil {
			_ = f.searchPaths.pack.handle.Close()
		}
		f.searchPaths = f.searchPaths.next
	}

	CmdLine.SetStandardQuake()
	f.gameNames = ""

	dir := f.baseDir
	if HostParams.userDir != HostParams.baseDir {
		dir = HostParams.userDir
	}

	f.gameDir = path.Join(dir, GameName)

	pathSegments := strings.Split(newDirs, ";")

	for pathIndex, pathSeg := range pathSegments {
		if pathSeg == GameName {
			// The base game was never actually unloaded
			continue
		}

		// Make sure this isn't a duplicate
		var firstMatch int
		for firstMatch = 0; firstMatch < pathIndex; firstMatch++ {
			if pathSegments[firstMatch] == pathSeg {
				break
			}
		}

		if firstMatch == pathIndex {
			f.AddGameDirectory(pathSeg)
		}
	}
}

func (f *FileSystem) FileExists(fileName string) bool {
	size, _, _ := f.findFile(fileName, false)
	return size > 0
}

func (f *FileSystem) OpenFile(fileName string) (size int, file BoundedReader, pathId int) {
	return f.findFile(fileName, true)
}

func (f *FileSystem) LoadFile(fileName string) (data []byte, pathId int) {
	size, file, pathId := f.OpenFile(fileName)
	if size <= 0 {
		return
	}

	data = make([]byte, size)
	_, err := io.ReadFull(&file, data)
	if err != nil {
		return nil, pathId
	}

	_ = file.Close()
	return
}

func (f *FileSystem) findFile(fileName string, openFile bool) (size int, file BoundedReader, pathId int) {
	isConfig := fileName == "config.cfg"

	for search := f.searchPaths; search != nil; search = search.next {
		if search.pack != nil {
			for _, entry := range search.pack.files {
				if entry.name != fileName {
					continue
				}

				pathId = search.pathId
				size = entry.fileLen

				if openFile {
					file = BoundedReaderFromPackFile(entry, search.pack)
				}

				return
			}
		} else if CVarRegistered.Value == 0 && (strings.Contains(fileName, "/") || strings.Contains(fileName, "\\")) {
			continue
		} else {
			var netPath string
			var found bool
			var fileInfo os.FileInfo
			var err error
			if isConfig {
				netPath = path.Join(search.fileName, "vkQuake.cfg")
				fileInfo, err = os.Stat(netPath)
				found = err == nil && fileInfo.Mode().IsRegular()
			}

			if !found {
				netPath = path.Join(search.fileName, fileName)
				fileInfo, err = os.Stat(netPath)
				if err != nil || !fileInfo.Mode().IsRegular() {
					continue
				}
			}

			pathId = search.pathId
			size = int(fileInfo.Size())

			if openFile {
				osFile, _ := os.Open(fileName)
				file = BoundedReaderFromOSFile(osFile, size)
			}

			return
		}
	}

	// TODO: Developer mode

	return -1, BoundedReader{}, -1
}

func (f *FileSystem) CheckRegistered() {
	size, file, _ := Files.OpenFile("gfx/pop.lmp")
	if size <= 0 {
		CVars.SetROM("registered", "0")
		log.Println("Playing shareware version.")
		if f.modified {
			log.Fatalf("You must have the registered version to use modified games.\n\n"+
				"Basedir is: %s\n\n"+
				"Check that this has an %s subdirectory containing pak0.pak and pak1.pak, "+
				"or use the -basedir command-line to specify another directory.",
				f.baseDir, GameName)
		}

		return
	}

	_ = file.Close()

	CVars.SetROM("cmdline", strings.TrimSpace(CmdLine.CmdLine()))
	CVars.SetROM("registered", "1")
	log.Println("Playing registered version.")
}

func (f *FileSystem) GameNames(full bool) string {
	if !full {
		return f.gameNames
	}

	if f.gameNames == "" {
		return GameName
	}

	return fmt.Sprintf("%s;%s", GameName, f.gameNames)
}

func ModForbiddenChars(path string) bool {
	return path == "" || path == "." || strings.Contains(path, "..") ||
		strings.Contains(path, string(os.PathSeparator)) || strings.Contains(path, ":") ||
		strings.Contains(path, "\"") || strings.Contains(path, ";")
}

func (f *FileSystem) CmdPath() {
	log.Println("Current search path:")

	for s := f.searchPaths; s != nil; s = s.next {
		if s.pack != nil {
			log.Printf("%s (%d files)\n", s.pack.fileName, len(s.pack.files))
		} else {
			log.Printf("%s\n", s.fileName)
		}
	}
}

func (f *FileSystem) CmdGame() {
	if Cmds.ArgCount() < 2 {
		log.Printf("\"game\" is \"%s\"\n", f.GameNames(true))
		return
	}

	if CVarRegistered.Value == 0 {
		log.Println("You must have the registered version to use modified games")
		return
	}

	pathSegs := make([]string, 1, Cmds.ArgCount())
	pathSegs[0] = GameName

	for pri := 0; pri <= 1; pri++ {
		for arg := 1; arg < Cmds.ArgCount(); arg++ {
			argVal := Cmds.Arg(arg)

			// Games with dashes in front of them are higher priority so process them when
			// pri==0 and don't when pri==1
			if (pri == 0) != (argVal[0] == '-') {
				continue
			}

			// remove dash to normalize args
			if pri == 0 {
				argVal = argVal[1:]
			}

			if ModForbiddenChars(argVal) {
				log.Println("gamedir should be a single directory name, not a path")
				return
			}

			if argVal == GameName {
				// base game is always loaded, don't need to specify it
				continue
			}

			pathSegs = append(pathSegs, argVal)
		}
	}

	games := strings.Join(pathSegs, ";")
	if games == f.GameNames(true) {
		log.Printf("\"game\" is already \"%s\"\n", games)
		return
	}

	f.modified = true

	// Kill the server
	// TODO: Disconnect and shut down

	// TODO: Center print clear

	// TODO: Write config file

	f.ResetGameDirectories(games)

	//TODO: Reset mods and clear sky
	if !IsDedicated {
		//TODO: New game
	}
	// TODO: Reset and rebuild and clear

	fmt.Printf("\"game\" changed to \"%s\"\n", f.GameNames(true))

	// TODO: vid lock
	Cmds.AddText("exec quake.rc\n")
	Cmds.AddText("vid_unlock\n")
}
