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
		for {
			err := nvimui.Serve()
			if err != nil {
				log.Printf("main: nvim.Serve: %e", err)
			}

			// if we land here, the child process quited
			if !nvimui.TryConnectNextAddr() {
				break
			}
			nvimui.AttachUI()
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
			default:
				nvimui.ProcessEvent(&event)
			}
		} else {
			log.Printf("main: WaitEvent: %e", err)
		}
	}

}
