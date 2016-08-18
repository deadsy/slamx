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
func (pwm *PWM) write(msg string) error {
	_, err := pwm.dev.WriteString(msg)
	if err != nil {
		log.Printf("%s: can't write to pwm device", pwm.Name)
		return err
	}
	return nil
}

//-----------------------------------------------------------------------------

// Open the PWM channel
func Open(name, pin string, val float32) (*PWM, error) {
	var pwm PWM
	pwm.Name = name
	log.Printf("%s.Open() pin=%s", pwm.Name, pin)
	pwm.pin = pin
	f, err := os.OpenFile("/dev/pi-blaster", os.O_WRONLY, 0660)
	if err != nil {
		log.Printf("%s: can't open pwm device", pwm.Name)
		return nil, err
	}
	pwm.dev = f
	pwm.Set(val)
	return &pwm, nil
}

// Close the PWM channel
func (pwm *PWM) Close() {
	log.Printf("%s.Close()", pwm.Name)
	pwm.write(fmt.Sprintf("release %s", pwm.pin))
	pwm.dev.Close()
}

// Set the PWM value
func (pwm *PWM) Set(val float32) {
	//log.Printf("%s.Set() = %f", pwm.Name, val)
	val = normalise(val)
	if val == pwm.val {
		// no change
		return
	}
	pwm.val = val
	pwm.write(fmt.Sprintf("%s=%.3f", pwm.pin, pwm.val))
}

//-----------------------------------------------------------------------------
