// Package pid provides a generic PID Controller.
//
// References:
// https://en.wikipedia.org/wiki/PID_controller
package pid

import (
	"errors"
	"log"
)

//-----------------------------------------------------------------------------

// PID controller state.
type PID struct {
	kp     float32 // proportional constant
	ki     float32 // integral constant
	kd     float32 // derivative constant
	iMax   float32 // max limit on iTerm
	iMin   float32 // min limit on iTerm
	oMax   float32 // max limit on output
	oMin   float32 // min limit on output
	sp     float32 // set point value (target)
	evPrev float32 // previous error value
	iTerm  float32 // integral sum
	dFlag  bool    // avoid spiking the derivative term on the first update
}

//-----------------------------------------------------------------------------

// Update the PID, return the control value.
func (p *PID) Update(pv float32) float32 {

	ev := p.sp - pv

	// proportional
	pTerm := p.kp * ev

	// integral
	if p.ki != 0.0 {
		p.iTerm += p.ki * ev
		// limit the integration sum
		if p.iTerm > p.iMax {
			p.iTerm = p.iMax
		}
		if p.iTerm < p.iMin {
			p.iTerm = p.iMin
		}
	}

	// derivative
	if !p.dFlag {
		// avoid spiking the dTerm on the first update
		p.evPrev = ev
		p.dFlag = true
	}
	dTerm := p.kd * (ev - p.evPrev)
	p.evPrev = ev

	// calculate and limit the output
	out := pTerm + p.iTerm + dTerm
	if out > p.oMax {
		out = p.oMax
		log.Printf("limiting max pid output %f", out)
	}
	if out < p.oMin {
		out = p.oMin
		log.Printf("limiting min pid output %f", out)
	}

	return out
}

//-----------------------------------------------------------------------------

// Set the PID setpoint value.
func (p *PID) Set(sp float32) {
	p.sp = sp
}

// Reset the PID controller state.
func (p *PID) Reset() {
	p.sp = 0
	p.evPrev = 0
	p.iTerm = 0
	p.dFlag = false
}

//-----------------------------------------------------------------------------

// Init initialises the PID controller state.
func Init(dt, kp, ki, kd, iMin, iMax, oMin, oMax float32) (*PID, error) {

	var p PID

	if dt < 0.0 || kp < 0.0 || ki < 0.0 || kd < 0.0 || iMin > iMax || oMin > oMax {
		return nil, errors.New("invalid PID parameters")
	}

	p.kp = kp
	p.ki = ki * dt
	p.kd = kd / dt

	p.iMin = iMin
	p.iMax = iMax

	p.oMin = oMin
	p.oMax = oMax

	p.sp = 0
	p.evPrev = 0
	p.iTerm = 0
	p.dFlag = false

	return &p, nil
}

//-----------------------------------------------------------------------------
