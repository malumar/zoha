package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/internal/example"
	"github.com/malumar/zoha/pkg/mtp"
	"github.com/malumar/zoha/pkg/proxy"
	"github.com/malumar/zoha/pkg/simplelmtpsession"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
)

var listenAddress string
var doRun bool

func main() {
	flag.Usage = usage
	flag.StringVar(&listenAddress, "address", "127.0.0.1:10800",
		"address and port on which the server listens for queries from nginx")
	flag.BoolVar(&doRun, "run", false, "start server")

	_ = flag.NewFlagSet("run", flag.ExitOnError)
	flag.Parse()

	ctx, cancelFunc := context.WithCancel(context.Background())

	srv := proxy.New(listenAddress,
		example.NewDummySupervisor(ctx, func(supervisor api.MaildirSupervisor, l *slog.Logger) mtp.Session {
			return simplelmtpsession.NewLmtpSession(supervisor, l)
		}),
	)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		slog.Info("ctrl-c pressed")
		close(quit)
		cancelFunc()
		os.Exit(0)

	}()

	if err := srv.Run(); err != nil {
		slog.Error(err.Error())
	}
}

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "Usage: %s [OPTIONS] command ...\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(w, "\nCommands:\n")
	fmt.Fprintf(w, " run\n \trun server\n")
	fmt.Fprintf(w, "\t\n")
	fmt.Fprintf(w, "\nOptions:\n")
	flag.PrintDefaults()
}
