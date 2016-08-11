package pid

type PID struct {
	kp float32
	ki float32
	kd float32
}

// Update the PID, return the control value
func Update(e float32) float32 {

	return 0.0
}

func Init(kp, ki, kd float32) *PID {

	var pid PID

	pid.kp = kp
	pid.ki = ki
	pid.kd = kd

	return &pid
}
