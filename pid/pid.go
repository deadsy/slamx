package pid

import (
	"log"
)

type PID struct {
	kp      float32 // prportional constant
	ki      float32 // integral constant
	kd      float32 // derivative constant
	i_max   float32 // max limit on i_term
	i_min   float32 // min limit on i_term
	o_max   float32 // max limit on output
	o_min   float32 // min limit on output
	sp      float32 // set point value (target)
	pv_prev float32 // previous process value
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
		pid.pv_prev = pv
		pid.d_flag = true
	}
	d_term := pid.kd * (pv - pid.pv_prev)
	pid.pv_prev = pv

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

func Init(kp, ki, kd, i_min, i_max, o_min, o_max float32) *PID {

	var pid PID

	pid.kp = kp
	pid.ki = ki
	pid.kd = kd

	pid.i_min = i_min
	pid.i_max = i_max

	pid.o_min = o_min
	pid.o_max = o_max

	pid.sp = 0.0
	pid.pv_prev = 0.0
	pid.i_term = 0.0
	pid.d_flag = false

	return &pid
}
