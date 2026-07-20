package sys

import (
	"os"
	"os/signal"
	"syscall"
	"sync"
	"log"	
)

var onExitMutex sync.Mutex

var onExitHandlers int

func init() {
	onExitHandlers = 0
}

type OnExitFunc func()

func OnExit(fn OnExitFunc) {
	onExitMutex.Lock()
	onExitHandlers++
	onExitMutex.Unlock()

	// shutdown hook
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// ctx, cancel := context.WithTimeout(context.Background(), *wait)
		// defer cancel()
		// Doesn't block if no connections, but will otherwise wait
		// until the timeout deadline.

		log.Println("Calling on exit...")

		// controller.Stop()
		fn()

		onExitMutex.Lock()
		onExitHandlers--
		onExitMutex.Unlock()

		// Optionally, you could run srv.Shutdown in a goroutine and block on
		// <-ctx.Done() if your application should wait for other services
		// to finalize based on context cancellation.

		if onExitHandlers == 0 {
			log.Println("no more exit handler, will exit")
			os.Exit(0)
		}
	}()
}
