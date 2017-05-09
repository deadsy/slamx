// +build rpi

//-----------------------------------------------------------------------------
/*

Raspberry Pi PWM Control

Using the pi-blaster service.
See- https://github.com/sarfata/pi-blaster

*/
//-----------------------------------------------------------------------------

package pwm

import (
	"fmt"
	"log"
	"math"
	"os"
)

type PWM struct {
	Name string
	pin  string
	val  float32
	dev  *os.File
}

//-----------------------------------------------------------------------------

const PWM_RESOLUTION = 1000.0 // NUM_SAMPLES per pi-blaster code

// normalise a pwm value
func normalise(val float32) float32 {
	// clamp the value between 0 and 1
	if val > 1.0 {
		return 1.0
	}
	if val < 0.0 {
		return 0.0
	}
	// I don't want to incur the expense of IO if the pwm value
	// is dancing around at a level below the provided resolution,
	// so remove any superfluous resolution.
	val = float32(math.Floor((float64(val)*PWM_RESOLUTION)+0.5) / PWM_RESOLUTION)
	return val
}

//-----------------------------------------------------------------------------

// Write to the PWM device
func (p *PWM) write(msg string) error {
	_, err := p.dev.WriteString(msg)
	if err != nil {
		log.Printf("%s: can't write to pwm device", p.Name)
		return err
	}
	return nil
}

//-----------------------------------------------------------------------------

// Open the PWM channel
func Open(name, pin string, val float32) (*PWM, error) {
	var p PWM
	p.Name = name
	log.Printf("%s.Open() pin=%s", p.Name, pin)
	p.pin = pin
	f, err := os.OpenFile("/dev/pi-blaster", os.O_WRONLY, 0660)
	if err != nil {
		log.Printf("%s: can't open pwm device", p.Name)
		return nil, err
	}
	p.dev = f
	p.Set(val)
	return &p, nil
}

// Close the PWM channel
func (p *PWM) Close() {
	log.Printf("%s.Close()", p.Name)
	p.write(fmt.Sprintf("release %s", p.pin))
	p.dev.Close()
}

// Set the PWM value
func (p *PWM) Set(val float32) {
	//log.Printf("%s.Set() = %f", p.Name, val)
	val = normalise(val)
	if val == p.val {
		// no change
		return
	}
	p.val = val
	p.write(fmt.Sprintf("%s=%.3f", p.pin, p.val))
}

//-----------------------------------------------------------------------------
