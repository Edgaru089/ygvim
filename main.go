package main

import (
	"log"
	"runtime"
	"sync/atomic"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"

	"edgaru089.ink/go/ygvim/internal/ui"
	"edgaru089.ink/go/ygvim/internal/util/itype"
)

func init() {
	runtime.LockOSThread()
}

func main() {

	cellsize := itype.Vec2i{8, 16}
	width, height := 100, 32

	sdl.LoadLibrary(sdl.Path())
	defer sdl.Quit()
	ttf.LoadLibrary(ttf.Path())
	defer ttf.Quit()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	if err := ttf.Init(); err != nil {
		panic(err)
	}

	renderDrivers := []string{}
	for i := 0; i < sdl.GetNumRenderDrivers(); i++ {
		renderDrivers = append(renderDrivers, sdl.GetRenderDriver(i))
	}
	log.Printf("available SDL render drivers: %v", renderDrivers)

	window, err := sdl.CreateWindow(
		"gvim",
		cellsize[0]*width, cellsize[1]*height,
		sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}

	ren, err := window.CreateRenderer("vulkan")
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	defer ren.Destroy()

	ren.SetVSync(1)

	nvimui := ui.InvokeNvim(window, ren, width, height, cellsize)

	window.StartTextInput()
	defer window.StopTextInput()

	// main loop
	var running atomic.Bool
	running.Store(true)

	go func() {
		err := nvimui.Serve()
		if err != nil {
			log.Printf("main: nvim.Serve: %e", err)
		}
		running.Store(false)
	}()

	go func() {
		for running.Load() {
			nvimui.WaitRender(ren)
		}
	}()

	nvimui.AttachUI()

	for running.Load() {
		var event sdl.Event

		err := sdl.WaitEvent(&event)
		if err == nil {
			switch event.Type {
			case sdl.EVENT_QUIT:
				running.Store(false)

			case sdl.EVENT_TEXT_INPUT:
				nvimui.InputText(event.TextInputEvent().Text)
			case sdl.EVENT_KEY_DOWN:
				nvimui.InputKey(event.KeyboardEvent().Key, event.KeyboardEvent().Mod)

			case sdl.EVENT_WINDOW_EXPOSED:
				nvimui.SetNeedFlip()
			case sdl.EVENT_WINDOW_RESIZED, sdl.EVENT_WINDOW_MAXIMIZED, sdl.EVENT_WINDOW_RESTORED:
				//w, h := event.WindowEvent().Data1, event.WindowEvent().Data2
				//window.SetSize(w, h)
				w, h, err := window.SizeInPixels()
				if err == nil {
					nvimui.SetWindowSize(int(w), int(h))
				}
			}
		} else {
			log.Printf("main: WaitEvent: %e", err)
		}
	}

}
