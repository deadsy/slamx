//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------

package view

import (
	"log"
	"math"
	"sync"

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

	view.renderer.SetDrawColor(255, 0, 0, 255)
	view.renderer.Clear()

	view.renderer.SetDrawColor(255, 255, 255, 255)

	view.plot_xy(0, 0)
	view.plot_xy(-100, 0)
	view.plot_xy(100, 0)
	view.plot_xy(0, -100)
	view.plot_xy(0, 100)

	for i := 0; i < 360; i++ {
		view.plot_polar(200, d2r(float32(i)))
	}

	view.renderer.Present()

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

func (view *View) Process(quit <-chan bool, wg *sync.WaitGroup) {
	log.Printf("%s.Process() enter", view.Name)
	defer wg.Done()

	for {
		select {
		case <-quit:
			log.Printf("%s.Process() exit", view.Name)
			return
		}
	}
}
