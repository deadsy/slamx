//-----------------------------------------------------------------------------
/*

Generic LIDAR Code

*/
//-----------------------------------------------------------------------------

package lidar

//-----------------------------------------------------------------------------

// 2D LIDAR Sample
type Sample_2D struct {
	Good            bool    // good data in this sample
	Too_Close       bool    // object too close
	Angle           float32 // angle in radians
	Distance        float32 // distance in meters
	Signal_Strength float32 // signal strength
}

// 2D LIDAR Scan
type Scan_2D struct {
	Samples []Sample_2D
}

//-----------------------------------------------------------------------------
