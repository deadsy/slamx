package pid

import (
	"errors"
	"log"
)

type PID struct {
	kp      float32 // proportional constant
	ki      float32 // integral constant
	kd      float32 // derivative constant
	i_max   float32 // max limit on i_term
	i_min   float32 // min limit on i_term
	o_max   float32 // max limit on output
	o_min   float32 // min limit on output
	sp      float32 // set point value (target)
	ev_prev float32 // previous error value
	i_term  float32 // integral sum
	d_flag  bool    // avoid spiking d_term on the first update
}

// Update the PID, return the control value
func (pid *PID) Update(pv float32) float32 {

	ev := pid.sp - pv

	// proportional
	p_term := pid.kp * ev

	// integral
	if pid.ki != 0.0 {
		pid.i_term += pid.ki * ev
		// limit the integration sum
		if pid.i_term > pid.i_max {
			pid.i_term = pid.i_max
		}
		if pid.i_term < pid.i_min {
			pid.i_term = pid.i_min
		}
	}

	// derivative
	if !pid.d_flag {
		// avoid spiking the d_term on the first update
		pid.ev_prev = ev
		pid.d_flag = true
	}
	d_term := pid.kd * (ev - pid.ev_prev)
	pid.ev_prev = ev

	// calculate and limit the output
	out := p_term + pid.i_term + d_term
	if out > pid.o_max {
		out = pid.o_max
		log.Printf("limiting max pid output %f", out)
	}
	if out < pid.o_min {
		out = pid.o_min
		log.Printf("limiting min pid output %f", out)
	}

	return out
}

// Set the process setpoint
func (pid *PID) Set(sp float32) {
	pid.sp = sp
}

func Init(dt, kp, ki, kd, i_min, i_max, o_min, o_max float32) (*PID, error) {

	var pid PID

	if dt < 0.0 || kp < 0.0 || ki < 0.0 || kd < 0.0 || i_min > i_max || o_min > o_max {
		return nil, errors.New("invalid PID parameters")
	}

	pid.kp = kp
	pid.ki = ki * dt
	pid.kd = kd / dt

	pid.i_min = i_min
	pid.i_max = i_max

	pid.o_min = o_min
	pid.o_max = o_max

	pid.sp = 0.0
	pid.ev_prev = 0.0
	pid.i_term = 0.0
	pid.d_flag = false

	return &pid, nil
}
