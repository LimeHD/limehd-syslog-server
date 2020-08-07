package lib

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Closer interface {
	CloseMessage() string
}

type Opener interface {
	Close()
}

func Notifier(openers ...Opener) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig

		fmt.Println("Stop Syslog server by signal")

		for _, opener := range openers {
			switch opener.(type) {
			case Closer:
				o := opener.(Closer)
				fmt.Println(o.CloseMessage())
			}

			opener.Close()
		}

		os.Exit(0)
	}()
}
