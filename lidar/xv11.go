//-----------------------------------------------------------------------------
/*

Driver for Neato XV11 LIDAR Unit

* Control the Motor using PWM to give constant RPM
* Read the serial stream from the LIDAR and repackage it as range data
* Feed the range data as messages to another Go routine

Baud Rate:
22 bytes/frame
90 frames/rev
300 revs/min
22 * 90 * 300 = 594000 bytes/min = 9900 bytes/sec
9900 * 11 bits/byte = 108900 bits/sec
So: The serial port runs at 115200 baud to keep up.

Serial Port:
The LIDAR is powered with +5V but the Rx/Tx lines ares 3.3V.
That's good because we can plug them directly into the RPi.

Motor Speed:
300 rpm is a good target speed giving a 5 Hz 360 degree scan.
Experimentally 3.11V @ 100% gives about 300 rpm.
Other voltages/duty cycles can be guessed at from that.

Compatability:
This driver has been tested against a v2.6 unit.
It probably works with v2.4 units.
It won't work with v2.1 units.

XV11 LIDAR Boot Output:
"""
Piccolo Laser Distance Scanner
Copyright (c) 2009-2011 Neato Robotics, Inc.
All Rights Reserved

Loader  V2.5.15295
CPU     F2802x/c001
Serial  KSH34313AA-0140063
LastCal [5371726C]
Runtime V2.6.15295
"""

References:
https://xv11hacking.wikispaces.com/LIDAR+Sensor

*/
//-----------------------------------------------------------------------------

package lidar

import (
	"fmt"
	"log"
	"time"

	"github.com/deadsy/slamx/pid"
	"github.com/deadsy/slamx/pwm"
	"github.com/tarm/serial"
)

//-----------------------------------------------------------------------------

type LIDAR struct {
	Name        string
	port        *serial.Port
	pwm         *pwm.PWM
	pid         *pid.PID
	rpm         float32     // measured rpm
	pid_on      bool        // is the PID turned on?
	frame       LIDAR_frame // frame being read from serial
	ofs         int         // offset into frame data
	good_frames uint        // good frames rx-ed
	bad_frames  uint        // bad frames rx-ed (invalid checksum)
}

//-----------------------------------------------------------------------------

const LIDAR_READ_PERIOD = 50   // read lidar frames every N ms
const LIDAR_MOTOR_PERIOD = 200 // update the motor pwm every N ms

//-----------------------------------------------------------------------------
// PID Parameters for Motor Speed Control

const LIDAR_RPM = 300.0          // target rpm
const LIDAR_RPM_SHUTDOWN = 330.0 // shutdown limit

const PID_PERIOD = float32(LIDAR_MOTOR_PERIOD) / 1000.0
const PID_KP = 0.0
const PID_KI = 0.0
const PID_KD = 0.0
const PID_IMIN = -1.0
const PID_IMAX = 1.0
const PID_OMIN = 0.0
const PID_OMAX = 0.5

//-----------------------------------------------------------------------------
/*
LIDAR Frame

A full revolution will yield 90 packets, containing 4 consecutive readings each.
This amounts to a total of 360 readings (1 per degree)
The length of a packet is 22 bytes.

Each packet is organized as follows:
<start> <index> <speed_L> <speed_H> [Data 0] [Data 1] [Data 2] [Data 3] <checksum_L> <checksum_H>

<start> is always 0xFA
<index >is the index byte in the 90 packets, going from 0xA0 (packet 0, readings 0 to 3) to 0xF9 (packet 89, readings 356 to 359).
<speed> is a two-byte information, little-endian. It represents the speed, in 64th of RPM (aka value in RPM represented in fixed point, with 6 bits used for the decimal part).
<data n> are the 4 readings. Each one is 4 bytes long, and organized as follows:

byte 0 : <distance 7:0>
byte 1 : <"invalid data" flag> <"strength warning" flag> <distance 13:8>
byte 2 : <signal strength 7:0>
byte 3 : <signal strength 15:8>
*/

const LIDAR_FRAME_SIZE = 22
const LIDAR_SAMPLE_SIZE = 4

const LIDAR_SOF_DELIMITER = 0xfa
const LIDAR_MIN_INDEX = 0xa0
const LIDAR_MAX_INDEX = 0xf9

const LIDAR_START_OFS = 0
const LIDAR_INDEX_OFS = 1
const LIDAR_RPM_OFS = 2
const LIDAR_SAMPLE_OFS = 4
const LIDAR_CHECKSUM_OFS = 20
const LIDAR_END_OFS = 21

type LIDAR_frame struct {
	ts   time.Time               // timestamp
	data [LIDAR_FRAME_SIZE]uint8 // frame data
}

// return the uint16 at an offset in the frame
func (frame *LIDAR_frame) get_uint16(ofs int) uint16 {
	return uint16(frame.data[ofs]) + (uint16(frame.data[ofs+1]) << 8)
}

// return the checksum of a frame
func (frame *LIDAR_frame) checksum() uint16 {
	var cs uint32
	for i := 0; i < LIDAR_CHECKSUM_OFS; i += 2 {
		cs = (cs << 1) + uint32(frame.get_uint16(i))
	}
	cs = ((cs & 0x7fff) + (cs >> 15)) & 0x7fff
	return uint16(cs)
}

// return the rpm of the LIDAR
func (frame *LIDAR_frame) rpm() float32 {
	return float32(frame.get_uint16(LIDAR_RPM_OFS)) / 64.0
}

// return the base angle of the samples
func (frame *LIDAR_frame) angle() int {
	return 4 * (int(frame.data[LIDAR_INDEX_OFS]) - LIDAR_MIN_INDEX)
}

// process a received lidar frame
func (lidar *LIDAR) process_frame() {
	f := &lidar.frame
	log.Printf("rpm %f theta %d", f.rpm(), f.angle())
	// store the rpm for the PID process value
	lidar.rpm = f.rpm()

	s0 := f.sample(0)
	s1 := f.sample(1)
	s2 := f.sample(2)
	s3 := f.sample(3)
	log.Printf("%d %d %d %d", s0.dist, s1.dist, s2.dist, s3.dist)
}

// receive a lidar frame from a buffer
func (lidar *LIDAR) rx_frame(buf []byte, ts time.Time) {
	// We look for a start of frame and a valid index to mark a frame.
	// We may get some false positives, but they will be weeded out with bad checksums.
	// Once we sync with the frame cadence we should be good.
	f := &lidar.frame
	for _, c := range buf {
		f.data[lidar.ofs] = c
		if lidar.ofs == LIDAR_START_OFS {
			// looking for start of frame
			if c == LIDAR_SOF_DELIMITER {
				// ok - set the timestamp
				f.ts = ts
				// now read the index
				lidar.ofs += 1
			}
		} else if lidar.ofs == LIDAR_INDEX_OFS {
			// looking for a valid index
			if c >= LIDAR_MIN_INDEX && c <= LIDAR_MAX_INDEX {
				// ok - now read the frame body
				lidar.ofs += 1
			} else {
				// not a frame - keep looking
				lidar.ofs = LIDAR_START_OFS
			}
		} else if lidar.ofs == LIDAR_END_OFS {
			// validate checksum
			calc_cs := f.checksum()
			frame_cs := f.get_uint16(LIDAR_CHECKSUM_OFS)
			if calc_cs == frame_cs {
				// good frame - process it
				lidar.good_frames += 1
				lidar.process_frame()
			} else {
				// bad frame
				lidar.bad_frames += 1
			}
			// reset for the next frame
			lidar.ofs = LIDAR_START_OFS
		} else {
			// reading the frame body
			lidar.ofs += 1
		}
	}
}

//-----------------------------------------------------------------------------
// LIDAR Sample

type LIDAR_sample struct {
	no_data   bool   // No return/max range/too low of reflectivity
	too_close bool   // object too close
	dist      uint16 // distance in mm
	ss        uint16 // signal strength
}

// get the i-th sample from the frame, i = 0..3
func (frame *LIDAR_frame) sample(i int) *LIDAR_sample {
	ofs := LIDAR_SAMPLE_OFS + (i * LIDAR_SAMPLE_SIZE)
	b0 := frame.data[ofs]
	b1 := frame.data[ofs+1]
	b2 := frame.data[ofs+2]
	b3 := frame.data[ofs+3]

	var sample LIDAR_sample
	sample.no_data = (b0>>7)&1 != 0
	sample.too_close = (b0>>6)&1 != 0
	sample.dist = ((uint16(b0) & 0x3f) << 8) + uint16(b1)
	sample.ss = (uint16(b2) << 8) + uint16(b3)
	return &sample
}

//-----------------------------------------------------------------------------

func Open(name, port_name, pwm_name string) (*LIDAR, error) {
	var lidar LIDAR
	lidar.Name = name

	log.Printf("lidar.Open() %s serial=%s pwm=%s", lidar.Name, port_name, pwm_name)

	// open the serial port
	cfg := &serial.Config{Name: port_name, Baud: 115200, ReadTimeout: 20 * time.Millisecond}
	port, err := serial.OpenPort(cfg)
	if err != nil {
		log.Printf("unable to open serial port %s", port_name)
		return nil, err
	}
	lidar.port = port

	// open the pwm channel
	pwm, err := pwm.Open(fmt.Sprintf("%s_pwm", lidar.Name), pwm_name, 0.0)
	if err != nil {
		log.Printf("unable to open pwm channel %s", pwm_name)
		return nil, err
	}
	lidar.pwm = pwm

	// Initialise the PID
	pid, err := pid.Init(PID_PERIOD, PID_KP, PID_KI, PID_KD, PID_IMIN, PID_IMAX, PID_OMIN, PID_OMAX)
	if err != nil {
		log.Printf("unable to setup pid")
		return nil, err
	}
	pid.Set(LIDAR_RPM)
	lidar.pid = pid
	lidar.pid_on = true

	return &lidar, nil
}

//-----------------------------------------------------------------------------

func (lidar *LIDAR) Close() error {
	log.Printf("lidar.Close() %s ", lidar.Name)
	log.Printf("good/bad %d/%d", lidar.good_frames, lidar.bad_frames)

	lidar.pwm.Set(0.0)
	lidar.pwm.Close()

	err := lidar.port.Flush()
	if err != nil {
		log.Printf("error flushing serial port")
		return err
	}

	err = lidar.port.Close()
	if err != nil {
		log.Printf("error closing serial port")
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------

// Read the serial port and process the frames
func (lidar *LIDAR) read_serial() {
	// Note: We'd like to get all the bytes in one read. How many bytes is that?
	// n = period * (rpm / 60) * 90 frames/rev * 22 bytes/frame
	// n = 0.05 * (300/60) * 90 * 22
	// n = 495
	// TODO: multiple reads to flush startup junk
	buf := make([]byte, 512)
	n, err := lidar.port.Read(buf)
	if err != nil {
		log.Printf("error on serial read")
	}
	if n != 0 {
		lidar.rx_frame(buf[:n], time.Now())
	}
}

// Update the PWM value using the PID
func (lidar *LIDAR) motor_control() {
	// prevent motor burnout during pid tuning
	if lidar.rpm > LIDAR_RPM_SHUTDOWN {
		log.Printf("max motor rpm exceeded %f > %f", lidar.rpm, LIDAR_RPM_SHUTDOWN)
		lidar.pwm.Set(0.0)
		lidar.pid_on = false
	}
	if lidar.pid_on {
		lidar.pwm.Set(lidar.pid.Update(lidar.rpm))
	}
}

func (lidar *LIDAR) Process() {
	log.Printf("lidar.Process() %s", lidar.Name)

	read_tick := time.NewTicker(LIDAR_READ_PERIOD * time.Millisecond).C
	motor_tick := time.NewTicker(LIDAR_MOTOR_PERIOD * time.Millisecond).C

	for {
		select {
		case <-read_tick:
			lidar.read_serial()
		case <-motor_tick:
			lidar.motor_control()
		}
	}
}

//-----------------------------------------------------------------------------
