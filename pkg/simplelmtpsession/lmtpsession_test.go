package simplelmtpsession

import (
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/internal/example"
	"github.com/malumar/zoha/internal/example/data"
	"github.com/malumar/zoha/pkg/mtp"
	"log/slog"
	"net/smtp"
	"testing"
	"time"
)

func TestDelivery(t *testing.T) {

	go example.ZohaLmtp(func(supervisor api.MaildirSupervisor, l *slog.Logger) mtp.Session {
		return NewLmtpSession(supervisor, l)
	})
	time.Sleep(100 * time.Millisecond)
	s, err := smtp.Dial(data.LmtpListenAddress)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err := s.Hello("ninja"); err != nil {
		t.Error(err.Error())
	}

	if err := s.Mail("hello@localhost"); err != nil {
		t.Error(err.Error())
	}
	if err := s.Rcpt(data.OfficeEmail); err != nil {
		t.Error(err.Error())
	}

	if wr, err := s.Data(); err != nil {
		t.Error(err)
	} else {
		if _, err := wr.Write([]byte(msg)); err != nil {
			t.Error(err.Error())
		}
	}
	err = smtp.SendMail(data.LmtpListenAddress, nil, "hello@localhost", []string{data.OfficeEmail}, []byte(msg))
	if err != nil {

	}
}

const msg = `FCC: imap://piro-test@mail.clear-code.com/Sent
X-Identity-Key: id1
X-Account-Key: account1
From: "piro-test@clear-code.com" <piro-test@clear-code.com>
Subject: test confirmation
To: ` + data.OfficeEmail + `
Message-ID: <05c18622-f2ad-cb77-2ce9-a0bbfc7d7ad0@clear-code.com>
Date: Thu, 15 Aug 2019 14:54:37 +0900
X-Mozilla-Draft-Info: internal/draft; vcard=0; receipt=0; DSN=0; uuencode=0;
 attachmentreminder=0; deliveryformat=4
Login-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:69.0) Gecko/20100101
 Thunderbird/69.0
MIME-Version: 1.0
Content-Type: multipart/mixed;
 boundary="------------26A45336F6C6196BD8BBA2A2"
Content-Language: en-US

This is a multi-part message in MIME format.
--------------26A45336F6C6196BD8BBA2A2
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 7bit

testtest
testtest
testtest
testtest
testtest
testtest



--------------26A45336F6C6196BD8BBA2A2
Content-Type: text/plain; charset=UTF-8;
 name="sha1hash.txt"
Content-Transfer-Encoding: base64
Content-Disposition: attachment;
 filename="sha1hash.txt"

NzRjOGYwOWRmYTMwZWFjY2ZiMzkyYjEzMjMxNGZjNmI5NzhmMzI1YSAqZmxleC1jb25maXJt
LW1haWwuMS4xMC4wLnhwaQpjY2VlNGI0YWE0N2Y1MTNhYmNlMzQyY2UxZTJlYzJmZDk2MDBl
MzFiICpmbGV4LWNvbmZpcm0tbWFpbC4xLjExLjAueHBpCjA3MWU5ZTM3OGFkMDE3OWJmYWRi
MWJkYzY1MGE0OTQ1NGQyMDRhODMgKmZsZXgtY29uZmlybS1tYWlsLjEuMTIuMC54cGkKOWQ3
YWExNTM0MThlYThmYmM4YmU3YmE2ZjU0Y2U4YTFjYjdlZTQ2OCAqZmxleC1jb25maXJtLW1h
aWwuMS45LjkueHBpCjgxNjg1NjNjYjI3NmVhNGY5YTJiNjMwYjlhMjA3ZDkwZmIxMTg1NmUg
KmZsZXgtY29uZmlybS1tYWlsLnhwaQo=
--------------26A45336F6C6196BD8BBA2A2
Content-Type: application/json;
 name="manifest.json"
Content-Transfer-Encoding: base64
Content-Disposition: attachment;
 filename="manifest.json"

ewogICJtYW5pZmVzdF92ZXJzaW9uIjogMiwKICAiYXBwbGljYXRpb25zIjogewogICAgImdl
Y2tvIjogewogICAgICAiaWQiOiAiZmxleGlibGUtY29uZmlybS1tYWlsQGNsZWFyLWNvZGUu
Y29tIiwKICAgICAgInN0cmljdF9taW5fdmVyc2lvbiI6ICI2OC4wIgogICAgfQogIH0sCiAg
Im5hbWUiOiAiRmxleCBDb25maXJtIE1haWwiLAogICJkZXNjcmlwdGlvbiI6ICJDb25maXJt
IG1haWxhZGRyZXNzIGFuZCBhdHRhY2htZW50cyBiYXNlZCBvbiBmbGV4aWJsZSBydWxlcy4i
LAogICJ2ZXJzaW9uIjogIjIuMCIsCgogICJsZWdhY3kiOiB7CiAgICAidHlwZSI6ICJ4dWwi
LAogICAgIm9wdGlvbnMiOiB7CiAgICAgICJwYWdlIjogImNocm9tZTovL2NvbmZpcm0tbWFp
bC9jb250ZW50L3NldHRpbmcueHVsIiwKICAgICAgIm9wZW5faW5fdGFiIjogdHJ1ZQogICAg
fQogIH0KfQ==
--------------26A45336F6C6196BD8BBA2A2--
`
