// A master Sender server, forwarding messages from Zoha-Sender-Client to the main MTA
// (e.g. the main instance of postfix), here the whole fun starts again and again
// the MTA can forward the message to ZOHA-LMTP
package main

import (
	"flag"
	"github.com/malumar/mslices"
	"io"
	"log/slog"
	"net"
	"strings"
)

var logger = slog.With("package", "zoha-sender-server")

var allowIps = make([]string, 0)

func main() {
	var allowlist mslices.SliceOfFlag[string]

	var localAddr = flag.String("l", "0.0.0.0:9999", "listen address")
	var remoteAddr = flag.String("d", "127.0.0.1:25", "proxying address (e.g) postfix instance")
	flag.Var(&allowlist, "ip", "IP address"+
		" that can be allowed to connect, you can use simple expressions and ? to match the results"+
		" (asterisk, anything, question mark, any character), "+
		"You can specify this parameter multiple times to allow access from multiple IP addresses")

	flag.Parse()

	for i, item := range allowlist {
		slog.Info("Enable listing match to", "pattern", item)
		allowlist[i] = item + ":"
	}

	logger.Info("Listening & Proxying", "listen", *localAddr, "proxying", *remoteAddr)
	listener, err := net.Listen("tcp", *localAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		logger.Info("new connection", "remote", conn.RemoteAddr())
		if err != nil {
			logger.Error("error accepting connection", "err", err)
			continue
		}
		go func() {
			defer conn.Close()
			var found bool
			for _, item := range allowlist {
				if strings.HasPrefix(conn.RemoteAddr().String(), item) {
					found = true
					break
				}
			}
			if !found {
				logger.Error("incoming connection rejected", "client", conn.RemoteAddr().String())
				return
			}

			conn2, err := net.Dial("tcp", *remoteAddr)
			if err != nil {
				logger.Error("error dialing to smtp server", "dest", *remoteAddr, "err")
				return
			}
			defer conn2.Close()
			closer := make(chan struct{}, 2)
			go copy(closer, conn2, conn)
			go copy(closer, conn, conn2)
			<-closer
			logger.Info("connection completed", "client", conn.RemoteAddr())
		}()
	}
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
