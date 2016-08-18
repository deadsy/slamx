//-----------------------------------------------------------------------------
/*

A Graphical View of System State

Notes:

1) This code uses the SDL2 library to do graphics.
2) SDL2 is not multithreading friendly - it uses thread local storage.
   To avoid problems We make all calls to SDL2 from the main thread.

*/
//-----------------------------------------------------------------------------

package view

import (
	"log"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

//-----------------------------------------------------------------------------

type View struct {
	Name     string
	window   *sdl.Window
	renderer *sdl.Renderer
}

//-----------------------------------------------------------------------------

const mm_per_pixel = 2.0
const pixel_per_mm = 1 / mm_per_pixel

func d2r(d float32) float32 {
	return math.Pi * (d / 180.0)
}

//-----------------------------------------------------------------------------

// world to screen coordinate conversion
// world = (0,0) is the center of the screen - ie the robot lidar center
// screen (0,0) is the top left corner of the display window
func world2screen(wx, wy float32) (sx, sy int) {
	sx = int((wx * pixel_per_mm) + (float32(WINDOW_X) / 2.0))
	sy = int((-wy * pixel_per_mm) + (float32(WINDOW_Y) / 2.0))
	return
}

// plot an (x.y) point given in world ccordinates
func (view *View) plot_xy(x, y float32) {
	sx, sy := world2screen(x, y)
	view.renderer.DrawPoint(sx, sy)
}

// plot an (r,theta) polar point given in world coordinates
func (view *View) plot_polar(r, theta float32) {
	x := float32(float64(r) * math.Cos(float64(theta)))
	y := float32(float64(r) * math.Sin(float64(theta)))
	view.plot_xy(x, y)
}

//-----------------------------------------------------------------------------

func Open(name string) (*View, error) {
	var view View
	view.Name = name

	log.Printf("%s.Open()", view.Name)

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Printf("%s: sdl.Init() failed %s", err)
		return nil, err
	}

	// create the window
	window, err := sdl.CreateWindow("slamx", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WINDOW_X, WINDOW_Y, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Printf("%s: sdl.CreateWindow() failed %s", err)
		return nil, err
	}
	view.window = window

	// create the renderer
	renderer, err := sdl.CreateRenderer(view.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Printf("%s: sdl.CreateRenderer() failed %s", err)
		return nil, err
	}
	view.renderer = renderer

	// setup the renderer
	view.renderer.SetLogicalSize(WINDOW_X, WINDOW_Y)

	return &view, nil
}

//-----------------------------------------------------------------------------

func (view *View) Close() {
	log.Printf("%s.Close()", view.Name)
	sdl.Quit()
	view.renderer.Destroy()
	view.window.Destroy()
}

//-----------------------------------------------------------------------------

func (view *View) Delay(ms uint32) {
	sdl.Delay(ms)
}

//-----------------------------------------------------------------------------

const steps = 100

func (view *View) line(tofs, t0, t1, x float32) {
	dt := (t1 - t0) / steps
	t := t0
	for i := 0; i < steps; i++ {
		d := float32(float64(x) / math.Cos(float64(t)))
		view.plot_polar(d, t+tofs)
		t += dt
	}
}

func (view *View) Render(ofs float32) {
	// clear the background
	view.renderer.SetDrawColor(0, 0, 0, 255)
	view.renderer.Clear()
	// draw a rotated square
	view.renderer.SetDrawColor(255, 255, 255, 255)
	view.line(d2r(0+ofs), d2r(-45), d2r(45), 200)
	view.line(d2r(90+ofs), d2r(-45), d2r(45), 200)
	view.line(d2r(180+ofs), d2r(-45), d2r(45), 200)
	view.line(d2r(270+ofs), d2r(-45), d2r(45), 200)
	// render to the window
	view.renderer.Present()
}

//-----------------------------------------------------------------------------

func (view *View) Events() bool {
	rc := true
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
      log.Printf("sdl event %+v", t)
			rc = false
		case *sdl.MouseMotionEvent:
		case *sdl.MouseButtonEvent:
		case *sdl.MouseWheelEvent:
		case *sdl.KeyDownEvent:
		case *sdl.KeyUpEvent:
		case *sdl.JoyAxisEvent:
		case *sdl.JoyBallEvent:
		case *sdl.JoyButtonEvent:
		case *sdl.JoyHatEvent:
		case *sdl.JoyDeviceEvent:
		default:
			// log.Printf("event %+v", t)
		}
	}
	return rc
}

//-----------------------------------------------------------------------------
