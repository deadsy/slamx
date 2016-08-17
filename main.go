//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------

package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/deadsy/slamx/lidar"
	"github.com/deadsy/slamx/view"
)

//-----------------------------------------------------------------------------

func main() {

	lidar0, err := lidar.Open("lidar0", lidar_serial, lidar_pwm)
	if err != nil {
		log.Fatal("unable to open lidar device")
	}

	view0, err := view.Open("view0")
	if err != nil {
		log.Fatal("unable to open view window")
	}

	quit := make(chan bool)
	wg := &sync.WaitGroup{}

	// start the LIDAR process
	wg.Add(1)
	go lidar0.Process(quit, wg)

	// start the view window
	wg.Add(1)
	go view0.Process(quit, wg)

	running := true

	// capture Ctrl-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Printf("captured %v, exiting", sig)
			running = false
		}
	}()

	for running {
		time.Sleep(1 * time.Second)
		log.Printf("main() timeout")
	}

	close(quit)
	wg.Wait()

	lidar0.Close()
	view0.Close()

	os.Exit(0)
}

//-----------------------------------------------------------------------------
