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
The motor spins counter-clockwise as viewed from above.

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
	"log"
	"sync"
	"time"

	"github.com/deadsy/slamx/motor"
	"github.com/deadsy/slamx/pid"
	"github.com/deadsy/slamx/util"
	"github.com/tarm/serial"
)

//-----------------------------------------------------------------------------

type LIDAR struct {
	Name       string       // user name for this device
	PortName   string       // serial port name
	Motor      *motor.Motor // motor driver
	Ctrl       chan Ctrl    // control channel
	Scan       chan Scan2D  // scan data channel
	RPM        float32      // measured rpm
	Running    bool         // is the PID turned on?
	GoodFrames uint         // good frames rx-ed
	BadFrames  uint         // bad frames rx-ed (invalid checksum)

	port     *serial.Port
	pid      *pid.PID
	rpm_lock sync.Mutex  // lock for access to rpm
	frame    LIDAR_frame // frame being read from serial
	ofs      int         // offset into frame data
	scan_idx int         // current scan index
	scan     Scan2D      // current scan data
}

//-----------------------------------------------------------------------------
// Motor Speed Control

const LIDAR_RPM = 300.0          // target rpm
const LIDAR_RPM_SHUTDOWN = 330.0 // shutdown limit
const LIDAR_MOTOR_PERIOD = 200   // update the motor pwm every N ms

// PID parameters
const PID_PERIOD = float32(LIDAR_MOTOR_PERIOD) / 1000.0
const PID_KP = 0.0
const PID_KI = 0.0
const PID_KD = 0.0
const PID_IMIN = -1.0
const PID_IMAX = 1.0
const PID_OMIN = 0.0
const PID_OMAX = 0.5

// The measured motor rpm is determined by read_serial() and used by the motor_control().
// These are different goroutines, so we use locking.

// set the motor rpm process value
func (l *LIDAR) set_rpm_pv(rpm float32) {
	l.rpm_lock.Lock()
	l.RPM = rpm
	l.rpm_lock.Unlock()
}

// get the motor rpm process value
func (l *LIDAR) get_rpm_pv() float32 {
	l.rpm_lock.Lock()
	rpm := l.RPM
	l.rpm_lock.Unlock()
	return rpm
}

// Update the PWM value using the PID
func (l *LIDAR) motor_control(quit <-chan bool, wg *sync.WaitGroup) {
	log.Printf("%s.motor_control() enter", l.Name)
	defer wg.Done()
	// perform pid/pwm updates at the LIDAR_MOTOR_PERIOD
	tick := time.NewTicker(LIDAR_MOTOR_PERIOD * time.Millisecond)
	for {
		select {
		case <-quit:
			log.Printf("%s.motor_control() exit", l.Name)
			tick.Stop()
			return
		case <-tick.C:
			rpm := l.get_rpm_pv()
			// prevent motor burnout during pid tuning
			if rpm > LIDAR_RPM_SHUTDOWN {
				log.Printf("%s: max motor rpm exceeded %f > %f", l.Name, rpm, LIDAR_RPM_SHUTDOWN)
				l.Motor.Set(0.0)
				l.Running = false
			}
			if l.Running {
				l.Motor.Set(l.pid.Update(rpm))
			}
		}
	}
}

//-----------------------------------------------------------------------------
// Samples and Scans

const SAMPLES_PER_SCAN = 360

func (l *LIDAR) alloc_scan() {
	l.scan_idx = 0
	l.scan = make([]Sample2D, SAMPLES_PER_SCAN)
}

func (scan Scan2D) add_sample(f *LIDAR_frame, base, idx int) {
	ofs := LIDAR_SAMPLE_OFS + (idx * LIDAR_SAMPLE_SIZE)
	b0 := f.data[ofs]
	b1 := f.data[ofs+1]
	b2 := f.data[ofs+2]
	b3 := f.data[ofs+3]

	idx += base
	s := &scan[idx]
	s.Good = (b1>>7)&1 == 0
	s.Too_Close = (b1>>6)&1 != 0
	s.Angle = util.DtoR(float32(idx))

	dist := ((int(b1) & 0x3f) << 8) + int(b0)
	ss := (int(b3) << 8) + int(b2)

	s.Distance = float32(dist) / 1000.0
	s.Signal_Strength = float32(ss)
}

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
func (l *LIDAR) process_frame() {
	f := &l.frame
	// set rpm for the PID process value
	l.set_rpm_pv(f.rpm())
	// add the frame samples to the current scan
	idx := f.angle()
	if idx < l.scan_idx {
		// report the scan
		log.Printf("%s: scan complete (%.1f rpm)", l.Name, f.rpm())
		l.Scan <- l.scan
		// alocate the next scan
		l.alloc_scan()
	}
	// add the 4 samples
	l.scan.add_sample(f, idx, 0)
	l.scan.add_sample(f, idx, 1)
	l.scan.add_sample(f, idx, 2)
	l.scan.add_sample(f, idx, 3)
	l.scan_idx = idx + 4
}

// receive a lidar frame from a buffer
func (l *LIDAR) rx_frame(buf []byte, ts time.Time) {
	// We look for a start of frame and a valid index to mark a frame.
	// We may get some false positives, but they will be weeded out with bad checksums.
	// Once we sync with the frame cadence we should be good.
	f := &l.frame
	for _, c := range buf {
		f.data[l.ofs] = c
		if l.ofs == LIDAR_START_OFS {
			// looking for start of frame
			if c == LIDAR_SOF_DELIMITER {
				// ok - set the timestamp
				f.ts = ts
				// now read the index
				l.ofs += 1
			}
		} else if l.ofs == LIDAR_INDEX_OFS {
			// looking for a valid index
			if c >= LIDAR_MIN_INDEX && c <= LIDAR_MAX_INDEX {
				// ok - now read the frame body
				l.ofs += 1
			} else {
				// not a frame - keep looking
				l.ofs = LIDAR_START_OFS
			}
		} else if l.ofs == LIDAR_END_OFS {
			// validate checksum
			calc_cs := f.checksum()
			frame_cs := f.get_uint16(LIDAR_CHECKSUM_OFS)
			if calc_cs == frame_cs {
				// good frame - process it
				l.GoodFrames += 1
				l.process_frame()
			} else {
				// bad frame
				l.BadFrames += 1
			}
			// reset for the next frame
			l.ofs = LIDAR_START_OFS
		} else {
			// reading the frame body
			l.ofs += 1
		}
	}
}

// Read the serial port and process the frames
func (l *LIDAR) read_serial(quit <-chan bool, wg *sync.WaitGroup) {
	log.Printf("%s.read_serial() enter", l.Name)
	defer wg.Done()
	for {
		select {
		case <-quit:
			log.Printf("%s.read_serial() exit", l.Name)
			return
		default:
			buf := make([]byte, 1024)
			n, err := l.port.Read(buf)
			if err == nil {
				l.rx_frame(buf[:n], time.Now())
				// Wait a while - there's a tradeoff here between data latency and cpu usage.
				// A smaller wait time gives lower latency and more cpu consumption.
				// 300 rpm = 200 ms/rev, so 50 ms is 1/4 revolution
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

//-----------------------------------------------------------------------------

func NewLIDAR(name, port_name string, motor *motor.Motor) (*LIDAR, error) {

	l := LIDAR{
		Name:     name,
		PortName: port_name,
		Motor:    motor,
	}
	log.Printf("NewLidar() %s", l.Name)

	// open the serial port
	cfg := &serial.Config{Name: l.PortName, Baud: 115200, ReadTimeout: 500 * time.Millisecond}
	port, err := serial.OpenPort(cfg)
	if err != nil {
		log.Printf("%s: unable to open serial port %s", l.Name, l.PortName)
		return nil, err
	}
	l.port = port

	// Initialise the PID
	pid, err := pid.Init(PID_PERIOD, PID_KP, PID_KI, PID_KD, PID_IMIN, PID_IMAX, PID_OMIN, PID_OMAX)
	if err != nil {
		log.Printf("%s: unable to setup pid", l.Name)
		return nil, err
	}
	pid.Set(0.0)
	l.pid = pid
	l.Running = false

	// allocate the initial scan
	l.alloc_scan()

	// setup lidar channels
	l.Ctrl = make(chan Ctrl)
	l.Scan = make(chan Scan2D)

	return &l, nil
}

func (l *LIDAR) Close() error {
	log.Printf("%s.Close()", l.Name)

	l.Motor.Set(0.0)
	err := l.port.Flush()

	if err != nil {
		log.Printf("%s: error flushing serial port", l.Name)
		return err
	}

	err = l.port.Close()
	if err != nil {
		log.Printf("%s: error closing serial port", l.Name)
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------

func (l *LIDAR) Process(quit <-chan bool, wg *sync.WaitGroup) {
	log.Printf("%s.Process() enter", l.Name)
	defer wg.Done()

	lidar_wg := &sync.WaitGroup{}

	// start motor control
	lidar_wg.Add(1)
	go l.motor_control(quit, lidar_wg)

	// start serial port reading
	lidar_wg.Add(1)
	go l.read_serial(quit, lidar_wg)

	for {
		select {
		case ctrl := <-l.Ctrl:
			switch ctrl {
			case Start:
				log.Printf("%s.Process() start", l.Name)
			case Stop:
				log.Printf("%s.Process() stop", l.Name)
			default:
				log.Printf("%s.Process() unknown ctrl %d", l.Name, ctrl)
			}
		case <-quit:
			lidar_wg.Wait()
			l.Close()
			log.Printf("%s.Process() exit", l.Name)
			return
		}
	}
}

//-----------------------------------------------------------------------------
