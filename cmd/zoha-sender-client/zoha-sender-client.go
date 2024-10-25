package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/malumar/domaintools"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/internal/example"
	"github.com/malumar/zoha/pkg/mtp"
	"github.com/malumar/zoha/pkg/mtp/headers"
	"github.com/malumar/zoha/pkg/simplelmtpsession"
	"io"
	"log/slog"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const PauseInSecondsIfErrorOccurred = 5

var logger = slog.With("package", "zoha-sender-server")

func main() {

	// fixme supervisor don't have implemented supervisor.MainSenderNode()
	supervisor := example.NewDummySupervisor(context.Background(), func(supervisor api.MaildirSupervisor, l *slog.Logger) mtp.Session {
		return simplelmtpsession.NewLmtpSession(supervisor, l)
	})

	logger.Info("ZohaSenderClient started", "spool", supervisor.AbsoluteSpoolPath())
	defer logger.Info("ZohaSenderClient exited")

	if len(supervisor.AbsoluteSpoolPath()) == 0 {
		logger.Error("spool path is empty")
		os.Exit(1)
	}

	for {
		logger.Info("I'm looking for emails that need to be sent to postfix")
		dirs, err := os.ReadDir(supervisor.AbsoluteSpoolPath())
		if err != nil {
			logger.Error("read error", "err", err.Error())
			time.Sleep(time.Second * PauseInSecondsIfErrorOccurred)
		}
		foundFiles = 0
		for _, entry := range dirs {
			if !entry.IsDir() {
				continue
			}
			found := forwardMessagesPendingToSubmissions(supervisor, entry.Name())
			foundFiles += found
		}

		if foundFiles == 0 {
			time.Sleep(time.Second * PauseInSecondsIfErrorOccurred)
		}

	}

}

func forwardMessagesPendingToSubmissions(supervisor api.MaildirSupervisor, inDir string) (founded int) {
	pth := filepath.Join(supervisor.AbsoluteSpoolPath(), inDir)
	files, err := os.ReadDir(pth)
	if err != nil {
		logger.Error("forward read error w %ss %v", pth, err.Error())
		return
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fn := filepath.Join(pth, f.Name())
		fi, err := f.Info()
		if err != nil {
			logger.Error("failed to get file info", "filename", fn, "err", err)
			continue
		} else {
			// nie...
			if fi.Size() < 10 {
				if err := os.Remove(fn); err != nil {
					logger.Error("the file size is zero, but I couldn't delete it", "filename", fn,
						"err", err)
					continue
				}
			}
			if err := sendMailToPostfix(supervisor, fn); err != nil {
				logger.Error("sending a message with an error", "filename", fn, "err", err)
			} else {
				founded++
			}

		}
		logger.Debug("processing mail", "no", founded+1, "filename", fn)
	}
	return
}

var foundFiles int

func sendMailToPostfix(supervisor api.Supervisor, filename string) (retErr error) {

	defer func(returnErr *error) {
		if r := recover(); r != nil {
			var ok bool
			var err error
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("recover panic pkg: %v -- err %v", r, err)
			}
			*returnErr = err
		}

	}(&retErr)

	finishHim := false

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_RDWR, 0600)
	logger.Debug("message opened", "filename", filename)
	if err != nil {
		logger.Error("i can't open message", "filename", filename, "err", err)
		retErr = err
		return
	}
	defer func() {
		if finishHim {
			if err := f.Truncate(0); err != nil {
				logger.Error("can't truncate message file", "filename", filename, "err", err)
			}
		}

		if err := f.Close(); err != nil {
			logger.Error("can't close message, maybe already closed?", "filename", filename, "err", err)
		}
		if finishHim {
			if err := os.Remove(filename); err != nil {
				logger.Error("i can't remove message",
					"filename", filename, "err", err)
			}
		}
	}()
	msg, err := mail.ReadMessage(f)
	if err != nil {
		logger.Error("i can't read mime message", "fileaname", filename, "err", err)
		retErr = err
		return
	}

	xaction := msg.Header.Get(headers.XAction)
	if len(xaction) == 0 {
		logger.Error("XAction is empty", "filename", filename)
		retErr = fmt.Errorf("XAction is empty %s", filename)
		return

	}

	xval := strings.Split(xaction, "|")
	if len(xval) != 2 {
		logger.Error("XAction does not consist of two parts", "filename", filename, "value", xval)
		retErr = fmt.Errorf("XAction does not consist of two parts %s %v", filename, xval)
		return

	}
	if xval[0] == "" {
		xval[0] = supervisor.MailerDaemonEmailAddress()
	}
	if !domaintools.IsValidEmailAddress(xval[0]) {
		logger.Error("XAction, the sender have error in email address", "filename", filename, "value", xval[0])
		retErr = fmt.Errorf("XAction, the sender have error in email address %s %v", filename, xval)
		return

	}

	if !domaintools.IsValidEmailAddress(xval[1]) {
		logger.Error("XAction, the recipient has the wrong email address", "filename", filename,
			"value", xval)
		retErr = fmt.Errorf("XAction, the recipient has the wrong email address %s %v", filename, xval)
		return

	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		retErr = err
		return
	}
	if err := toPostfix(f, supervisor, xval[0], xval[1]); err != nil {
		// if we don't eat it, we report it, otherwise we delete it
		if strings.Index(err.Error(), "Recipient address rejected: User unknown in virtual mailbox table") == -1 {
			retErr = err
			return
		} else {
			logger.Error("XAction, the recipient does not exist, delete the message",
				"recipient", xval, "filename", filename, "err", err.Error())

		}
	}
	finishHim = true
	return nil
}

// toPostfix send message to postfix
// @zohaSenderServerNode host and port of zoha-sender-server on main postfix instance
func toPostfix(f *os.File, supervisor api.Supervisor, from, to string) error {

	scanner := bufio.NewScanner(f)

	var pos int64
	for scanner.Scan() {
		pos = int64(len(scanner.Text()) + 1)
		break
	}

	if _, err := f.Seek(pos, io.SeekStart); err != nil {
		return err
	}

	c, err := smtp.Dial(supervisor.MainSenderNode())
	if err != nil {
		return fmt.Errorf("I can't connect to MainSenderNode %v err:%v", supervisor.MainSenderNode(), err)
	}
	// your full hostname
	if err := c.Hello(supervisor.Hostname()); err != nil {
		return err
	}
	// todo add authorization

	// Set the sender and recipient first
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}

	closer := make(chan struct{}, 2)
	go copy(closer, wc, f)

	<-closer

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
