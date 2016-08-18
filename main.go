//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------

package main

import (
	"log"
	"os"
	"sync"

	"github.com/deadsy/slamx/lidar"
	"github.com/deadsy/slamx/view"
)

//-----------------------------------------------------------------------------

func main() {

	view0, err := view.Open("view0")
	if err != nil {
		log.Fatal("unable to open view window")
	}

	lidar0, err := lidar.Open("lidar0", lidar_serial, lidar_pwm)
	if err != nil {
		log.Fatal("unable to open lidar device")
	}

	quit := make(chan bool)
	wg := &sync.WaitGroup{}

	// start the LIDAR process
	wg.Add(1)
	go lidar0.Process(quit, wg)

	running := true

	angle := 0
	for running {
		view0.Delay(10)
		view0.Render(float32(angle))
		angle += 1
	}

	// stop all go routines
	close(quit)
	wg.Wait()

	lidar0.Close()
	view0.Close()

	os.Exit(0)
}

//-----------------------------------------------------------------------------
