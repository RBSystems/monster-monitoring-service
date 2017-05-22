package salt

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/byuoitav/monster-monitoring-service/badger"
)

func Listen() {

}

func Start(timer chan bool) {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	go Wait(signals, timer)
	<-timer
	log.Printf("SIGTERM interrupt detected. Exiting...")

	badger.Store().Close()
	(*Connection().Connection).Close()
	os.Exit(0)

}

func Wait(signals chan os.Signal, timer chan bool) {

	<-signals
	timer <- true
}
