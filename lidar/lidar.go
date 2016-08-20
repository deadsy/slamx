/* generic lidar code */

package lidar

import (
	"time"
)

// 2D LIDAR Sample
type Sample_2D struct {
	angle float32 // angle in radians
	dist  float32 // distance in meters
}

// 2D LIDAR Scan
type Scan_2D struct {
	ts      time.Time // timestamp
	samples []Sample_2D
}
