//-----------------------------------------------------------------------------
/*

Driver for Neato XV11 LIDAR Unit

* Control the Motor using PWM to give constant RPM
* Read the serial stream from the LIDAR and repackage it as range data
* Feed the range data as messages to another Go routine

22 bytes/frame
90 frames/rev
300 revs/min
22 * 90 * 300 = 594000 bytes/min = 9900 bytes/sec
9900 * 11 bits/byte = 108900 bits/sec
So: The serial port runs at 115200 baud to keep up.

References:

https://xv11hacking.wikispaces.com/LIDAR+Sensor

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
	Name        string
	port        *serial.Port
	pwm         *pwm.PWM
	frame       LIDAR_frame // frame being read from serial
	ofs         int         // offset into frame data
	good_frames uint        // good frames rx-ed
	bad_frames  uint        // bad frames rx-ed (invalid checksum)
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
	log.Printf("ts %s", f.ts)
	log.Printf("rpm %f", f.rpm())
	log.Printf("theta %d", f.angle())
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
				log.Printf("bad checksum calc %04x frame %04x", calc_cs, frame_cs)
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

func (lidar *LIDAR) test_frame() {

	b0 := []byte{0xfa, 0xc8, 0xd9, 0x32, 0x22, 0x1a, 0x3f, 0x00, 0x22, 0x1a, 0x3f, 0x00, 0x22, 0x1a, 0x3f, 0x00, 0x22, 0x1a, 0x3f, 0x00, 0xb2, 0x7c}
	b1 := []byte{0xfa, 0xbf, 0xf9, 0x49, 0x50, 0x1a, 0x3f, 0x00, 0x50, 0x1a, 0x3f, 0x00, 0x50, 0x1a, 0x3f, 0x00, 0x50, 0x1a, 0x3f, 0x00, 0x49, 0x3b}
	b2 := []byte{0xFA, 0xF9, 0x16, 0x4A, 0x35, 0x1A, 0x00, 0x00, 0x90, 0x02, 0x17, 0x00, 0xE5, 0x02, 0xAC, 0x01, 0xE4, 0x02, 0x1A, 0x01, 0x16, 0x22}
	b3 := []byte{0xfa, 0xde, 0x5e, 0x4a, 0xd8, 0x07, 0x1f, 0x00, 0xe5, 0x07, 0x1b, 0x00, 0xf0, 0x0b, 0x09, 0x00, 0xfb, 0x0b, 0x08, 0x00, 0xcd, 0x3f}

	junk := []byte{0xde, 0xad, 0xbe, 0xef}

	lidar.rx_frame(b0, time.Now())
	lidar.rx_frame(junk, time.Now())
	lidar.rx_frame(b1, time.Now())
	lidar.rx_frame(junk, time.Now())
	lidar.rx_frame(b2, time.Now())
	lidar.rx_frame(junk, time.Now())
	lidar.rx_frame(b3, time.Now())
	lidar.rx_frame(junk, time.Now())
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
	sample.dist = ((uint16(b0) & 0x3F) << 8) + uint16(b1)
	sample.ss = (uint16(b2) << 8) + uint16(b3)
	return &sample
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

	lidar.test_frame()

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
