//-----------------------------------------------------------------------------
/*

Driver for Neato XV11 LIDAR Unit

* Control the Motor using PWM to give constant RPM
* Read the serial stream from the LIDAR and repackage it as range data
* Feed the range data as a stream messages to another Go routine

*/
//-----------------------------------------------------------------------------

package lidar

import (
	"fmt"
	"log"
	"time"

	"github.com/deadsy/slamx/pwm"
	"github.com/tarm/serial"
)

//-----------------------------------------------------------------------------

type LIDAR struct {
	Name string
	port *serial.Port
	pwm  *pwm.PWM
}

//-----------------------------------------------------------------------------

func (lidar *LIDAR) stop() {
	log.Printf("lidar.stop() %s\n", lidar.Name)
	// turn the motor off
	lidar.pwm.Set(0.0)
}

//-----------------------------------------------------------------------------

func Open(name, port_name, pwm_name string) (*LIDAR, error) {
	var lidar LIDAR
	lidar.Name = name

	log.Printf("lidar.Open() %s serial=%s pwm=%s\n", lidar.Name, port_name, pwm_name)

	// open the serial port
	cfg := &serial.Config{Name: port_name, Baud: 115200}
	port, err := serial.OpenPort(cfg)
	if err != nil {
		log.Printf("unable to open serial port %s\n", port_name)
		return nil, err
	}
	lidar.port = port

	// open the pwm channel
	pwm, err := pwm.Open(fmt.Sprintf("%s_pwm", lidar.Name), pwm_name, 0.0)
	if err != nil {
		log.Printf("unable to open pwm channel %s\n", pwm_name)
		return nil, err
	}
	lidar.pwm = pwm

	return &lidar, nil
}

//-----------------------------------------------------------------------------

func (lidar *LIDAR) Close() error {
	log.Printf("lidar.Close() %s \n", lidar.Name)

	lidar.stop()
	lidar.pwm.Close()

	err := lidar.port.Flush()
	if err != nil {
		log.Printf("error flushing serial port\n")
		return err
	}

	err = lidar.port.Close()
	if err != nil {
		log.Printf("error closing serial port\n")
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------

func (lidar *LIDAR) Process() {
	log.Printf("lidar.Process() %s\n", lidar.Name)

	for true {
		time.Sleep(100 * time.Millisecond)
		log.Printf("lidar.Process() %s timeout\n", lidar.Name)
	}
}

//-----------------------------------------------------------------------------
