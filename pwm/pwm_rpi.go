// +build rpi

//-----------------------------------------------------------------------------
/*

Raspberry Pi PWM Control

Using the pi-blaster service.

*/
//-----------------------------------------------------------------------------

package pwm

import (
	"log"
	"os"
)

// write to the pwm device
func (pwm *PWM) write(msg string) error {
	f, err := os.OpenFile("/dev/pi-blaster", os.O_WRONLY, 0660)
	if err != nil {
		log.Printf("can't open pwm device")
		return err
	}
	defer f.Close()
	_, err = f.WriteString(msg)
	if err != nil {
		log.Printf("can't write to pwm device")
		return err
	}
	return nil
}

// Close the PWM channel
func (pwm *PWM) Close() {
	log.Printf("pwm.Close() %s\n", pwm.Name)
	pwm.Set(0.0)
	err := pwm.write(fmt.Sprintf("release %s", pwm.Pin))
	if err != nil {
		log.Printf("error releasing pwm device")
	}
}

// Set the PWM value
func (pwm *PWM) Set(val float32) {
	log.Printf("pwm.Set() %s = %f\n", pwm.Name, pwm.Val)
	pwm.Val = clamp(val)
	err := pwm.write(fmt.Sprintf("%s=%.3f", pwm.Pin, pwm.Val))
	if err != nil {
		log.Printf("error writing to pwm device")
	}
}

//-----------------------------------------------------------------------------
