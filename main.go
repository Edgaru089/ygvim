package main

import (
	"runtime"

	"github.com/Zyko0/go-sdl3/sdl"

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

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	window, ren, err := sdl.CreateWindowAndRenderer(
		"gvim",
		cellsize[0]*width, cellsize[1]*height,
		sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	defer ren.Destroy()

	ren.SetVSync(1)

	nvimui := ui.InvokeNvim(window, width, height, cellsize)
	defer func(*ui.UI) {}(nvimui)

	window.StartTextInput()
	defer window.StopTextInput()

	// main loop
	running := true
	for running {
		var event sdl.Event

		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				running = false
			case sdl.EVENT_TEXT_INPUT:
				nvimui.InputText(event.TextInputEvent().Text)
			case sdl.EVENT_KEY_DOWN:
				nvimui.InputKey(event.KeyboardEvent().Key, event.KeyboardEvent().Mod)
			}
		}

		ren.SetDrawColor(0, 0, 0, 255)
		ren.Clear()

		nvimui.Render(ren)

		ren.Present()
	}

}
