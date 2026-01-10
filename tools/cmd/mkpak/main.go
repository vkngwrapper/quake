package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type PackFile struct {
	Name    [56]rune
	FilePos int32
	FileLen int32
}

func writeBytes(w io.Writer, bytes []byte) error {
	index := 0
	for index < len(bytes) {
		l, err := w.Write(bytes[index:])
		if err != nil {
			return err
		}
		index += l
	}

	return nil
}

func writeHeader(w io.Writer, dirOffset int32, dirSize int32) error {
	err := writeBytes(w, []byte{'P', 'A', 'C', 'K'})
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, dirOffset)
	if err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, dirSize)
}

func main() {
	args := os.Args
	if len(args) < 4 {
		log.Fatalln("Usage: mkpak [output.pak] [root dir for files] [toc file] [depfile (optional)]")
	}

	outputPakPath := args[1]
	rootDir := args[2]
	tocFilePath := args[3]
	var depFilePath string

	tocFile, err := os.Open(tocFilePath)
	if err != nil {
		log.Fatalf("Could not open toc_file file '%s': %s\n", tocFilePath, err)
	}
	defer func() {
		_ = tocFile.Close()
	}()

	var depFile *os.File
	var depFileWriter *bufio.Writer

	if len(args) > 4 {
		depFilePath = args[4]
		depFile, err = os.Create(depFilePath)
		if err != nil {
			log.Fatalf("Could not open dep_file file '%s': %s", depFilePath, err)
		}

		depFileWriter = bufio.NewWriter(depFile)

		defer func() {
			_ = depFileWriter.Flush()
			_ = depFile.Close()
		}()

		_, err = fmt.Fprintf(depFileWriter, "%s: %s", outputPakPath, tocFilePath)
		if err != nil {
			log.Fatalf("Could not write to dep_file file '%s': %s", depFilePath, err)
		}
	}

	tocScanner := bufio.NewScanner(tocFile)
	tocScanner.Split(bufio.ScanLines)

	var tocFiles []string
	for tocScanner.Scan() {
		text := strings.TrimSpace(tocScanner.Text())
		if text == "" {
			continue
		}

		tocFiles = append(tocFiles, text)
	}

	if tocScanner.Err() != nil {
		log.Fatalf("Error while reading toc_file '%s': %s\n", tocFilePath, tocScanner.Err())
	}

	dirOffset := int32(12)
	dirSize := int32(64 * len(tocFiles))
	fileOffset := dirOffset + dirSize

	outFile, err := os.Create(outputPakPath)
	if err != nil {
		log.Fatalf("Error while opening output file '%s': %s\n", outputPakPath, err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	err = writeHeader(outFile, dirOffset, dirSize)
	if err != nil {
		log.Fatalf("Error writing output.pak '%s': %s\n", outputPakPath, err)
	}

	for fileIndex, tocEntry := range tocFiles {
		tocEntryPath := path.Join(rootDir, tocEntry)

		if depFileWriter != nil {
			_, err = fmt.Fprintf(depFileWriter, " %s", tocEntryPath)
			if err != nil {
				log.Fatalf("Error while writing to dep file '%s': %s", depFilePath, err)
			}
		}

		inputFileBytes, err := os.ReadFile(tocEntryPath)
		if err != nil {
			log.Fatalf("Error while opening input file '%s': %s", tocEntryPath, err)
		}

		var name [56]rune
		tocEntryRunes := []rune(tocEntry)
		copy(name[:], tocEntryRunes)
		if len(tocEntryRunes) < 56 {
			name[len(tocEntryRunes)] = 0
		}

		fileLen := int32(len(inputFileBytes))
		packFile := PackFile{
			Name:    name,
			FilePos: fileOffset,
			FileLen: fileLen,
		}

		_, err = outFile.Seek(int64(dirOffset)+64*int64(fileIndex), 0)
		if err != nil {
			log.Fatalf("Error while seeking within output file '%s': %s", outputPakPath, err)
		}
		err = binary.Write(outFile, binary.LittleEndian, packFile)
		if err != nil {
			log.Fatalf("Error while writing to output file '%s': %s", outputPakPath, err)
		}

		_, err = outFile.Seek(int64(fileOffset), 0)
		if err != nil {
			log.Fatalf("Error while seeking within output file '%s': %s", outputPakPath, err)
		}
		err = writeBytes(outFile, inputFileBytes)
		if err != nil {
			log.Fatalf("Error while writing to output file '%s': %s", outputPakPath, err)
		}

		fileOffset += fileLen
	}

	if depFileWriter != nil {
		_, err = fmt.Fprintln(depFileWriter)
		if err != nil {
			log.Fatalf("Error while writing to dep file '%s': %s", depFilePath, err)
		}
	}
}
