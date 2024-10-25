package proxy

import (
	"github.com/malumar/zoha/api"
	"log/slog"
	"net/http"
	"strings"
)

var logger = slog.With("package", "proxy")

const (
	InvalidLoginOrPasswordResp = "Invalid login or password / Bledny login lub haslo"
	InvalidProtocol            = "Invalid protocol"
)

func New(address string, supervisor api.Supervisor) *Server {
	s := &Server{
		address:    address,
		supervisor: supervisor,
		supportedProtocols: map[string]api.Service{
			"pop3": api.PopService,
			"smtp": api.SmtpOutService,
			"imap": api.ImapService,
		},
	}
	return s
}

type Server struct {
	address            string
	supervisor         api.Supervisor
	supportedProtocols map[string]api.Service
}

func (self *Server) Load() error {

	return nil
}

func (self *Server) Run() error {

	http.HandleFunc("/auth", self.authRequst)

	logger.Info("Starting server")
	if err := http.ListenAndServe(self.address, nil); err != nil {
		logger.Error(err.Error())
	} else {
		logger.Info("Server stopped")
	}
	return nil
}

func (self *Server) authRequst(resp http.ResponseWriter, req *http.Request) {

	c := self.getClient(req)
	if len(c.Login) == 0 || len(c.Password) == 0 {
		logger.Error("no login or password entered",
			"user", c.Login, "password", len(c.Password) > 0)
		self.invalidLoginOrPassword(c.Ip, c.Login, resp)
		return
	}

	if len(c.Protocol) == 0 {
		logger.Error("erroneous protocol", "user", c.Login, "protocol", c.Protocol)
		self.invalidProtocol(c.Ip, c.Login, resp)
		return
	}

	var dest *api.Mailbox

	if service, found := self.supportedProtocols[c.Protocol]; !found {
		slog.Error("not supported protocol", "user", c.Login, "protocol", c.Protocol)
		self.invalidProtocol(c.Ip, c.Login, resp)
		return
	} else {
		dest = self.supervisor.Authorization(c.Login, c.Password, service, c.UseSsl)
	}

	if dest == nil {
		slog.Error("I can't authorize", "protocol", c.Protocol, "user", c.Login)
		self.invalidLoginOrPassword(c.Ip, c.Login, resp)
		return
	}

	host, port := c.getDestination(dest)

	if len(host) == 0 || len(port) == 0 {
		logger.Error("i can't authorize no destination host or port",
			"account", dest.Account, "email", dest.EmailLowerAscii,
			"protocol", c.Protocol, "destHost", host, "destPort", port)
		self.fail(c.Ip, c.Login, "Denied access to the service "+c.Protocol, resp)
		return
	}

	self.pass(c.Ip, dest, host, port, resp)

}

func (self *Server) getClient(req *http.Request) *client {
	cl := client{}

	cl.Login = req.Header.Get("Auth-User")
	if len(cl.Login) == 0 {
		cl.Login = req.Header.Get("Http_auth_user")
	}
	cl.Password = req.Header.Get("Auth-Pass")
	if len(cl.Password) == 0 {
		cl.Password = req.Header.Get("Http_auth_pass")
	}

	cl.Ip = req.Header.Get("Client-Ip")
	if len(cl.Ip) == 0 {
		cl.Ip = req.Header.Get("Http_client_ip")
	}
	usedSslVal := req.Header.Get("Auth-Ssl")

	if len(usedSslVal) == 0 {
		usedSslVal = req.Header.Get("Http_auth_sasl")
	}

	cl.UseSsl = api.Answer(usedSslVal == "on")

	// fixing nginxa values
	cl.Password = strings.Replace(cl.Password, "%20", " ", -1)
	cl.Password = strings.Replace(cl.Password, "%25", "%", -1)

	cl.Protocol = self.getProtocol(req)

	return &cl
}

func (self *Server) getProtocol(req *http.Request) string {
	protocol := req.Header.Get("Auth-Protocol")
	if len(protocol) == 0 {
		protocol = req.Header.Get("Http_auth_protocol")
	}
	return protocol
}

func (self *Server) invalidProtocol(clientIp, user string, resp http.ResponseWriter) {
	self.fail(clientIp, user, InvalidProtocol, resp)
}

func (self *Server) invalidLoginOrPassword(clientIp, user string, resp http.ResponseWriter) {
	self.fail(clientIp, user, InvalidLoginOrPasswordResp, resp)
}

func (self *Server) fail(clientIp, user string, message string, resp http.ResponseWriter) {
	logger.Info("clientIp: %s, user: %s odmowa:%s ", clientIp, user, message)
	resp.Header().Set("Auth-Status", message)
	return
}

func (self *Server) pass(clientIp string, dest *api.Mailbox, host, port string, resp http.ResponseWriter) {
	resp.Header().Set("Auth-Status", "OK")
	resp.Header().Set("Auth-Server", host)
	resp.Header().Set("Auth-Port", port)
	resp.WriteHeader(http.StatusNoContent)
	logger.Info("passed", "destHost", host, "destPort", port, "clientIp", clientIp,
		"email", dest.EmailLowerAscii, "account", dest.Account)
	return
}
