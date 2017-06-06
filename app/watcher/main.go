package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"watcher"
)

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("watcher %s\n", Version)
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	w := watcher.NewWatcher(watcher.NewCfg(Version))

	w.Run()
	<-signalChan
	w.Exit()
}
