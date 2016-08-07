//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------

package main

import (
	"log"
	"time"

	"github.com/deadsy/slamx/lidar"
)

//-----------------------------------------------------------------------------

func main() {

	lidar, err := lidar.Open("lidar0", "/dev/ttyUSB0", "gpio17")
	if err != nil {
		log.Fatal("unable to open lidar device")
	} else {
		defer lidar.Close()
	}

	// start the LIDAR process
	go lidar.Process()

	for true {
		time.Sleep(1 * time.Second)
		log.Printf("main() timeout\n")
	}
}

//-----------------------------------------------------------------------------
