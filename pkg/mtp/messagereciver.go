package mtp

import (
	"context"
	"fmt"
	"github.com/malumar/filebuf"
	"github.com/malumar/zoha/pkg/mtp/headers"
	"io"
	"net/mail"
	"time"
)

const (
	PrivateHeader = iota
	AllHeaders
	MailHeader
)

const (
	bufSize = 1024 // kb
)

type MessageReceiverProxy interface {
	InitialMessageSize() int
	GetBodyOffset() int64
	GetMessage() *mail.Message
	GetBuffer() *filebuf.Buf
	MoveBufferPosToStart() error
	ArrivalTime() time.Time
	MoveBufferPosToMessageStartLine() error
}

type DeliveryProxy interface {
	GetUniqueId() string
	GetFrom() AddressEmail
	GetTo() AddressEmail
	GetMailbox() Maildir
	GetCustomMailbox() string
	GetImapFlags() ImapFlag
	IsRecent() bool
	Filesize() int
	AddHeader(name, value string)
	DelHeader(name string)
	SetHeader(name string, value string)
}

func NewMessageReceiver() MessageReceiver {
	return MessageReceiver{
		deliveries: make([]Delivery, 0),
		arrival:    time.Now(),
	}
}

type MessageReceiver struct {
	messageSize int
	// data przybycia do nas
	arrival    time.Time
	deliveries []Delivery
	buf        *filebuf.Buf
	message    *mail.Message
	// bodyOffset specifies the position where the body of the message begins (there are headers at the beginning)
	bodyOffset int64
}

func (mr *MessageReceiver) ToAllEmails() []AddressEmail {
	e := make([]AddressEmail, len(mr.deliveries))
	for i, d := range mr.deliveries {
		e[i] = d.To
	}

	return e
}

func (mr *MessageReceiver) ToAllEmailsAsString() []string {
	e := make([]string, len(mr.deliveries))
	for i, d := range mr.deliveries {
		e[i] = d.To.String()
	}

	return e
}
func (mr *MessageReceiver) ArrivalTime() time.Time {
	return mr.arrival
}

// InitialMessageSize The size of the message right after downloading,
// without the extra headers you created
func (mr *MessageReceiver) InitialMessageSize() int {
	return mr.messageSize
}
func (mr *MessageReceiver) GetMessage() *mail.Message {
	return mr.message
}

func (mr *MessageReceiver) GetBodyOffset() int64 {
	return mr.bodyOffset
}

func (mr *MessageReceiver) GetBuffer() *filebuf.Buf {
	return mr.buf
}

func (mr *MessageReceiver) SeekMessageToBeginning(client *Client) bool {
	return mr.SeekMessage(client, 0, io.SeekStart)
}

func (mr *MessageReceiver) SeekMessageToBody(client *Client) bool {
	return mr.SeekMessage(client, mr.bodyOffset, io.SeekStart)
}

// MoveBufferPosToStart move to the beginning of the file
func (mr *MessageReceiver) MoveBufferPosToStart() error {
	if _, err := mr.buf.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return nil
}

// MoveBufferPosToMessageStartLine move to the beginning of the line where the message begins
func (mr *MessageReceiver) MoveBufferPosToMessageStartLine() error {
	if _, err := mr.buf.Seek(mr.bodyOffset, io.SeekStart); err != nil {
		return err
	}
	return nil
}

func (mr *MessageReceiver) SeekMessage(client *Client, offset int64, whence int) bool {
	// just in case, we restart again
	if _, err := mr.buf.Seek(offset, whence); err != nil {
		client.WriteCmd("471 Please try again later")
		client.state.current = StatePassive
		client.ErrorsCount++
		return false
	}

	return true
}

func (mr *MessageReceiver) RecipientsCount() int {
	return len(mr.deliveries)
}

// Close You clean everything here
func (mr *MessageReceiver) Close() {
	for _, d := range mr.deliveries {
		d.Close()
	}

	mr.deliveries = nil

	if mr.buf != nil {
		mr.buf.Close()
		mr.buf = nil
	}
	if mr.message != nil {
		mr.message.Body = nil
		mr.message.Header = nil
		mr.message = nil
	}
}

func (mr *MessageReceiver) AcceptRecipient(currentHostname string, from, to AddressEmail, client *Client) bool {
	// Client REV_HOST
	// Client Time Connected
	// Client HELLO
	// SERVER REV_HOST
	// SERVER APP_NAME
	// SESSION_NUMBER
	// Client connect time) error {
	for _, key := range mr.deliveries {
		if key.To == to {
			client.WriteCmd(fmt.Sprintf("500 recipient \"%s\" added twice", to))
			client.ErrorsCount++
			return false
		}
	}

	if err := client.Session.AcceptRecipient(to); err != nil {

		if sended := sendedErrorToClientIfReject(client, err); !sended {
			client.logger.Error("AcceptMessageFromEmail", "err", err,
				ActionNameField, "acceptRecipient",
			)
			client.WriteCmd("451 4.7.1 Unable to complete command")
			client.ErrorsCount++
		}
		return false

	} else {
		client.logger.Info("recipient accepted",
			FromField, from.String(),
			ToField, to.String(),
			ActionNameField, "acceptRecipient",
		)
	}

	delivery := NewDeliveryFromClient(client, from, to)

	// the actual name of the mailbox - if RCPT TO directed to an alias
	// then here it should be an actual email, not an alias!
	//delivery.AddHeaderAs(headers.DeliveredTo, to.String(), true)
	DeliveryAddHeaderAs(&delivery, headers.DeliveredTo, to.String(), true)

	DeliveryAddHeaderAs(&delivery, "Received",
		fmt.Sprintf("by %s [%s] with SMTP id %s; %s",
			//from,
			client.ListenerReverseHostname,
			client.GetInfo().RemoteAddr,
			delivery.UniqueId,
			time.Now().Format(time.RFC1123Z),
		),
		false,
	)
	mr.deliveries = append(mr.deliveries, delivery)
	return true
}

func (mr *MessageReceiver) AcceptData(ctx context.Context, c *Client, r io.Reader) (err error) {
	var msgSize int
	if mr.buf != nil {
		mr.buf.Close()
		mr.buf = nil
	}
	mr.buf, msgSize, err = mr.readAllData(ctx, c, r)

	if err != nil {
		return err
	}

	mr.messageSize = msgSize

	// if 0 then not limited
	if c.Session.MaxMessageSize() > 0 {
		if uint64(msgSize) > c.Session.MaxMessageSize() {
			return NewRejectErr(QuotaLowRejectErr)
		}
	}

	if err := mr.buf.SwitchToRead(); err != nil {
		c.logger.Error("read buffer switch error", "err", err.Error())
		return err
	}

	if _, err := mr.buf.Seek(0, io.SeekStart); err != nil {
		c.logger.Error("error switching buffer to top", "err", err.Error())
		return NewRejectErr(TemporaryServerRejectErr)

	}

	// malformed MIME header: missing colon: \"This is the email body\""
	if m, err := mail.ReadMessage(mr.buf); err != nil {
		c.logger.Error("error reading mail", "err", err.Error())
		return NewRejectErrMessage(TemporaryServerRejectErr, "bad mime message format")
	} else {
		mr.message = m
		mr.bodyOffset = mr.buf.Offset()
	}

	// From now on, we have a message and we should decide what to do with it next
	return nil
}

// Load all data into the cache
func (mr *MessageReceiver) readAllData(ctx context.Context, c *Client, r io.Reader) (*filebuf.Buf, int, error) {
	size := 0
	c.logger.Debug("Start receiving data")
	fb := filebuf.New(1024*1024*5, true, true)
	if err := fb.SwitchToWrite(); err != nil {
		c.logger.Error("error switching to write", "err", err.Error())
		return nil, 0, err
	}

	wc, err := io.Copy(&fb, r)
	if err == nil || err == io.EOF {

	} else {
		c.state.dataCorrupted = true
		if err != io.ErrUnexpectedEOF {
			c.logger.Error("error reading data", "err", err.Error())
			return nil, 0, err
		}
	}
	// 0 bytes??
	// well we return the error of the wrong quota
	if wc == 0 {
		return nil, 0, NewRejectErr(QuotaLowRejectErr)
	}
	return &fb, size, nil

	/*
		buf := make([]byte, bufSize)
		loop := true
		for loop {
			rb, readEr := r.Read(buf)
			if readEr == nil || readEr == io.EOF {
				if rb <= 0 {
					break
				}
			} else {
				// Probably the client closed the channel without waiting for our answer the question
				// is what we should do about it, but I would suggest adding to detached but marking as corrupted,
				// and such files should be marked as spamis what we should do about it,
				// but I would suggest adding to detached but marking as corrupted;
				// and such files should be marked as spam
				c.state.dataCorrupted = true
				if readEr == io.ErrUnexpectedEOF {
					break
				}
				fatalLog(c, "message body read error %v", readEr)
				return nil, 0, readEr
			}

			if wb, errW := fb.Write(buf[:rb]); errW != nil {
				return nil, size, errW
			} else {
				size += wb
				if wb != rb {
					return nil, size, fmt.Errorf("read %d bytes and write %d", rb, wb)
				}
			}

		}

		// 0 bytes??
		// well we return the error of the wrong quota
		if size == 0 {
			return nil, 0, mtp.NewRejectErr(mtp.QuotaLowRejectErr)
		}

		return &fb, size, nil
	*/
}
