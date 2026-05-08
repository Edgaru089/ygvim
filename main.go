package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/go-gl/gl/all-core/gl"

	"edgaru089.ink/go/ygvim/internal/ui"
	"edgaru089.ink/go/ygvim/internal/util/itype"
)

func init() {
	runtime.LockOSThread()
}

func main() {

	cellsize := itype.Vec2i{17, 34}
	width, height := 80, 25

	sdl.LoadLibrary(sdl.Path())
	defer sdl.Quit()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow(
		"Hello world",
		cellsize[0]*width, cellsize[1]*height,
		sdl.WINDOW_OPENGL|sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	// create context (per-window)
	_, err = sdl.GL_CreateContext(window)
	if err != nil {
		panic(err)
	}

	// vsync
	sdl.GL_SetSwapInterval(1)

	// initialize go-gl with given function lookup
	if err = gl.InitWithProcAddrFunc(func(name string) unsafe.Pointer {
		return unsafe.Pointer(sdl.GL_GetProcAddress(name))
	}); err != nil {
		panic(err)
	}

	nvimui := ui.InvokeNvim(window, width, height, cellsize)
	defer func(*ui.UI) {}(nvimui)

	// main loop
	running := true
	for running {
		var event sdl.Event

		for sdl.PollEvent(&event) {
			if event.Type == sdl.EVENT_QUIT {
				running = false
			}
		}

		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		sdl.GL_SwapWindow(window)
	}

	log.Printf("len([]interface{}{\"str\"})=%d", len([]any{"str"}))

}
