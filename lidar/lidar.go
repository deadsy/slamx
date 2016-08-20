//-----------------------------------------------------------------------------
/*

Generic LIDAR Code

*/
//-----------------------------------------------------------------------------

package lidar

//-----------------------------------------------------------------------------

// 2D LIDAR Sample
type Sample_2D struct {
	no_data   bool    // no return, max range, too low of reflectivity
	too_close bool    // object too close
	angle     float32 // angle in radians
	dist      float32 // distance in meters
	ss        float32 // signal strength
}

// 2D LIDAR Scan
type Scan_2D struct {
	samples []Sample_2D
}

//-----------------------------------------------------------------------------
