package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type ConfigFile struct {
	configFile *BoundedReader
}

var Config = &ConfigFile{}

func (f *ConfigFile) ReadCVars(vars []string) error {
	if f.configFile == nil || len(vars) < 1 {
		return nil
	}

	scanner := bufio.NewScanner(Config.configFile)
	scanner.Split(bufio.ScanLines)

	var foundVars int

	for scanner.Scan() {
		buff := scanner.Text()
		buff = strings.ReplaceAll(buff, "\t", " ")

		// Find first space
		firstSpaceIndex := strings.Index(buff, " ")
		if firstSpaceIndex < 1 {
			continue
		}

		varName := buff[:firstSpaceIndex]
		var foundVarName bool
		for _, varStr := range vars {
			if varName == varStr {
				foundVarName = true
				break
			}
		}

		if !foundVarName {
			continue
		}

		if buff[len(buff)-1] != '"' {
			// Line must end with quotation mark
			continue
		}

		// Value is first quote until the quote the value ends with
		firstQuoteIndex := strings.Index(buff, "\"")
		if firstQuoteIndex == len(buff)-1 {
			// there are supposed to be two quotes
			continue
		}

		quotedVal := buff[firstQuoteIndex : len(buff)-1]

		CVars.Set(varName, quotedVal)
		foundVars++

		if foundVars == len(vars) {
			break
		}
	}

	_, err := Config.configFile.Seek(0, io.SeekStart)
	return err
}

func (f *ConfigFile) ReadCVarOverrides(vars []string) {
	if len(vars) < 1 {
		return
	}

	for _, varName := range vars {
		index := CmdLine.CheckParam("+" + varName)
		if index > 0 && index < CmdLine.ArgCount()-1 {
			argVal := CmdLine.Arg(index)
			if argVal[0] != '-' && argVal[0] != '+' {
				CVars.Set(varName, argVal)
			}
		}
	}
}

func (f *ConfigFile) Close() {
	if f.configFile != nil {
		_ = f.configFile.Close()
		f.configFile = nil
	}
}

func (f *ConfigFile) OpenConfig(configName string) bool {
	f.Close()

	var length int

	if MultiUser {
		prefPath := sdl.GetPrefPath("", "vkngQuake")
		path := fmt.Sprintf("%s/config.cfg", prefPath)

		info, err := os.Stat(path)
		if err == nil {
			length = int(info.Size())
			file, err := os.Open(path)

			if err == nil && file != nil {
				reader := BoundedReaderFromOSFile(file, length)
				f.configFile = &reader
				return true
			}
		}
	}

	length, file, _ := Files.OpenFile(configName)
	if length <= 0 {
		return false
	}

	f.configFile = &file

	return true
}
