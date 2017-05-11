// +build pc

//-----------------------------------------------------------------------------
/*

PC Development System GPIO Control

*/
//-----------------------------------------------------------------------------

package gpio

import (
	"fmt"
	"log"
)

// clamp a value from 0.0 to 1.0
func clamp(x float32) float32 {
	if x > 1.0 {
		return 1.0
	}
	if x < 0.0 {
		return 0.0
	}
	return x
}

//-----------------------------------------------------------------------------

type GPIO struct {
	Name string
}

func NewGPIO(name string) (*GPIO, error) {
	g := GPIO{
		Name: name,
	}
	log.Printf("NewGPIO() %s", g.Name)
	return &g, nil
}

func (g *GPIO) Close() {
	log.Printf("%s.Close() %s", g.Name)
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
}

// Clear the output pin (0)
func (p *Output) Clr() {
	log.Printf("%s.Clr()", p.Name)
}

// Close the output.
func (p *Output) Close() {
	log.Printf("%s.Close()", p.Name)
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
	p.val = clamp(val)
}

// Close the PWM channel
func (p *PWM) Close() {
	log.Printf("%s.Close()", p.Name)
}

//-----------------------------------------------------------------------------
