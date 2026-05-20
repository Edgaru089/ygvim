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

func main() {

	cellsize := itype.Vec2i{8, 16}
	width, height := 100, 32

	loadlibrary()
	defer unloadlibrary()

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

	var rendriver string
	if runtime.GOOS == "windows" {
		rendriver = "direct3d11"
	} else {
		rendriver = "vulkan"
	}
	ren, err := window.CreateRenderer(rendriver)
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
					nw, nh := nvimui.SetWindowSize(int(w), int(h))
					window.SetSize(int32(nw), int32(nh))
				}

			case sdl.EVENT_MOUSE_BUTTON_DOWN:
				nvimui.MousePress(sdl.MouseButtonFlags(event.MouseButtonEvent().Button))
			case sdl.EVENT_MOUSE_BUTTON_UP:
				nvimui.MouseRelease(sdl.MouseButtonFlags(event.MouseButtonEvent().Button))
			case sdl.EVENT_MOUSE_MOTION:
				e := event.MouseMotionEvent()
				nvimui.MouseMove(int(e.X), int(e.Y))

			case sdl.EVENT_MOUSE_WHEEL:
				e := event.MouseWheelEvent()
				x, y := int(e.IntegerX), int(e.IntegerY)
				switch e.Direction {
				case sdl.MOUSEWHEEL_NORMAL:
				case sdl.MOUSEWHEEL_FLIPPED:
					x, y = -x, -y
				}

				// SDL's positive Y goes up.
				nvimui.MouseScroll(itype.Vec2i{x, -y})
			}
		} else {
			log.Printf("main: WaitEvent: %e", err)
		}
	}

}
