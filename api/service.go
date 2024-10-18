package api

type Service int

const (
	UnknownService Service = iota
	SmtpInService
	SmtpOutService
	ImapService
	PopService
	Kwark
)
