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
// PID Parameters for Motor Speed Control

const LIDAR_RPM = 300.0        // target value
const LIDAR_DEFAULT_PWM = 0.20 // initial setting

const LIDAR_PID_KP = 0.0
const LIDAR_PID_TI = 0.0
const LIDAR_PID_TD = 0.0
const LIDAR_PID_DT = 0.0

//-----------------------------------------------------------------------------
// LIDAR Frame

const LIDAR_SOF_DELIMITER = 0xfa
const LIDAR_MIN_INDEX = 0xa0
const LIDAR_MAX_INDEX = 0xf9
const SAMPLES_PER_FRAME = 4

type LIDAR_frame struct {
	index    uint8  // 0xa0 - 0xf9 (0-89 offset, 4 samples per frame = 360 measurements)
	start    uint8  // 0xfa
	speed    uint16 // little endian, rpm = speed/64
	samples  [SAMPLES_PER_FRAME]uint32
	checksum uint16
}

//-----------------------------------------------------------------------------
// LIDAR Sample

type LIDAR_sample struct {
	no_data   bool   // No return/max range/too low of reflectivity
	too_close bool   // Object too close, possible poor reading due to proximity < 0.6m
	dist      uint16 // distance
	ss        uint16 // ??
}

// extract information from a frame sample
func (frame *LIDAR_frame) sample(i int) *LIDAR_sample {

	x := frame.samples[i]
	b0 := (x >> 0) & 0xff
	b1 := (x >> 8) & 0xff
	b2 := (x >> 16) & 0xff
	b3 := (x >> 24) & 0xff

	var sample LIDAR_sample
	sample.no_data = (b0>>7)&1 != 0
	sample.too_close = (b0>>6)&1 != 0
	sample.dist = uint16(((b0 & 0x3F) << 8) + b1)
	sample.ss = uint16((b2 << 8) + b3)
	return &sample
}

//-----------------------------------------------------------------------------

// endian flip
func flip16(data uint16) uint16 {
	return (data >> 8) | (data << 8)
}

// return the checksum for a LIDAR data frame
func calc_checksum(data [10]uint16) uint16 {
	var cs uint32
	for i := 0; i < 10; i++ {
		cs = (cs << 1) + uint32(flip16(data[i]))
	}
	cs = ((cs & 0x7fff) + (cs >> 15)) & 0x7fff
	return uint16(cs)
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
