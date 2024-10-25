package watcher

import (
	"context"
	"fmt"
	"github.com/malumar/filekit"
	"io"
	"log/slog"
	"os"
	"time"
)

var logger = slog.With("watcher")

func WatchModificationOfFile(ctx context.Context, filename string, maxSizeToRead int, searchThisLastContent string, hook func(content string) bool) {
	lastContent := searchThisLastContent
	first := true
	logger.Info("last loaded timestamp", "timestamp", lastContent)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// the first time we load without falling asleep
			if !first {
				time.Sleep(time.Second * 4)
			} else {
				first = false
			}
			dat, err := readNoMoreThan(filename, maxSizeToRead)
			if err != nil {
				logger.Error("read error", "filename", filename, "err", err)
				continue
			}
			sData := string(dat)
			if (len(lastContent) == 0 && len(sData) == 0) || sData == "" {
				continue
			}

			if lastContent != sData {
				// if the content differs, we replace it
				if hook(sData) {
					lastContent = sData
					logger.Info("Config reloaded")
					continue
				}
			}
		}
	}
}

func readNoMoreThan(name string, maxSize int) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF

	// If a file claims a small size, read at least 512 bytes.
	// In particular, files in Linux's /proc claim size 0 but
	// then do not work right if read in small pieces,
	// so an initial read of 1 byte would not work correctly.
	if size < 512 {
		size = 512
	}

	if size > maxSize {
		if maxSize > 512 {
			size = maxSize
		}
	}

	data := make([]byte, 0, size)
	totalReaded := 0
	for {
		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}
		totalReaded += n
		if totalReaded >= size {
			return data, nil

		}
	}
}

func CreateConfigDir(dir string) error {
	if err := filekit.MkDirIfNotExists(dir, 0600); err != nil {
		return fmt.Errorf("error creating configuration folder %v", dir)
	}
	return nil
}
