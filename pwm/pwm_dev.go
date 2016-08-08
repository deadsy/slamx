// +build dev

//-----------------------------------------------------------------------------
/*

PC Development System PWM Control

*/
//-----------------------------------------------------------------------------

package pwm

import (
	"log"
)

// Close the PWM channel
func (pwm *PWM) Close() {
	log.Printf("pwm.Close() %s\n", pwm.Name)
	pwm.Set(0.0)
	// TODO ...
}

// Set the PWM value
func (pwm *PWM) Set(val float32) {
	log.Printf("pwm.Set() %s = %f\n", pwm.Name, pwm.Val)
	pwm.Val = clamp(val)
	// TODO ....
}

//-----------------------------------------------------------------------------
