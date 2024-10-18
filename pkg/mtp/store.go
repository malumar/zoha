package mtp

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	DefaultFolderPerm os.FileMode = 0700
	DefaultFilePerm   os.FileMode = 0640
)

// DefaultNewFileFromDelivery Default function to save the message to disk
func DefaultNewFileFromDelivery(proxy MessageReceiverProxy, delivery Delivery, hostName string, baseStoragePath string) (filename string, dataSize int64, err error) {
	return NewFileFromMailboxDelivery(proxy, delivery, hostName, baseStoragePath, DefaultFolderPerm, DefaultFilePerm)
}

func NewFileFromMailboxDelivery(proxy MessageReceiverProxy, delivery Delivery, hostName string, baseStoragePath string, folderPerm, filePerm os.FileMode) (filename string, dataSize int64, err error) {

	pth, err := delivery.Mailbox.
		CreateNewMailDirIfNotExists(baseStoragePath, delivery.CustomMailbox, 0700, 0640)
	if err != nil {
		return "", 0, err
	}

	buf := bytes.Buffer{}
	buf.WriteString(delivery.HeaderToString())
	if err := proxy.MoveBufferPosToStart(); err != nil {
		return "", 0, err
	}

	fs := buf.Len() + proxy.InitialMessageSize()

	fn := DovecotMessageFilenameV2(&delivery, proxy, hostName, fs)
	fullFn := filepath.Join(pth, fn)

	var removeFile bool
	var dstFile *os.File

	defer func() {
		if dstFile != nil {
			if err := dstFile.Close(); err != nil {
				logger.Error("error closing file with received mail", "filename", fullFn, "err", err)
			}
			if removeFile {
				if err := os.Remove(fullFn); err != nil {
					logger.Error("error deleting a message that ended with an error", "err", err)
				}
			}
		}
		dstFile = nil
	}()

	dstFile, err = OpenFile(fullFn, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		return "", 0, err
	}

	hs, err := dstFile.Write(buf.Bytes())
	if err != nil {
		logger.Error("error writing message headers", "err", err)
		removeFile = true
		return "", 0, err
	}

	ds, err := io.Copy(dstFile, proxy.GetBuffer())
	if err != nil {
		logger.Error("error writing message body", "err", err)

		removeFile = true
		return "", 0, err
	}

	dataSize = ds + int64(hs)

	return fullFn, dataSize, nil
}

/*
backup against StoreMessage
func NewFileFromMailboxDelivery(proxy MessageReceiverProxy, delivery Delivery, hostName string, baseStoragePath string, folderPerm, filePerm os.FileMode) (filename string, dataSize int64, err error) {

	pth, err := delivery.Maildir.
		CreateNewMailDirIfNotExists(baseStoragePath, delivery.CustomMailbox, 0700, 0640)
	if err != nil {
		return "", 0, err
	}

	buf := bytes.Buffer{}
	buf.WriteString(delivery.HeaderToString())
	if err := proxy.MoveBufferPosToStart(); err != nil {
		return "", 0, err
	}

	fs := buf.Len() + proxy.InitialMessageSize()

	fn := DovecotMessageFilenameV2(&delivery, proxy, hostName, fs)
	fullFn := filepath.Join(pth, fn)

	var removeFile bool
	var dstFile *os.File

	defer func() {
		if dstFile != nil {
			if err := dstFile.Close(); err != nil {
				logger.Error("error closing file with received mail %s: %v", fullFn, err)
			}
			if removeFile {
				if err := os.Remove(fullFn); err != nil {
					logger.Error("error deleting a message whose saving ended with an error: %v", err)
				}
			}
		}
		dstFile = nil
	}()

	dstFile, err = OpenFile(fullFn, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		return "", 0, err
	}

	hs, err := dstFile.Write(buf.Bytes())
	if err != nil {
		logger.Fatal("error writing message headers", "err", err)
		removeFile = true
		return "", 0, err
	}

	ds, err := io.Copy(dstFile, proxy.GetBuffer())
	if err != nil {
		logger.Error("error writing message body", "err", err)
		removeFile = true
		return "", 0, err
	}

	dataSize = ds + int64(hs)

	return fullFn, dataSize, nil
}
*/

// StoreMessage file to remember that the folder you are saving to must have already been created by you,
//  so you need to create it yourself before calling this function
// @filename the full save path
func StoreMessage(appendBytes []byte, proxy MessageReceiverProxy, delivery Delivery, filename string) (fileSize int64, err error) {
	buf := bytes.Buffer{}
	buf.WriteString(delivery.HeaderToString())
	err = proxy.MoveBufferPosToStart()
	if err != nil {
		return 0, err
	}

	//	fs := int64(buf.Len() + proxy.InitialMessageSize())

	//fn := DovecotMessageFilenameV2(&delivery, proxy, hostName, fs)
	//fullFn := filepath.Join(pth, fn)

	var removeFile bool
	var dstFile *os.File

	defer func() {
		if dstFile != nil {
			if err := dstFile.Close(); err != nil {
				logger.Error("file close error", "filename", filename, "err", err)
			}
			if removeFile {
				if err := os.Remove(filename); err != nil {
					logger.Error("error deleting a message that ended with an error", "err", err)
				}
			}
		}
		dstFile = nil
	}()

	dstFile, err = OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		return 0, err
	}
	if len(appendBytes) > 0 {
		hs, err := dstFile.Write(appendBytes)
		if err != nil {
			removeFile = true
			logger.Error("failed to write additional bytes", "err", err)
			return 0, fmt.Errorf("failed to write additional bytes %v", err)
		}
		if hs != len(appendBytes) {
			removeFile = true
			errMsg := fmt.Errorf("we wanted to write %d but we only wrote %d", len(appendBytes), hs)
			logger.Error(errMsg.Error())
			return 0, errMsg
		}
	}
	hs, err := dstFile.Write(buf.Bytes())
	if err != nil {
		logger.Error("Error writing message header", "err", err)
		removeFile = true
		return 0, fmt.Errorf("Error writing message header %v", err)
	}

	ds, err := io.Copy(dstFile, proxy.GetBuffer())
	if err != nil {
		logger.Error("error writing message body", "err", err)
		removeFile = true
		return 0, fmt.Errorf("Error writing message body %v", err)
	}

	return ds + int64(hs), nil

}

func StoreMessageFromWriter(appendBytes []byte, wrto io.WriterTo, filename string) (fileSize int64, err error) {

	var removeFile bool
	var dstFile *os.File

	defer func() {
		if dstFile != nil {
			if err := dstFile.Close(); err != nil {
				logger.Error("error closing file", "filename", filename, "err", err)
			}
			if removeFile {
				if err := os.Remove(filename); err != nil {
					logger.Error("error deleting a message whose saving ended in error", "err", err)
				}
			}
		}
		dstFile = nil
	}()

	dstFile, err = OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		return 0, err
	}
	if len(appendBytes) > 0 {
		hs, err := dstFile.Write(appendBytes)
		if err != nil {
			removeFile = true
			logger.Error("failed to write additional bytes", "err", err)
			return 0, fmt.Errorf("failed to write additional bytes %v", err)
		}
		if hs != len(appendBytes) {
			removeFile = true
			errMsg := fmt.Errorf("We wanted to write %d but we only wrote %d", len(appendBytes), hs)
			logger.Error(errMsg.Error())
			return 0, errMsg
		}
	}

	hs, err := wrto.WriteTo(dstFile)
	if err != nil {
		logger.Error("error writing message body", "err", err)
		removeFile = true
		return 0, fmt.Errorf("Error writing message body %v", err)
	}

	return int64(hs), nil

}
