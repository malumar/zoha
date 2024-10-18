package mtp

import (
	"fmt"
	"os"
	"time"
)

const (
	maxTries          = 5
	sleepStepDuration = time.Millisecond * 250
)

func OpenFile(filename string, flag int, perms os.FileMode) (f *os.File, err error) {
	return TryOpenFile(filename, maxTries, sleepStepDuration, flag, perms)
}

func TryOpenFile(filename string, howManyTries int, sleepStep time.Duration, flag int, perms os.FileMode) (f *os.File, err error) {
	tryNo := 0
	for {
		tryNo++
		f, err = os.OpenFile(filename, flag, perms)
		if err == nil {
			break
		}
		logger.Error(fmt.Sprintf("error opening file for writing, attempt number %d", tryNo),
			"filename", filename, "err", err)
		time.Sleep(sleepStep * time.Duration(tryNo))
		logger.Error("Działamy dalej")
		if tryNo == howManyTries {
			logger.Error("can't open the file", "filename", filename, "err", err)
			return
		}
	}

	return
}
