package main

import (
	"context"
	"github.com/malumar/filekit"
	"github.com/malumar/zoha/pkg/mtp"
	"log/slog"
	"os"
	"path/filepath"
)

const reverseHostname = "serveros-lmtp-test"
const address = "0.0.0.0:2100"

func main() {
	Zocha()
}

func Zocha() {
	storagePath, err := os.UserHomeDir()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	storagePath = filepath.Join(storagePath, "zoha-lmtp-example")
	spoolPath := filepath.Join(storagePath, "spool")
	mailsPath := filepath.Join(storagePath, "mails")

	if err := filekit.MkDirIfNotExists(storagePath, 0750); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	if err := filekit.MkDirIfNotExists(spoolPath, 0750); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	if err := filekit.MkDirIfNotExists(mailsPath, 0750); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	srv := mtp.NewDefaultListenerExt(address, reverseHostname, mtp.ReceiveFromLocalUsers|
		mtp.ReceiveFromRemoteUsers|
		mtp.AllowLocalhostConnection|
		mtp.LmtpService)

	if err := srv.Listen(); err != nil {
		panic(err)
	}

	if err := srv.Run(ctx, NewSupervisor(reverseHostname, mailsPath, spoolPath, ctx), cancelFunc); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

}
