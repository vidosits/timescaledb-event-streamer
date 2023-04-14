package offsetstorages

import (
	"encoding/binary"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/go-errors/errors"
	"github.com/noctarius/timescaledb-event-streamer/spi/offset"
	"os"
	"path/filepath"
)

type fileOffsetStorage struct {
	path    string
	offsets map[string]*offset.Offset
}

func NewFileOffsetStorage(path string) (offset.Storage, error) {
	directory := filepath.Dir(path)
	fi, err := os.Stat(directory)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(directory, 0777); err != nil {
				return nil, errors.Wrap(err, 0)
			}
		} else {
			return nil, errors.Wrap(err, 0)
		}
	}

	if !fi.IsDir() {
		return nil, errors.Errorf(
			"path '%s' cannot be created since the parent-path '%s' is no directory", path, directory,
		)
	}

	fi, err = os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, 0)
		}
	}

	if fi.IsDir() {
		return nil, errors.Errorf("path '%s' exists already but is not a file", path)
	}

	return &fileOffsetStorage{
		path:    path,
		offsets: make(map[string]*offset.Offset, 0),
	}, nil
}

func (f *fileOffsetStorage) Start() error {
	return f.Load()
}

func (f *fileOffsetStorage) Stop() error {
	return nil
}

func (f *fileOffsetStorage) Save() error {
	writer, err := ioutils.NewAtomicFileWriter(f.path, 0777)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	defer writer.Close()

	buffer := make([]byte, 4)
	writeUint32 := func(val uint32) (int, error) {
		binary.BigEndian.PutUint32(buffer[0:4], val)
		return writer.Write(buffer[0:4])
	}

	writeOffsetWithLength := func(val *offset.Offset) (int, error) {
		data, err := val.MarshalBinary()
		if err != nil {
			return 0, errors.Wrap(err, 0)
		}

		if _, err := writeUint32(uint32(len(data))); err != nil {
			return 0, errors.Wrap(err, 0)
		}

		if _, err := writer.Write(data); err != nil {
			return 0, errors.Wrap(err, 0)
		}
		return 4 + len(data), nil
	}

	writeStringWithLength := func(val string) (int, error) {
		byteString := []byte(val)
		if _, err := writeUint32(uint32(len(byteString))); err != nil {
			return 0, errors.Wrap(err, 0)
		}

		if _, err := writer.Write(byteString); err != nil {
			return 0, errors.Wrap(err, 0)
		}
		return 4 + len(byteString), nil
	}

	numOfOffsets := uint32(len(f.offsets))
	if _, err := writeUint32(numOfOffsets); err != nil {
		return errors.Wrap(err, 0)
	}

	for key, value := range f.offsets {
		if _, err := writeStringWithLength(key); err != nil {
			return errors.Wrap(err, 0)
		}
		if _, err := writeOffsetWithLength(value); err != nil {
			return errors.Wrap(err, 0)
		}
	}

	return nil
}

func (f *fileOffsetStorage) Load() error {
	fi, err := os.Stat(f.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, 0)
		} else {
			// Reset internal map
			f.offsets = make(map[string]*offset.Offset, 0)
			return nil
		}
	}

	if fi.IsDir() {
		return errors.Errorf("path '%s' exists already but is not a file", f.path)
	}

	if fi.Size() == 0 {
		// Reset internal map
		f.offsets = make(map[string]*offset.Offset, 0)
		return nil
	}

	file, err := os.Open(f.path)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	buffer := make([]byte, fi.Size())
	if _, err := file.Read(buffer); err != nil {
		return errors.Wrap(err, 0)
	}

	readerOffset := int64(0)
	readUint32 := func() uint32 {
		val := binary.BigEndian.Uint32(buffer[readerOffset : readerOffset+4])
		readerOffset += 4
		return val
	}

	readString := func() string {
		length := readUint32()
		val := string(buffer[readerOffset : readerOffset+int64(length)])
		readerOffset += int64(length)
		return val
	}

	readOffset := func() (*offset.Offset, error) {
		length := readUint32()
		o := &offset.Offset{}
		if err := o.UnmarshalBinary(buffer[readerOffset : readerOffset+int64(length)]); err != nil {
			return nil, err
		}
		readerOffset += int64(length)
		return o, nil
	}

	numOfOffsets := readUint32()
	for i := uint32(0); i < numOfOffsets; i++ {
		key := readString()
		value, err := readOffset()
		if err != nil {
			return errors.Wrap(err, 0)
		}
		f.offsets[key] = value
	}
	return nil
}

func (f *fileOffsetStorage) Get() (map[string]*offset.Offset, error) {
	return f.offsets, nil
}

func (f *fileOffsetStorage) Set(key string, value *offset.Offset) error {
	f.offsets[key] = value
	return nil
}