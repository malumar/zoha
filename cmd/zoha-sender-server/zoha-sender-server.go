// Zoha-Sender-Server: forwards messages from Zoha-Sender-Client to the main MTA (e.g., Postfix instance)

package main

import (
	"flag"
	"github.com/malumar/mslices"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
)

var logger = slog.With("package", "zoha-sender-server")

func main() {
	var allowlist mslices.SliceOfFlag[string]

	localAddr := flag.String("l", "0.0.0.0:9999", "listen address")
	remoteAddr := flag.String("d", "127.0.0.1:25", "proxying address (e.g., Postfix instance)")
	flag.Var(&allowlist, "ip", "Allowed IP addresses with optional wildcards (?) for single character matching. "+
		"Multiple IP addresses can be specified for access control.")

	flag.Parse()
	formatAllowlist(&allowlist)

	logger.Info("Starting server", "listening", *localAddr, "proxying", *remoteAddr)
	listener, err := net.Listen("tcp", *localAddr)
	if err != nil {
		logger.Error("failed to start listening", "error", err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("connection accept error", "error", err)
			continue
		}
		go handleConnection(conn, *remoteAddr, allowlist)
	}
}

func formatAllowlist(allowlist *mslices.SliceOfFlag[string]) {
	for i, item := range *allowlist {
		logger.Info("Enabled IP pattern", "pattern", item)
		(*allowlist)[i] = item + ":"
	}
}

func handleConnection(conn net.Conn, remoteAddr string, allowlist mslices.SliceOfFlag[string]) {
	defer conn.Close()
	clientIP := conn.RemoteAddr().String()

	if !isAllowedIP(clientIP, allowlist) {
		logger.Error("connection rejected", "client", clientIP)
		return
	}

	conn2, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		logger.Error("failed to connect to MTA", "destination", remoteAddr, "error", err)
		return
	}
	defer conn2.Close()

	logger.Info("Connection accepted", "client", clientIP)
	closer := make(chan struct{}, 2)
	go proxyData(closer, conn2, conn)
	go proxyData(closer, conn, conn2)
	<-closer
	logger.Info("Connection closed", "client", clientIP)
}

func isAllowedIP(clientIP string, allowlist mslices.SliceOfFlag[string]) bool {
	for _, pattern := range allowlist {
		if strings.HasPrefix(clientIP, pattern) {
			return true
		}
	}
	return false
}

func proxyData(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{}
}
