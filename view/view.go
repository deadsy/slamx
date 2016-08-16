package view

import (
	"errors"
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

const mm_per_pixel = 2.0
const pixel_per_mm = 1 / mm_per_pixel

func d2r(d float32) float32 {
	return math.Pi * (d / 180.0)
}

// world to screen coordinate conversion
// world = (0,0) is the center of the screen - ie the robot lidar center
// screen (0,0) is the top left corner of the display window
func world2screen(wx, wy float32) (sx, sy int) {
	sx = int((wx * pixel_per_mm) + (float32(WINDOW_X) / 2.0))
	sy = int((-wy * pixel_per_mm) + (float32(WINDOW_Y) / 2.0))
	return
}

func plot_xy(renderer *sdl.Renderer, x, y float32) {
	sx, sy := world2screen(x, y)
	renderer.DrawPoint(sx, sy)
}

func plot_polar(renderer *sdl.Renderer, r, theta float32) {
	x := float32(float64(r) * math.Cos(float64(theta)))
	y := float32(float64(r) * math.Sin(float64(theta)))
	plot_xy(renderer, x, y)
}

func Process() error {

	sdl.Init(sdl.INIT_EVERYTHING)

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WINDOW_X, WINDOW_Y, sdl.WINDOW_SHOWN)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create window: %s\n", err))
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create renderer: %s\n", err))
	}
	defer renderer.Destroy()

	renderer.Clear()

	renderer.SetDrawColor(255, 255, 255, 255)

	plot_xy(renderer, 0, 0)
	plot_xy(renderer, -100, 0)
	plot_xy(renderer, 100, 0)
	plot_xy(renderer, 0, -100)
	plot_xy(renderer, 0, 100)

	for i := 0; i < 360; i++ {
		plot_polar(renderer, 200, d2r(float32(i)))
	}

	renderer.Present()

	sdl.Delay(5000)
	sdl.Quit()
	return nil
}
