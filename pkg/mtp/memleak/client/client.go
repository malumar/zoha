package main

import (
	"context"
	"fmt"
	"github.com/malumar/zoha/pkg/mtp"
	"github.com/malumar/zoha/pkg/mtp/mockup"
	"math/rand"
	"net/smtp"
	"strings"
	"sync/atomic"
	"time"
)

func main() {
	coruse(1)
	time.Sleep(time.Second * 10)
	coruse(2)
	time.Sleep(time.Second * 60)
	coruse(2)
	time.Sleep(time.Second * 90)
	coruse(1)
	time.Sleep(time.Second * 10)
	coruse(1)
	time.Sleep(time.Second * 120)
	coruse(1)
	time.Sleep(time.Second * 180)
	coruse(1)
	time.Sleep(time.Second * 15)
	coruse(1)
	time.Sleep(time.Second * 60)
	coruse(2)
	time.Sleep(time.Second * 60)
	coruse(2)
	time.Sleep(time.Second * 60)
}

// przed wyjsciem bylo 20,1 mb pamieci ram
// i jest ok 508083959
// po dluszym czasie spada do 19,0 mb
func coruse(minut int) {
	// Najwieksza przepustowosc na poziomie 110-120 klientow
	for i := 0; i < 120; i++ {
		go send(100000)
	}
	// 16,6 MB o 19:53
	// 16,9 MB o 20.01 ale ile watko?
	// 17,9 MB 20.08
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(minut))
	if ctx != nil {

	}
	for {
		select {
		case <-ctx.Done():
			// ?
			cancel()
			return
		}
	}
}
func send(count int) {
	for i := 0; i < count; i++ {

		con, err := smtp.Dial(mtp.DefaultAddress)
		if dropConnection() {
			return
		}
		if err != nil && !strings.HasPrefix(err.Error(), "221 ") {
			fmt.Println("dial error", err)
			continue
		}
		if dropConnection() {
			return
		}

		if err := con.Hello("abcd"); err != nil && !strings.HasPrefix(err.Error(), "221 ") {
			fmt.Println("hello error", err)
			continue

		}
		if dropConnection() {
			return
		}
		if err := con.Mail(mockup.RemoteUser1); err != nil && !strings.HasPrefix(err.Error(), "221 ") {
			fmt.Println("mail error", err)
			continue

		}

		if dropConnection() {
			return
		}
		if err := con.Rcpt(mockup.LocalUser1Email); err != nil && !strings.HasPrefix(err.Error(), "221 ") {
			fmt.Println("rcpt error", err)
			continue

		}
		if dropConnection() {
			return
		}
		wc, err := con.Data()
		if err != nil {
			fmt.Println("data error", err)
			continue

		}

		if dropConnection() {
			return
		}
		bc, err := wc.Write([]byte(mockup.Message1))
		if err != nil {
			fmt.Println("write error", err)
			continue

		}
		if dropConnection() {
			return
		}
		if bc != len(mockup.Message1) {
			fmt.Println("length error", err)
			continue

		}
		if dropConnection() {
			return
		}

		if err := wc.Close(); err != nil && !strings.HasPrefix(err.Error(), "221 ") {
			fmt.Println("close error", err)
			continue
		}

		if dropConnection() {
			return
		}
		con.Quit()

	}
}

var breakedConnections int64

func dropConnection() bool {
	max := 1000
	min := 0
	val := rand.Intn(max-min) + min
	if val == 100 {
		breakNo := atomic.AddInt64(&breakedConnections, 1)
		fmt.Println("I'm breaking the connection, broken connection No.", breakNo)
		return true
	}
	return false
}
