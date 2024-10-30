package mtp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/bitmask"
	"io"
	"log/slog"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"net"
	"regexp"
)

const (
	DefaultMaxErrorPerConnection = 5
	DefaultMaxConnections        = 1000
	DefaultAddress               = "localhost:1025"

	// DefaultProtocolTimeoutSec How many seconds do we wait for a command to be issued from the client
	DefaultProtocolTimeoutSec = 8

	// DefaultDataStartReceivingTimeoutSec How many seconds do we wait for the start of data transmission
	// (just after receiving the data command
	DefaultDataStartReceivingTimeoutSec = 10

	// DefaultDataReceivingTimeoutSec How many seconds do we wait for data after the data command
	DefaultDataReceivingTimeoutSec = 10 // 5 * 60
)

const (
	DefaultListenerNameFormat              = "Super %s"
	DefaultListenerVersion                 = "1.0.0"
	DefaultListenerMaxMessageSizeInMB      = 25
	DefaultListenerMaxMessageSizeInBytes   = DefaultListenerMaxMessageSizeInMB * (1024 * 1024 * 1024)
	DefaultListenerMaxRecipientsPerMessage = 100
)

func NewLmtp(hostname string) *Listener {
	return NewDefaultListenerExt(DefaultAddress, hostname, 0)
}

// NewDefaultListenerExt Mainly for the needs of testing, but also for simple customers
func NewDefaultListenerExt(address, reverseHostname string, flags bitmask.Flag64) *Listener {
	var name string
	if flags.HasFlag(Authorization) {
		name = "SMTP"
	} else {
		name = "LMTP"
	}

	l := NewDefaultListenerWithAddress(address)
	l.reverseHostName = reverseHostname
	l.name = fmt.Sprintf(DefaultListenerNameFormat, name)
	l.version = DefaultListenerVersion
	l.flags = flags
	return l
}

func NewDefaultListener() *Listener {
	return NewDefaultListenerWithAddress(DefaultAddress)
}

func NewDefaultListenerWithAddress(address string) *Listener {
	return NewListener(address, DefaultMaxErrorPerConnection, DefaultMaxConnections, 0)
}

func NewListener(address string, maxErrorsPerConnection int, maxConnections int, flags bitmask.Flag64) *Listener {
	s := &Listener{
		flags:                   flags,
		MaxErrorsPerConnection:  maxErrorsPerConnection,
		MaxConnections:          maxConnections,
		MaxRecipientsPerMessage: DefaultListenerMaxRecipientsPerMessage,
		MaxMessageSizeInBytes:   DefaultListenerMaxMessageSizeInBytes,
		address:                 address,
		//wg:                      sync.WaitGroup{},

		ProtocolTimeoutDuration:           time.Duration(DefaultProtocolTimeoutSec) * time.Second,
		DataStartReceivingTimeoutDuration: time.Duration(DefaultDataStartReceivingTimeoutSec) * time.Second,
		DataReceivingTimeoutDuration:      time.Duration(DefaultDataReceivingTimeoutSec) * time.Second,

		metrics: &Metrics{},
	}
	// s.init(name, store)
	return s
}

type SupervisorExt interface {
	api.Supervisor
	api.MaildirStorage

	// OpenSession Opening or ending the session.
	// If we don't want to connect, we need to give the content to rejectMessage
	OpenSession(ci ConnectionInfo, l *slog.Logger) (session Session, err error)
	CloseSession(session Session)
}

type Listener struct {
	// flags in what mode we work
	flags           bitmask.Flag64
	reverseHostName string
	name            string
	version         string

	address        string
	MaxConnections int
	// MaxMessageSizeInBytes   int
	MaxErrorsPerConnection  int
	MaxRecipientsPerMessage int
	MaxMessageSizeInBytes   int64
	supervisor              SupervisorExt
	wg                      sync.WaitGroup
	netListener             net.Listener
	activeConnections       chan int
	quitChan                chan bool

	ProtocolTimeoutDuration           time.Duration
	DataReceivingTimeoutDuration      time.Duration
	DataStartReceivingTimeoutDuration time.Duration

	metrics *Metrics
}

func (l *Listener) Address() string {
	return l.address
}

func (l *Listener) Listen() (err error) {
	if l.netListener != nil {
		err := fmt.Errorf("listener is already running on %s", l.address)
		logger.Error(err.Error())
		return err
	}

	l.quitChan = make(chan bool, 5)

	l.netListener, err = net.Listen("tcp", l.address)
	if err != nil {
		logger.Error("listener has not been opened", "err", err.Error())
		return
	}

	logger.Info("start listen", "address", l.address)

	return nil
}

func (l *Listener) Run(ctx context.Context, supervisor SupervisorExt, cancelFunc context.CancelFunc) error {

	if l.netListener == nil {
		err := fmt.Errorf("listener has not been started yet")
		logger.Error(err.Error())
		return err
	}

	defer func() {
		logger.Info("Terminating the listener. I'm waiting for all listener threads to finish")
		l.wg.Wait()
		close(l.activeConnections)
		// fmt.Println("Widzisz mnie?")
	}()
	l.supervisor = supervisor

	// currently active client list
	l.activeConnections = make(chan int, l.MaxConnections)

	logger.Info("I'm waiting for clients")
	// przez to bedziemy czekac w nieskonczonosc

	l.wg.Add(1)
	go l.waitForCommands(ctx)

	var clientId int64
	clientId = 1
	for {
		logger.Debug("I'm waiting for next client")
		con, err := l.netListener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				cancelFunc()
				break

			}
			continue
		}

		// We are waiting for the client
		l.activeConnections <- 1
		l.wg.Add(1)
		go l.handleConnection(ctx, l.newConnection(con, clientId))
		clientId++
	}

	return nil
}

func (l *Listener) waitForCommands(ctx context.Context) {

	defer func() {
		l.doClose()
	}()

	ticker := time.NewTicker(500 * time.Millisecond)

	// by this select there will be problems 100% cpu because
	// it is a full loop you should wait for a signal something like finishim
	//for {
	//	select {
	//	case <-ctx.Done():
	//		logger.Info("Termination by context")
	//		return
	//	case <-l.quitChan:
	//		logger.Info("Termination by quitChan")
	//		return
	//	default:
	//
	//	}
	//}
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			logger.Info("Termination by context")
			return
		case <-l.quitChan:
			logger.Info("Termination by quitChan")
			return
		default:

		}
	}
}

// Stop You can quit by Shutdown or closing the context (ctx.Done) which is equivalent
func (l *Listener) Stop() {
	// only if you use waitForCommands
	l.quitChan <- true
	// If not, then this but it causes us to destroy the thread while we still have a clien
	// l.doClose()
}

func (l *Listener) doClose() {
	if l.netListener != nil {
		if errc := l.netListener.Close(); errc != nil {
			logger.Error("Close listener", "err", errc)
		} else {
			logger.Info("Listener has been closed")

		}
	}
	l.metrics = nil
}
func (l *Listener) newConnection(con net.Conn, clientId int64) *Client {
	c := &Client{

		Con:                     con,
		Info:                    NewConnectionInfo(clientId, l.flags),
		MaxErrorsPerConnection:  l.MaxErrorsPerConnection,
		MaxRecipientsPerMessage: l.MaxRecipientsPerMessage,
		MaxMessageSizeInBytes:   l.MaxMessageSizeInBytes,
		Flags:                   l.flags,
		ListenerReverseHostname: l.reverseHostName,
		ListenerAppName:         l.name,
		state:                   NewConnectionState(l.flags),

		DataStartReceivingTimeoutDuration: l.DataStartReceivingTimeoutDuration,
		ProtocolTimeoutDuration:           l.ProtocolTimeoutDuration,
	}

	return c
}

func (l *Listener) handleConnection(ctx context.Context, client *Client) {
	defer func() {
		var connections int32
		if l.metrics != nil {
			connections = l.metrics.ConnectedClients()
		}
		client.logger.Info(client.state.disconnectReason,
			ActionNameField, "connection:close",
			"connections", connections,
		)
		if l.metrics != nil {
			l.metrics.ClientDisconnected()
		}
		l.closeClient(client)
		l.wg.Done()
		// All threads here are closed
	}()
	client.CurrentClientNo = l.metrics.ClientConnected()

	l.letsGetToKnowEachOther(client)

	for !client.WeClose {
		select {
		// do we have a request to terminate
		case <-ctx.Done():
		case <-l.quitChan:
			return
		default:
			lastState := client.state.current
			switch lastState {
			case StatePassive:
				client.Con.SetDeadline(time.Now().Add(l.ProtocolTimeoutDuration))

				client.logger.Debug("handleConnection StatePassive",
					ActionNameField, "state:waitingForCommands",
					DecisionKey, Dunno.String(),
				)

				l.executeCommands(ctx, client.MaxErrorsPerConnection, client)

			case StateData:
				// separately the limit for waiting for the start of data transfer is set in the Data command
				client.logger.Debug("handleConnection StateData",
					ActionNameField, "state:waitingForCommands",
					DecisionKey, Dunno.String(),
				)

				if l.receiveData(ctx, client) {
					return
				}

			}

		}
	}
}

// letsGetToKnowEachOther
func (l *Listener) letsGetToKnowEachOther(client *Client) {
	client.Info.UUId = client.generateUid()

	client.logger = logger.With("cid", client.Info.UUId)

	addr := client.Con.RemoteAddr().String()
	portPos := strings.Index(addr, ":")
	if portPos > 0 {
		addr = addr[:portPos]
	}
	client.Text = textproto.NewConn(client.Con)
	client.Info.LocalAddr = client.Con.LocalAddr().String()
	client.Info.RemoteAddr = addr
	client.Info.ReverseHostName = ""
	client.Info.StartConnection = time.Now()
	var reverseHostnameErr error
	revlist, er := net.LookupAddr(client.Info.RemoteAddr)
	if er == nil {
		for _, value := range revlist {
			if len(client.Info.ReverseHostName) > 0 {
				client.Info.ReverseHostName += ", "
			}
			client.Info.ReverseHostName += value
		}
	} else {
		reverseHostnameErr = er
	}

	client.logger.Info(
		"letsGetToKnowEachOther",
		ActionNameField, "connection:open",
		"localAddr", client.Info.LocalAddr,
		"remoteAddr", client.Info.RemoteAddr,
		"reverseHostName", client.Info.ReverseHostName,
		"helo", client.Helo,
		"reverseHostErr", reverseHostnameErr,
	)

	//	client.MessageListener = mailnt.NewSmptMessageListener(&client.Listener.Store)
	//	client.MessageListener = mailnt.NewSmptMessageListener(client.Listener.Manager)
	// client.MessageListener = mailnt.NewSmptMessageListener(mailnt.MailManager)
	client.state.current = StatePassive

	client.isReady = true

	l.openSession(client)

}

func (l *Listener) openSession(client *Client) {
	if session, err := l.supervisor.OpenSession(client.Info, client.logger); err != nil {
		// 	(host mx-aol.mail.gm0.yahoodns.net[x.x.x.x] said: 421 4.7.0
		//	[TSS04] Messages from x.x.x.x temporarily deferred due to user complaints - x.x.x.x;
		//	see https://help.yahoo.com/kb/postmaster/SLN3434.html (in reply to MAIL FROM command
		// to jest nasz problem
		if sendedErrorToClientIfReject(client, err) {
			client.logger.Error("openSession",
				ActionNameField, "session:reject",
				DeliveryStatusField, DeliveryRejected,
			)

		} else {
			client.WriteCmd("471 Please try again later / open session error\n221 Bye")
			client.WeClose = true
		}
		return
	} else {
		if session == nil {
			client.logger.Error("session is empty",
				ActionNameField, "session:open",
				DeliveryStatusField, DeliveryRejected,
			)
			client.WriteCmd("471 Please try again later\n221 Bye")
			client.WeClose = true
		} else {
			client.Session = session
			client.WriteCmd(fmt.Sprintf("220 %s %s %s", l.reverseHostName, l.name, l.version))
		}
	}
}

func (l *Listener) closeClient(client *Client) {
	defer func() {
		if client.Session != nil {
			l.supervisor.CloseSession(client.Session)
			client.Session = nil
		}
		client.state.Close()

		client = nil
	}()
	client.Con.Close()
	client.Text.Close()

	//	client.Con.Shutdown()
	//	client.Listener = nil
	client.state.callFunc = nil
	client.Con = nil
	client.Text = nil
	client.RejectInfo = nil

	<-l.activeConnections
}

func (l *Listener) receiveData(ctx context.Context, client *Client) (closeConnection bool) {

	// set again so that there is an accurate amount of time to read the data
	client.Con.SetReadDeadline(time.Now().Add(l.DataReceivingTimeoutDuration))

	if err := client.state.receiver.AcceptData(ctx, client, client.Text.DotReader()); err != nil {
		if !sendedErrorToClientIfReject(client, err) {

			client.logger.Error("Refusal to send messages as the specified user",
				ActionNameField, "receiveData",
				FromField, client.state.from.String(),
				ToField, strings.Join(client.state.receiver.ToAllEmailsAsString(), ","),
				DecisionKey, Reject,
			)

			client.WriteCmd("471 Please try again later")
			client.state.current = StatePassive
			client.ErrorsCount++

		}

		return
	}

	if err := client.Session.AcceptMessage(client.state.receiver.message); err != nil {
		if !sendedErrorToClientIfReject(client, err) {
			client.logger.Error("cannot accept message", "err", err,
				ActionNameField, "session.AcceptMessage",
				FromField, client.state.from.String(),
				ToField, strings.Join(client.state.receiver.ToAllEmailsAsString(), ","),
				DecisionKey, Reject,
			)

			client.WriteCmd("471 Please try again later")
			client.state.current = StatePassive
			client.ErrorsCount++

		}

		return
	}

	// Let's go back to the beginning because you could have rearranged the file, but are you sure,
	// when we move it again, the message will not point to the message itself, only the headers
	// ...
	// So you have to remember where the message starts and go back there
	if !client.state.receiver.SeekMessageToBeginning(client) {
		return
	}

	qid := bytes.Buffer{}
	for _, d := range client.state.receiver.deliveries {
		// We should forward the indicator to the news
		if err := client.Session.ProcessDelivery(&client.state.receiver, d, l.reverseHostName); err != nil {
			if send := sendedErrorToClientIfReject(client, err); !send {
				client.logger.Error("cannot accept message", "err", err,
					ActionNameField, "session.ProcessDelivery",
					FromField, client.state.from.String(),
					ToField, d.To.String(),
					DecisionKey, Reject,
				)

				client.WriteCmd("471 Please try again later")
				client.state.current = StatePassive
				client.ErrorsCount++

			}
			return
		}
		//infoLog(client, "RECEIVED FROM: <%s> TO: <%s> MID: %s", d.From, d.To, d.UniqueId)
		client.logger.Info("message accepted",
			ActionNameField, "receivedMessage",
			DeliveryStatusField, DeliveryQueued,
			FromField, d.From.String(),
			ToField, d.To.String(),
			MessageUidField, d.UniqueId,
		)
		if qid.Len() > 0 {
			qid.WriteString(", ")
		}
		qid.WriteString(d.UniqueId)
	}

	/*
		// receiving the message by the target class
		if err := client.Session.ReceiveMessage(client.Text.DotReader()); err != nil {
			if sendedErrorToClientIfReject(client, err) {
				return true
			} else {
				client.WriteCmd("471 Please try again later")
				client.state.current = mtp.StatePassive
				client.DidDataRecived = true

				client.resetMessageState()
				return false
			}

		}
	*/
	client.state.current = StatePassive
	client.state.didDataReceived = true

	client.resetMessageState()
	client.WriteCmd(fmt.Sprintf(OkResponseFmt, fmt.Sprintf("message queued as %s", qid.String())))

	/*
		if server.SimpleServerDefaultMaxMessageSizeInBytes() > 0 && cap(msgBody) > server.SimpleServerDefaultMaxMessageSizeInBytes() {
			client.Text.Cmd("552 Too much mail data")
			client.resetMessageState()
		}

	*/
	return false
}

func sendedErrorToClientIfReject(client *Client, err error) (messageSended bool) {
	if re, ok := err.(RejectErr); ok {
		sendRejectError(client, re)
		return true
	} else {
		client.logger.Error("sendedErrorToClientIfReject", "err", err)
		return false
	}
}

func sendRejectError(client *Client, reErr RejectErr) {
	client.logger.Error(reErr.Error())
	client.ErrorsCount++
	client.WriteCmd(reErr.ToMtpResponse())
	if reErr.CloseConnection {
		client.logger.Error(reErr.ToMtpResponse(),
			DeliveryStatusField, DeliveryRejected,
			ActionNameField, "messageRejected",
		)
		client.WeClose = true
		client.WriteCmd("221 Bye")
	} else {
		client.logger.Error(reErr.ToMtpResponse(),
			ActionNameField, "messageRejected",
		)

	}
}

// executeCommands Passive mode, waiting for commands
func (l *Listener) executeCommands(ctx context.Context, maxErrorsPerConnection int, client *Client) {
	if msg, ok := readLine(client); !ok {
		return
	} else {
		if client.state.callFunc != nil {
			client.logger.Debug("executeCmd", "msg", msg, ActionNameField, "executeCommands:callFunc")
			client.state.callFunc(msg)
		} else {

			client.logger.Debug("executeCmd", "msg", msg, ActionNameField, "executeCommands:callRegisteredCommand")
			if l.callRegisteredCommand(msg, client) {
				return
			}
		}
		CheckErrorsCount(client, maxErrorsPerConnection)
	}
}

func (l *Listener) callRegisteredCommand(msg string, client *Client) bool {

	cmd, args := parseMessageAsCommand(msg)

	// Calling registered commands
	if cmdf, found := Commands[cmd]; found {
		cmdf(client, l.supervisor, args)
		return true
	} else {
		client.UnknownCommand(cmd, msg)
	}
	return false
}

func parseMessageAsCommand(msg string) (string, string) {

	var args, cmd string

	firstSpacePos := strings.Index(msg, " ")

	if firstSpacePos > 0 {
		cmd = strings.ToUpper(msg[:firstSpacePos])
		if firstSpacePos < len(msg) {
			args = msg[firstSpacePos:]
		}
	} else {
		cmd = strings.ToUpper(msg)
	}
	return cmd, args
}

// readLine reads a line of text from the client
func readLine(client *Client) (msg string, ok bool) {
	if line, err := client.ReadLine(); err == nil {
		return line, true
	} else {
		client.WeClose = true
		client.ErrorsCount++

		if err == io.EOF {
			client.state.disconnectReason = "Closing the connection by the client"
		} else {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				client.logger.Error("data reading error, too slow client response or timeout", "err", neterr)
			} else {
				if strings.HasSuffix(err.Error(), "read: connection reset by peer") {
					client.state.disconnectReason = "connection reset by peer"
				} else {
					client.state.disconnectReason = fmt.Sprintf("error reading client data %v", err)
				}
			}
		}
		return "", false
	}
}

// CheckErrorsCount check the number of errors and end the connection if necessary
func CheckErrorsCount(client *Client, maxErrorsPerConnection int) {
	if client.ErrorsCount > maxErrorsPerConnection {
		client.WriteCmd("500 Too many errors\n221 Bye")
		client.WeClose = true
		if client.state.disconnectReason == "" {
			client.state.disconnectReason = fmt.Sprintf("DROPPED BY TOO MANY ERRORS (%d)", client.ErrorsCount)
		}
	}
}

var addressMatch = regexp.MustCompile(`(tcp|unix)?(:[a-z0-9]+)(:[0-9]+)`)
