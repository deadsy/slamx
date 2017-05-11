//-----------------------------------------------------------------------------
/*
Motor Driver

SparkFun ROB-09457 TB6612FNG Motor Driver
https://www.sparkfun.com/products/9457

STBY turns the board on/off
PWM controls the speed.
The board is wired for CCW operation.
*/
//-----------------------------------------------------------------------------

package motor

import (
	"log"

	"github.com/deadsy/slamx/gpio"
)

//-----------------------------------------------------------------------------

type Motor struct {
	Name string
	pwm  *gpio.PWM
	stby *gpio.Output
}

func NewMotor(name string, pwm *gpio.PWM, stby *gpio.Output) (*Motor, error) {
	m := Motor{
		Name: name,
		pwm:  pwm,
		stby: stby,
	}
	log.Printf("NewMotor() %s", m.Name)
	m.stby.Set()
	m.pwm.Set(0)
	return &m, nil
}

func (m *Motor) Close() {
	log.Printf("%s.Close()", m.Name)
	m.stby.Clr()
	m.pwm.Set(0)
}

func (m *Motor) Set(val float32) {
	m.pwm.Set(val)
}

//-----------------------------------------------------------------------------
