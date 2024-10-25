package api

type MaildirStorage interface {
	// GenerateNextSpooledMessageName Generate a new email ID to be stored in spool
	GenerateNextSpooledMessageName() string

	// GetAbsoluteMaildirPath build maildir path prefix
	GetAbsoluteMaildirPath(mb *Mailbox) string
}

type Supervisor interface {
	// Hostname of your node, this name must coincide with the value of Mailbox.ImapHostname,
	// if it is different, we will not accept messages e.g. via LMTP,
	// because it means that it is intended for another node and this should be directed
	// to the appropriate LMTP at the Postfix level
	Hostname() string

	// MailerDaemonEmailAddress returns what the MAILER_DAEMON address is,
	// i.e. in the case of empty email addresses sent by Postfix,
	// it must be written in lowercase and in ASCII format during the returns
	MailerDaemonEmailAddress() string

	// Authorization return immutable mailbox struct if it is enabled
	// @service which login service
	// @useSsl whether the client is using an SSL connection
	Authorization(username string, password string, service Service, useSsl Maybe) *Mailbox

	// FindMailbox return immutable mailbox struct or nil, don't check it is enabled
	FindMailbox(name string) *Mailbox

	// IsLocalDomain Is the indicated domain in our resources? More specifically,
	// whether we host at least one mailbox on this domain
	//
	// @return if you have problems with the database (e.g. no connection)
	// or you suspect that your configurations are out of sync, return Maybe.DontKnow,
	// this will cause us to ask the sender to resend the shipment at a later date
	IsLocalDomain(asciiDomainNameLowerCase string) Maybe

	// IsLocalEmail Is the email address local, do not check other things,
	// the question is only whether this address is supported by our server
	//
	// @return if you have problems with the database (e.g. no connection)
	// or you suspect that your configurations are out of sync, return Maybe.DontKnow,
	// this will cause us to ask the sender to resend the shipment at a later date
	IsLocalEmail(emailAsciiLowerCase string) Maybe

	// MainSenderNode main host and port of zoha-sender-server
	// the host must be accessible from each ZOHA-LMTP instance, so it must be a public or tunneled host
	// ZohaSenderClient uses MainSenderNode as an intermediary node to send messages directed to aliases
	// or in the case of an autoresponder
	MainSenderNode() string
}

// MaildirSupervisor is a Supervisor with mail storage support
type MaildirSupervisor interface {
	MaildirStorage
	Supervisor
}
