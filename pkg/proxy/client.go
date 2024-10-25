package proxy

import (
	"github.com/malumar/zoha/api"
	"log/slog"
)

type client struct {
	Login    string
	Password string
	Ip       string
	UseSsl   api.Maybe
	Service  api.Service
	Protocol string
}

func (self *client) getDestination(mailbox *api.Mailbox) (host string, port string) {
	switch self.Service {
	case api.ImapService:
		host, port = self.getImapDestination(mailbox)
		break
	case api.PopService:
		host, port = self.getPop3Destination(mailbox)
		break
	case api.SmtpOutService:
		host, port = self.getSmtpOutDestination(mailbox)
		break
	default:
		slog.Error("not supported api service", "service", self.Service)
	}
	return
}
func (self *client) getImapDestination(mailbox *api.Mailbox) (host string, port string) {
	if mailbox.PrivatePortSfx == "" {
		// it doesn't matter because the traffic is already encrypted when we are on localhost
		port = "1430"
		// fixme: Is it still needed?
		if self.UseSsl.True() {
			//destPort = "993"
		} else {
			//destPort = "1430"
		}
		// fixme: here we should probably return our public IP?
		//  destIp = "127.0.0.1"
		host = mailbox.ImapIp
	} else {
		// 51XXX for imaps
		if self.UseSsl.True() {
			host = "127.0.0.1"
			port = "51" + mailbox.PrivatePortSfx
		} else {
			// not encrypted, so we're going naked
			host = mailbox.ImapIp
			port = "143"
		}
	}
	return
}

func (self *client) getPop3Destination(mailbox *api.Mailbox) (host string, port string) {
	if mailbox.PrivatePortSfx == "" {
		// it doesn't matter because the traffic is already encrypted when we are on localhost
		port = "1110"
		// fixme: Is it still needed?
		if self.UseSsl.True() {
			//destPort = "995"
		} else {
			//destPort = "110"
		}
		// fixme: here we should probably return our public IP?
		//  destIp = "127.0.0.1"
		host = mailbox.ImapIp
	} else {
		// 51XXX for imaps
		if self.UseSsl.True() {
			host = "127.0.0.1"
			port = "52" + mailbox.PrivatePortSfx
		} else {
			// not encrypted, so we're going naked
			host = mailbox.ImapIp
			port = "110"
		}
	}
	return
}

func (self *client) getSmtpOutDestination(mailbox *api.Mailbox) (host string, port string) {
	if mailbox.PrivatePortSfx == "" {
		// it doesn't matter because the traffic is already encrypted when we are on localhost
		port = "5870"
		// fixme: Is it still needed?
		if self.UseSsl.True() {
			//destPort = "465"
		} else {
			//destPort = "587"
		}
		// fixme: here we should probably return our public IP?
		//  destIp = "127.0.0.1"
		host = mailbox.ImapIp
	} else {
		// 51XXX for imaps
		if self.UseSsl.True() {
			host = "127.0.0.1"
			port = "53" + mailbox.PrivatePortSfx
		} else {
			// not encrypted, so we're going naked
			host = mailbox.ImapIp
			port = "587"
		}
	}
	return
}
