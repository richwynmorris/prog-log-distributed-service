package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	*os.File
	mu sync.Mutex
	buf *bufio.Writer
	size uint64
}

// newStore finds the current size of the file if it exists and returns a new store object.
func newStore(f *os.File) (*store, error) {
	fileInfo, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fileInfo.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

// Append adds a record to the store by writing the length of bytes to the file.
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	// Mutex ensures only one goroutine has access to append to the store at a time.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the current position of the store by its size.
	pos = s.size

	// write the encoded binary data to the buffered writer instead of the file to reduce system calls on the file.
	err = binary.Write(s.buf, enc, uint64(len(p)))
	if err != nil {
		return 0,0, err
	}

	// Writes the encoded binary data to the file.
	w, err := s.buf.Write(p)
	if err != nil {
		return 0,0, err
	}

	// Update the size of the file with its current file size.
	w += lenWidth
	s.size += uint64(w)

	// return the length of the bytes written, the position of the record in the store and any errors.
	return uint64(w), pos, nil
}

// Read returns the record stored at the given position
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// flushes the buffer in the instance that we're about to read a record that hasn't been written to the file yet.
	err := s.buf.Flush()
	if err != nil {
		return nil, err
	}


	// Find out how many bytes we have to read to get the whole record.
	size := make([]byte, lenWidth)
	_, err = s.File.ReadAt(size, int64(pos))
	if err != nil {
		return nil, err
	}

	// Get the whole record.
	b := make([]byte, enc.Uint64(size))
	_, err = s.File.ReadAt(b, int64(pos+lenWidth))
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ReadAt reads the length of byte from the offset in the store's file.
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.buf.Flush()
	if err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

// Close persists any buffered data before closing the file.
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}

