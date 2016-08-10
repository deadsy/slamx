//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deadsy/slamx/lidar"
)

//-----------------------------------------------------------------------------

const lidar_serial = "/dev/serial0"
const lidar_pwm = "21"

//-----------------------------------------------------------------------------

func cleanup() {
	log.Printf("cleanup()")
}

//-----------------------------------------------------------------------------

func main() {

	lidar, err := lidar.Open("lidar0", lidar_serial, lidar_pwm)
	if err != nil {
		log.Fatal("unable to open lidar device")
	}

	// start the LIDAR process
	go lidar.Process()

	running := true

	// capture Ctrl-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Printf("captured %v, exiting", sig)
			cleanup()
			running = false
		}
	}()

	for running {
		time.Sleep(1 * time.Second)
		log.Printf("main() timeout")
	}

	lidar.Close()
	os.Exit(0)
}

//-----------------------------------------------------------------------------
