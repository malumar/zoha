package mtp

import (
	"github.com/malumar/domaintools"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/emailutil"
	"strings"
)

/*
What domain name to use in outbound mail

The myorigin parameter specifies the domain that appears in mail that is posted on this machine. The default is to use the local machine name, $myhostname, which defaults to the name of the machine. Unless you are running a really small site, you probably want to change that into $mydomain, which defaults to the parent domain of the machine name.

For the sake of consistency between sender and recipient addresses, myorigin also specifies the domain name that is appended to an unqualified recipient address.

Examples (specify only one of the following):

/etc/postfix/main.cf:

myorigin = $myhostname (default: send mail as "user@$myhostname")
myorigin = $mydomain   (probably desirable: "user@$mydomain")

http://www.postfix.org/BASIC_CONFIGURATION_README.html
===

You need to have 2 proper lines in your main.cf file (/etc/postfix/main.cf):

mydomain = mydomain1.net (while mydomain1.net is your domain)

myorigin = $mydomain

The hostname part is usable for the MX part:

myhostname = mx.mydomain1.net (which will give you the hostname and )
myorigin = $myhostname

As mandated by RFC5321 Section 4.5.5, DSNs MUST be sent with a NULL Return-Path (or MAIL FROM) sender.
The From: header MAILER-DAEMON@example.com is set by the receiving MTA, based on the NULL sender address.

*/

type EmailAddressType int

func ValidEmail(value string) api.Maybe {
	if value == "" {
		return api.DontKnow
	}

	if _, h, err := emailutil.ExtractEmail(domaintools.SafeAsciiDomainName(value)); err != nil {
		return api.No
	} else {
		if strings.Index(h, ".") == -1 {
			return api.DontKnow
		}
	}
	return api.Yes
}
