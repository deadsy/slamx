//-----------------------------------------------------------------------------
/*

Generic LIDAR Code

*/
//-----------------------------------------------------------------------------

package lidar

//-----------------------------------------------------------------------------

// 2D LIDAR Sample
type Sample2D struct {
	Good            bool    // good data in this sample
	Too_Close       bool    // object too close
	Angle           float32 // angle in radians
	Distance        float32 // distance in meters
	Signal_Strength float32 // signal strength
}

// 2D LIDAR Scan
type Scan2D []Sample2D

// Control Values
type Ctrl int

const (
	Stop  Ctrl = iota // stop scanning
	Start             // start scanning
)

//-----------------------------------------------------------------------------
