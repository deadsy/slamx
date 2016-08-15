package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
)

const WINDOW_W = 640
const WINDOW_H = 400

func main() {
	sdl.Init(sdl.INIT_EVERYTHING)

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WINDOW_W, WINDOW_H, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	fmt.Printf("surface is %dx%d\n", surface.W, surface.H)

	rect := sdl.Rect{0, 0, 200, 200}
	surface.FillRect(&rect, sdl.MapRGB(surface.Format, 255, 0, 0))
	window.UpdateSurface()

	sdl.Delay(5000)
	sdl.Quit()
}
