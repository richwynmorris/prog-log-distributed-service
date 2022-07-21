package log

import (
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
