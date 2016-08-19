//-----------------------------------------------------------------------------
/*

SLAM

A port to Go of BreezySLAM
See: https://github.com/simondlevy/BreezySLAM

*/
//-----------------------------------------------------------------------------

package slam

//-----------------------------------------------------------------------------

const NO_OBSTACLE = 65500
const OBSTACLE = 0

//-----------------------------------------------------------------------------

type Position struct {
	x_mm          float64
	y_mm          float64
	theta_degrees float64
}

type Pixel uint16

type Map struct {
	Name          string
	pixels        []Pixel
	size_pixels   int
	size_meters   float64
	pixels_per_mm float64
}

type Scan struct {
	rate_hz float64 // scans per second
	size    int     // number of rays per scan

	//double * x_mm;
	//double * y_mm;
	//int * value;
	//int npoints;
	//int span;

	//double detection_angle_degrees;     /* e.g. 240, 360 */
	//double distance_no_detection_mm;    /* default value when the laser returns 0 */
	//int detection_margin;               /* first scan element to consider */
	//double offset_mm;                   /* position of the laser wrt center of rotation */

}

//-----------------------------------------------------------------------------

// out of bounds, true if val < 0 or val >= bound
func oob(val, bound int) bool {
	return (val < 0) || (val >= bound)
}

//-----------------------------------------------------------------------------

func Map_Init(name string, size_pixels int, size_meters float64) *Map {
	var m Map
	m.Name = name
	m.size_pixels = size_pixels
	m.size_meters = size_meters
	m.pixels_per_mm = float64(size_pixels) / (size_meters * 1000.0)
	m.pixels = make([]pixel, size_pixels*size_pixels)
	for i, _ := range m.pixels {
		m.pixels[i] = (OBSTACLE + NO_OBSTACLE) / 2
	}
	return &m
}

//-----------------------------------------------------------------------------
/* Scan Operations

The LIDAR provides us with a set of (angle, distance) values.
We want to convert this to (x,y) coordinates in the 2D world model.
To do this we need the LIDAR position (x,y,theta).
The LIDAR motion may be significant during the scan.
We'll improve accuracy by adding in the LIDAR velocity (dx,dy,dtheta).
The time between scan samples can be derived from the rotational speed of the LIDAR

The samples from the LIDAR maybe considered too coarse. They can be upsampled by
some integer factor to give scan values with a finer resolution. I don't plan to
do any interpolation between upsamples. That may be the wrong thing to do,
although any interpolation is likely to lead to strangeness at sharp object
boundaries (distance discontinuities).
*/

// Update the scan with some new LIDAR samples
func (scan *Scan) Update(samples []lidar.Sample) {
}

//-----------------------------------------------------------------------------
