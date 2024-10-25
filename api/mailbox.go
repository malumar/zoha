package api

import (
	"time"
)

// todoThe structure of Mailbox needs to be reduced,
// as it contains parts of the code used inside our company, which will not be published here
type Mailbox struct {
	IsAlias bool

	// if alias, here is the address where to send the messages, otherwise, where to send a copy to,
	// enter email addresses separated by a comma
	TransportTo          string
	Created              time.Time
	Account              string
	EmailLowerAscii      string
	DomainLowerAscii     string
	FullNameOfEmailOwner string

	// Salted password
	Password string

	// mount point on disk e.g. /mnt/drive2
	MountPath string

	// reserved disk space for the mailbox
	Quota uint64
	// space used by the mailbox
	UsedQuota uint64

	// outgoing mail functionality for the mailbox
	SmtpOutEnabled bool
	// incoming mail functionality for the mailbox
	SmtpInEnabled bool
	// IMAP functionality for the mailbox
	ImapEnabled bool

	ImapUnsecureOff bool
	// POP3 functionality for the mailbox
	PopEnabled     bool
	PopUnsecureOff bool
	WebmailOff     bool
	ImapSharedOff  bool
	// This option is used by Courier-IMAP in calculating access control lists. This option places the account as
	// a member of access group name. Instead of granting access rights on individual mail folders to individual accounts,
	// the access rights can be granted to an access group “name”, and all members of this group get the specified access rights.
	//
	// The access group name “administrators” is a reserved group. All accounts in the administrators group automatically
	// receive all rights to all accessible folders.
	//
	//Note
	//This option may be specified multiple times to specify that the account belongs to multiple account groups.
	ImapGroup string

	// Another option used by Courier-IMAP. Append "name" to the name of the top level virtual shared folder index file.
	// This setting restricts which virtual shared folders this account could possibly access (and that's on top of whatever
	// else the access control lists say). See the virtual shared folder documentation for more information.

	// For technical reasons, group names may not include comma, tab, "/" or "|" characters.
	ImapSharedGroup string

	// a list of email addresses that are treated as trusted senders that should never fall into spam
	WhiteList []string

	// a list of email addresses that are spammers, and always mark messages from hims as spam
	BlackList []string

	// plane name
	Packet string

	// MoveSpam
	// if true: when incoming mail is recognized as unwanted, add the SPAM prefix to the subject
	// and leave it in the current folder
	// if false:  when incoming mail is recognized as unwanted move to Spam folder
	MoveToSpam bool

	SysUid uint
	SysGid uint

	// Disabled true if you do not want to currently handle the mailbox
	Disabled bool
	// AccountDisabled  true if you do not want to currently handle any mailbox from this Account
	AccountDisabled bool

	// AutoresponderEnabled send an automatic reply to every new incoming message that is classified as wanted
	AutoresponderEnabled bool
	// AutoresponderSubject subject of autoresponder message
	// insert %t% anywhere in the string to paste the message title you are replying to, e.g.:
	// Re: %t%
	AutoresponderSubject string
	// AutoresponderBody plain text of autoresponder body
	AutoresponderBody string

	// Incoming mail host address, or where to route messages to servers
	SmtpInHost string

	// For logging for clients, for NGINX Proxy,
	// info where to direct proxy people for clients for imap and pop, addressIp
	SmtpOutHost string
	SmtpOutIp   string

	ImapIp string

	// 04.10.23
	// The IMAP server checks if it matches its hostname
	// otherwise it does not allow you to store or receive messages,
	// must be same as hostname of supervisor, because the configuration base on all hosts is the same,
	// and the supervisor himself decides which mailboxes to support
	ImapHostname string `json:"imaphn"`
	// The SMTP server that is responsible for sending the message,
	// checks if it matches its Hostname and then allows you to send the message from itself,
	// if SmptOutHostname agrees and ImapHostname is different,
	// then we redirect lmtp to ImapHostname for stored messages, then it must replace SmtpOutHost and SmtpOutIp
	SmtpOutHostname string `json:"smtpouthn"`

	// The host that is responsible for the storage
	ImapHost string

	// ImapStoreLmtp where we store messages after receiving
	// 0 or empty is the default value - so it's like using maildir on postfix
	// jeżeli nie jest puste:
	//		-  if there is a SmtpInHost, then compare SmtpInHost is equal to the host of your system
	//		and if so deliver locally - otherwise pass it on via lmtp:{sourceFrom_ImapLmtp}
	ImapStoreLmtp string

	// do not limit the number of messages sent from this mail
	UnlimitedOutgoingSmtp bool `json:"ulimitsmtp"`

	// todo internal use in our company, remove from the code
	PrivatePortSfx string `json:"ppsfx"`
}
