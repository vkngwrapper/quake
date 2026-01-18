package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type BoundedReader struct {
	file   io.ReadSeekCloser
	start  int
	length int
	pos    int
}

func BoundedReaderFromOSFile(file *os.File, size int) BoundedReader {
	return BoundedReader{
		file:   file,
		start:  0,
		length: size,
		pos:    0,
	}
}

func BoundedReaderFromPackFile(packFile PackFile, pack *GamePack) BoundedReader {
	_, _ = pack.handle.Seek(int64(packFile.filePos), io.SeekStart)
	return BoundedReader{
		file:   pack.handle,
		start:  packFile.filePos,
		length: packFile.fileLen,
		pos:    0,
	}
}

func (f *BoundedReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, errors.New("no buffer sent to read")
	}

	if f.pos >= f.length {
		return 0, io.EOF
	}

	byteSize := len(p)

	if byteSize > f.length-f.pos {
		byteSize = f.length - f.pos
	}

	bytesRead, err := f.file.Read(p[:byteSize])
	if err != nil {
		return 0, err
	}

	f.pos += bytesRead

	return bytesRead, nil
}

func (f *BoundedReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		break
	case io.SeekCurrent:
		offset += int64(f.pos)
		break
	case io.SeekEnd:
		offset = int64(f.length) + offset
		break
	default:
		return -1, fmt.Errorf("invalid whence value %d", whence)
	}

	if offset > int64(f.length) {
		offset = int64(f.length)
	}

	ret, err := f.file.Seek(int64(f.start)+offset, io.SeekStart)
	if err != nil || ret < 0 {
		return ret, err
	}

	f.pos = int(offset)
	return 0, nil
}

func (f *BoundedReader) Close() error {
	return f.file.Close()
}

func (f *BoundedReader) Tell() int {
	return f.pos
}

func (f *BoundedReader) Size() int {
	return f.length
}
