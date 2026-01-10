package main

import (
	"log"

	"golang.org/x/sys/cpu"
)

type ParseOverflowBehavior int

const (
	ParseOverflowFail ParseOverflowBehavior = iota
	ParseOverflowTruncate
)

const MaxParseTokenSize int = 4096

func InitCommon() {
	if cpu.IsBigEndian {
		log.Fatalln("Unsupported endianism. Only little endian is supported")
	}

	if CmdLine.CheckParam("-fitz") > 0 {
		FitzMode = true
	}
	if CmdLine.CheckParam("-validation") > 0 {
		// TODO: Vulkan validation
	}
	if CmdLine.CheckParam("-multiuser") > 0 {
		MultiUser = true
	}
	// TODO: Null Entity setup
}

func ParseToken(data []rune) string {
	return ParseTokenWithOverflowBehavior(data, ParseOverflowFail)
}

func ParseTokenWithOverflowBehavior(data []rune, overflow ParseOverflowBehavior) string {
	var parsedToken [MaxParseTokenSize]rune
	var parsedTokenlen int

	if len(data) == 0 {
		return ""
	}
	dataIndex := 0

skipWhitespace:
	for dataIndex < len(data) && data[dataIndex] <= ' ' {
		dataIndex++
	}

	if dataIndex >= len(data) {
		return ""
	}

	r := data[dataIndex]
	if r == '/' && data[dataIndex+1] == '/' {
		// Single line comment
		for dataIndex < len(data) && data[dataIndex] != '\n' {
			dataIndex++
		}
		goto skipWhitespace
	}

	if r == '/' && data[dataIndex] == '*' {
		dataIndex += 2
		for dataIndex < len(data)-1 && (data[dataIndex] != '*' || data[dataIndex+1] != '/') {
			dataIndex++
		}
		if dataIndex < len(data) {
			dataIndex += 2
		}
		goto skipWhitespace
	}

	// Handle quoted string
	if r == '"' {
		dataIndex++
		for {
			if dataIndex < len(data) {
				r = data[dataIndex]
				dataIndex++
			} else {
				return ""
			}

			if r == '"' {
				return string(parsedToken[:parsedTokenlen])
			}

			if parsedTokenlen < MaxParseTokenSize {
				parsedToken[parsedTokenlen] = r
				parsedTokenlen++
			} else if overflow == ParseOverflowFail {
				return ""
			}
		}
	}

	// Parse single characters
	if r == '{' || r == '}' || r == '(' || r == ')' || r == '\'' || r == ':' {
		if parsedTokenlen < MaxParseTokenSize {
			parsedToken[parsedTokenlen] = r
			parsedTokenlen++
		} else if overflow == ParseOverflowFail {
			return ""
		}
		return string(parsedToken[:parsedTokenlen])
	}

	for r > 32 {
		if parsedTokenlen < MaxParseTokenSize {
			parsedToken[parsedTokenlen] = r
			parsedTokenlen++
		} else if overflow == ParseOverflowFail {
			return ""
		}
		dataIndex++
		r = data[dataIndex]

		if r == '{' || r == '}' || r == '(' || r == ')' || r == '\'' {
			break
		}
	}
	return string(parsedToken[:parsedTokenlen])
}
