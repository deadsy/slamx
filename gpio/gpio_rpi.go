// +build rpi

//-----------------------------------------------------------------------------
/*

Raspberry Pi GPIO Control

Using the pi-blaster service.
See- https://github.com/sarfata/pi-blaster

*/
//-----------------------------------------------------------------------------

package gpio

import (
	"fmt"
	"log"
	"math"
	"os"
)

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

type GPIO struct {
	Name   string
	device *os.File
}

// Create a new GPIO device.
func NewGPIO(name string) (*GPIO, error) {
	g := GPIO{
		Name: name,
	}
	log.Printf("NewGPIO() %s", g.Name)
	f, err := os.OpenFile("/dev/pi-blaster", os.O_WRONLY, 0660)
	if err != nil {
		log.Printf("%s: can't open gpio device", g.Name)
		return nil, err
	}
	g.device = f
	return &g, nil
}

// Close the GPIO device.
func (g *GPIO) Close() {
	log.Printf("%s.Close()", g.Name)
	g.device.Close()
}

// Write to the GPIO device.
func (g *GPIO) write(msg string) error {
	_, err := g.device.WriteString(msg)
	if err != nil {
		log.Printf("%s: can't write to gpio device", g.Name)
		return err
	}
	return nil
}

//-----------------------------------------------------------------------------

type Output struct {
	Name string
	gpio *GPIO
	pin  string
}

// Create a new GPIO output.
func (g *GPIO) NewOutput(pin string, val int) (*Output, error) {
	p := Output{
		Name: fmt.Sprintf("%s_out_%s", g.Name, pin),
		gpio: g,
		pin:  pin,
	}
	log.Printf("NewOutput() %s", p.Name)
	if val != 0 {
		p.Set()
	} else {
		p.Clr()
	}
	return &p, nil
}

// Set the output pin (1)
func (p *Output) Set() {
	log.Printf("%s.Set()", p.Name)
	p.gpio.write(fmt.Sprintf("%s=1\n", p.pin))
}

// Clear the output pin (0)
func (p *Output) Clr() {
	log.Printf("%s.Clr()", p.Name)
	p.gpio.write(fmt.Sprintf("%s=0\n", p.pin))
}

// Close the output.
func (p *Output) Close() {
	log.Printf("%s.Close()", p.Name)
	p.gpio.write(fmt.Sprintf("release %s\n", p.pin))
}

//-----------------------------------------------------------------------------

type PWM struct {
	Name string
	gpio *GPIO
	pin  string
	val  float32
}

// Create a new PWM device.
func (g *GPIO) NewPWM(pin string, val float32) (*PWM, error) {
	p := PWM{
		Name: fmt.Sprintf("%s_pwm_%s", g.Name, pin),
		gpio: g,
		pin:  pin,
	}
	log.Printf("NewPWM() %s", p.Name)
	p.Set(val)
	return &p, nil
}

// Set the PWM value
func (p *PWM) Set(val float32) {
	log.Printf("%s.Set() %f\n", p.Name, val)
	val = normalise(val)
	if val != 0.0 && val == p.val {
		// no change
		return
	}
	p.val = val
	p.gpio.write(fmt.Sprintf("%s=%.3f\n", p.pin, p.val))
}

// Close the PWM channel
func (p *PWM) Close() {
	log.Printf("%s.Close()", p.Name)
	p.gpio.write(fmt.Sprintf("release %s\n", p.pin))
}

//-----------------------------------------------------------------------------
