package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = posWidth + offWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	file, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	idx.size = uint64(file.Size())
	err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes))
	if err != nil {
		return nil, err
	}

	idx.mmap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return idx, nil
}

// Read takes an offset index and returns the associated records position in the store
func (i *index) Read(index int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	// Set the out to the last offset position in the index.
	if index == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(index)
	}
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

// Write appends the given offset and position to the index
func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += entWidth
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}

func (i *index) Close() error {
	// Makes sure memory-mapped file has synced data to the persisted file
	err := i.mmap.Sync(gommap.MS_SYNC)
	if err != nil {
		return err
	}

	// Makes sure that the persisted file has flushed its contents to stable storage.
	err = i.file.Sync()
	if err != nil {
		return err
	}

	// Truncates the persisted file to the amount of data that's used in the file.
	err = i.file.Truncate(int64(i.size))
	if err != nil {
		return err
	}

	// Closes the file safely.
	return i.file.Close()
}
