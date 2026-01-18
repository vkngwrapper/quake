package main

import (
	"compress/flate"
	"io"
	"log"
	"os"
)

func main() {
	args := os.Args

	if len(args) != 3 {
		log.Fatalln("Usage: deflate [output] [input]")
	}

	outputFile := args[1]
	inputFile := args[2]

	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Could not open input file '%s': %s", inputFile, err)
	}
	defer func() {
		_ = inFile.Close()
	}()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Could not open output file '%s': %s", outputFile, err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	outCompression, err := flate.NewWriter(outFile, flate.BestCompression)
	if err != nil {
		log.Fatalf("Could not begin compression for output file '%s': %s", outputFile, err)
	}
	defer func() {
		_ = outCompression.Close()
	}()

	_, err = io.Copy(outCompression, inFile)
	if err != nil {
		log.Fatalf("Could not copy data to compressed file: %s", err)
	}

	log.Printf("Data compressed to file '%s'.", outputFile)
}
