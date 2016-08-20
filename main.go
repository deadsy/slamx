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
	scan_ch := make(chan lidar.Scan_2D)
	go lidar0.Process(quit, wg, scan_ch)

	angle := float32(0)

	// run the event loop
	running := true
	for running {
		select {
		case scan := <-scan_ch:
			log.Printf("rxed %d", len(scan.Samples))
			//view0.Render(&scan)
			view0.Render2(angle)
			angle += 1
		default:
			running = view0.Events()
			view0.Delay(30)
		}
	}

	// stop all go routines
	close(quit)
	wg.Wait()

	lidar0.Close()
	view0.Close()

	os.Exit(0)
}

//-----------------------------------------------------------------------------
