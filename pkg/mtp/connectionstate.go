package mtp

import (
	"github.com/malumar/zoha/pkg/bitmask"
)

// NewConnectionState As mandated by RFC5321 Section 4.5.5, DSNs MUST be sent with a NULL Return-Path (or MAIL FROM) sender.
// The From: header MAILER-DAEMON@example.com is set by the receiving MTA, based on the NULL sender address.
func NewConnectionState(flags bitmask.Flag64) ConnectionState {
	return ConnectionState{
		flags:            flags,
		from:             AddressEmail{},
		didDataReceived:  false,
		current:          StatePassive,
		callFunc:         nil,
		receiver:         NewMessageReceiver(),
		disconnectReason: "",
	}
}

type ConnectionState struct {
	// Did we successfully send the FROM command?
	// Note: This does not mean that the from address is set because it may be empty!
	// Whether it goes through depends on whether you accept empty emails - as an LMTP you have to!
	// As mandated by RFC5321 Section 4.5.5, DSNs MUST be sent with a NULL Return-Path (or MAIL FROM) sender.
	// The From: header MAILER-DAEMON@example.com is set by the receiving MTA, based on the NULL sender address.
	// dlatego musi byc tutaj a nie w kliencie, poniewaz klient moze dac reset state i zaczyna
	// ndawanie wiadomosci do nowej oosby
	fromCommandPassed bool
	flags             bitmask.Flag64
	from              AddressEmail
	// whether the client sent the email content
	didDataReceived bool
	// something went wrong while sending the email content and the data may be incomplete,
	// it may also be a sign of a spammer who disconnects after sending the message
	dataCorrupted    bool
	current          State
	callFunc         func(cmd string)
	disconnectReason string
	receiver         MessageReceiver
}

// Close If you have any pointers, remove them and close all streams
func (mp ConnectionState) Close() {
	mp.callFunc = nil
	mp.receiver.Close()
}
