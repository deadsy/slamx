//-----------------------------------------------------------------------------
/*

PWM Control

*/
//-----------------------------------------------------------------------------

package pwm

import (
	"log"
)

//-----------------------------------------------------------------------------

type PWM struct {
	Name string
	Val  float32
}

//-----------------------------------------------------------------------------

// Open the PWM channel
func Open(name, pwm_name string, val float32) (*PWM, error) {
	var pwm PWM
	pwm.Name = name

	log.Printf("pwm.Open() %s (%s)\n", pwm.Name, pwm_name)
	pwm.Set(val)

	return &pwm, nil
}

//-----------------------------------------------------------------------------

// Close the PWM channel
func (pwm *PWM) Close() error {
	log.Printf("pwm.Close() %s\n", pwm.Name)
	return nil
}

//-----------------------------------------------------------------------------

// Set the PWM value
func (pwm *PWM) Set(val float32) {
	log.Printf("pwm.Set() %s = %f\n", pwm.Name, pwm.Val)

	if val > 1.0 {
		val = 1.0
	}
	if val < 0.0 {
		val = 0.0
	}

	pwm.Val = val
	// TODO do it
}

//-----------------------------------------------------------------------------
