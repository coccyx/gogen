package outputter

import (
	"io"
	"os"
	"strconv"
	"sync"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

type file struct {
	initialized bool
	file        *os.File
	fileSize    int64
	mutex       *sync.Mutex
}

func (f *file) Send(item *config.OutQueueItem) error {
	if f.initialized == false {
		info, err := os.Stat(item.S.Output.FileName)
		// File doesn't exist, so create
		if os.IsNotExist(err) {
			f.file, err = os.Create(item.S.Output.FileName)
			if err != nil {
				log.Panicf("Cannot create file referenced at '%s' for sample '%s' with error %s", item.S.Output.FileName, item.S.Name, err)
			}
			// We got an unexpected error from Stat
		} else if err != nil {
			log.Panicf("Cannot stat file referenced at '%s' for sample '%s' with error: %s", item.S.Output.FileName, item.S.Name, err)
			// File exists, check the size and open it
		} else {
			f.fileSize = info.Size()
			f.file, err = os.OpenFile(item.S.Output.FileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if err != nil {
				log.Panicf("Cannot open file referenced at '%s' for sample '%s' with error: %s", item.S.Output.FileName, item.S.Name, err)
			}
		}
		f.mutex = &sync.Mutex{}
		f.initialized = true
	}
	// File output is the rare exception, we must be single threaded
	f.mutex.Lock()
	defer f.mutex.Unlock()
	bytes, err := io.Copy(f.file, item.IO.R)

	f.fileSize += bytes
	if f.fileSize >= item.S.Output.MaxBytes {
		log.Infof("Reached %d bytes which exceeds MaxBytes for sample '%s', rotating files", f.fileSize, item.S.Name)
		f.rotate(item)
	}
	return err
}

func (f *file) Close() error {
	if f.initialized {
		if f.file != nil {
			f.file.Close()
			f.file = nil
		}
		f.initialized = false
	}
	return nil
}

func (f *file) rotate(item *config.OutQueueItem) {
	f.file.Close()
	// Remove the oldest file
	_ = os.Remove(item.S.Output.FileName + "." + strconv.Itoa(item.S.Output.BackupFiles))

	// Rotate through all other files and move them to one older
	for i := item.S.Output.BackupFiles; i > 1; i-- {
		err := os.Rename(item.S.Output.FileName+"."+strconv.Itoa(i-1), item.S.Output.FileName+"."+strconv.Itoa(i))
		if err != nil && !os.IsNotExist(err) {
			log.Panicf("Could not rename '%s' to '%s' in sample '%s', err: %s", item.S.Output.FileName+"."+strconv.Itoa(i-1), item.S.Output.FileName+"."+strconv.Itoa(i), item.S.Name, err)
		}
	}

	err := os.Rename(item.S.Output.FileName, item.S.Output.FileName+".1")
	if err != nil && !os.IsNotExist(err) {
		log.Panicf("Could not rename '%s' to '%s' in sample '%s', err: %s", item.S.Output.FileName, item.S.Output.FileName+".1", item.S.Name, err)
	}
	f.file, err = os.Create(item.S.Output.FileName)
	if err != nil {
		log.Panicf("Cannot create file referenced at '%s' for sample '%s' with error %s", item.S.Output.FileName, item.S.Name, err)
	}
	f.fileSize = 0
}
