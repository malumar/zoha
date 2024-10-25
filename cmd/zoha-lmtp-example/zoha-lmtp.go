// An example LMTP client that should be installed on each node that stores EML files
package main

import (
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/internal/example"
	"github.com/malumar/zoha/pkg/mtp"
	"github.com/malumar/zoha/pkg/simplelmtpsession"
	"log/slog"
)

func main() {
	example.ZohaLmtp((func(supervisor api.MaildirSupervisor, l *slog.Logger) mtp.Session {
		return simplelmtpsession.NewLmtpSession(supervisor, l)
	}))
}
