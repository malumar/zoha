// Forwards messages created by the service "send copy to" commands, autoresponder
// or aliases, to the main zoha-sender-server
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
	var foundFiles int
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

func forwardMessagesPendingToSubmissions(supervisor api.MaildirSupervisor, inDir string) (found int) {
	pth := filepath.Join(supervisor.AbsoluteSpoolPath(), inDir)
	files, err := os.ReadDir(pth)
	if err != nil {
		logger.Error("failed to read directory", "path", pth, "error", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(pth, file.Name())
		info, err := file.Info()
		if err != nil {
			logger.Error("failed to retrieve file info", "filename", filePath, "error", err)
			continue
		}

		// Remove empty or very small files
		if info.Size() < 10 {
			if err := os.Remove(filePath); err != nil {
				logger.Error("failed to delete small file", "filename", filePath, "error", err)
			}
			continue
		}

		// Dispatch the mail message
		if err := dispatchMailMessage(supervisor, filePath); err != nil {
			logger.Error("failed to dispatch message", "filename", filePath, "error", err)
		} else {
			found++
			logger.Debug("message dispatched successfully", "index", found, "filename", filePath)
		}
	}

	return found
}

func dispatchMailMessage(supervisor api.Supervisor, filename string) (retErr error) {
	defer handlePanic(&retErr)

	file, err := openMailFile(filename)
	if err != nil {
		return err
	}
	defer cleanupFile(file, filename, &retErr)

	msg, err := mail.ReadMessage(file)
	if err != nil {
		logger.Error("failed to read MIME message", "filename", filename, "error", err)
		return err
	}

	sender, recipient, err := readMessageHeader(supervisor, msg, filename)
	if err != nil {
		return err
	}

	if err := validateEmailAddresses(sender, recipient, filename); err != nil {
		return err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	return dispatchMail(file, supervisor, sender, recipient, filename)
}

// Function to handle panics and wrap errors
func handlePanic(returnErr *error) {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			*returnErr = fmt.Errorf("panic recovered: %w", err)
		} else {
			*returnErr = fmt.Errorf("panic recovered: %v", r)
		}
	}
}

// Opens the mail file and logs error if needed
func openMailFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		logger.Error("failed to open message file", "filename", filename, "error", err)
		return nil, err
	}
	return file, nil
}

// Reads and verifies the XAction header in the message
func readMessageHeader(supervisor api.Supervisor, msg *mail.Message, filename string) (sender, recipient string, err error) {
	xAction := msg.Header.Get(headers.XAction)
	if len(xAction) == 0 {
		logger.Error("XAction header is empty", "filename", filename)
		return "", "", fmt.Errorf("XAction header missing in file %s", filename)
	}

	parts := strings.Split(xAction, "|")
	if len(parts) != 2 {
		logger.Error("XAction header has invalid format", "filename", filename, "value", parts)
		return "", "", fmt.Errorf("XAction header format error in file %s", filename)
	}

	sender = parts[0]
	recipient = parts[1]
	if sender == "" {
		sender = supervisor.MailerDaemonEmailAddress()
	}
	return sender, recipient, nil
}

// Validates the email addresses for sender and recipient
func validateEmailAddresses(sender, recipient, filename string) error {
	if !domaintools.IsValidEmailAddress(sender) {
		logger.Error("invalid sender email in XAction", "filename", filename, "email", sender)
		return fmt.Errorf("invalid sender email in file %s", filename)
	}
	if !domaintools.IsValidEmailAddress(recipient) {
		logger.Error("invalid recipient email in XAction", "filename", filename, "email", recipient)
		return fmt.Errorf("invalid recipient email in file %s", filename)
	}
	return nil
}

// Handles dispatch and any necessary error handling
func dispatchMail(file *os.File, supervisor api.Supervisor, sender, recipient, filename string) error {
	err := dispatchMessageToPostfix(file, supervisor, sender, recipient)
	if err != nil && !strings.Contains(err.Error(), "User unknown in virtual mailbox table") {
		return err
	}
	if err != nil {
		logger.Error("recipient does not exist, message deleted", "filename", filename, "recipient", recipient, "error", err)
	}
	return nil
}

// Cleans up the file, truncating or deleting as necessary
func cleanupFile(file *os.File, filename string, retErr *error) {
	if *retErr == nil {
		if err := file.Truncate(0); err != nil {
			logger.Error("failed to truncate message file", "filename", filename, "error", err)
		}
	}
	if err := file.Close(); err != nil {
		logger.Error("failed to close message file", "filename", filename, "error", err)
	}
	if *retErr == nil {
		if err := os.Remove(filename); err != nil {
			logger.Error("failed to remove message file", "filename", filename, "error", err)
		}
	}
}

// dispatchMessageToPostfix dispatch message to postfix
// @zohaSenderServerNode host and port of zoha-sender-server on main postfix instance
func dispatchMessageToPostfix(f *os.File, supervisor api.Supervisor, from, to string) error {

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
		return fmt.Errorf("I can't connect to MainSenderNode %v err:%w", supervisor.MainSenderNode(), err)
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
	go copier(closer, wc, f)

	<-closer

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func copier(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
