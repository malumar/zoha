package example

import (
	"context"
	"github.com/malumar/filekit"
	"github.com/malumar/zoha/internal/example/data"
	"github.com/malumar/zoha/pkg/mtp"
	"log/slog"
	"os"
	"path/filepath"
)

// ZohaLmtp Example of ZohaLmtp
func ZohaLmtp(sessionFactoryHandler SessionFactoryHandler) {
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
	srv := mtp.NewDefaultListenerExt(data.LmtpListenAddress, data.ReverseHostname, mtp.ReceiveFromLocalUsers|
		mtp.ReceiveFromRemoteUsers|
		mtp.AllowLocalhostConnection|
		mtp.LmtpService)

	if err := srv.Listen(); err != nil {
		panic(err)
	}
	if err := srv.Run(ctx,
		NewDummyMaildirSupervisor(ctx, sessionFactoryHandler, mailsPath, spoolPath), cancelFunc); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

}
