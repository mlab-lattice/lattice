package util

import (
	"os"
	"os/signal"
	"syscall"
)

func SetupSignalHandler() <-chan struct{} {
	stopCh := make(chan struct{})
	signalCh := make(chan os.Signal, 2)

	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		close(stopCh)
		<-signalCh
		os.Exit(1)
	}()

	return stopCh
}
