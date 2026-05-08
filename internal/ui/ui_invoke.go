package ui

import (
	"edgaru089.ink/go/ygvim/internal/util/itype"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/neovim/go-client/nvim"
)

func InvokeNvim(window *sdl.Window, width, height int, cellsize itype.Vec2i) (ui *UI) {

	ui = &UI{
		grids:  make(map[int]*Grid),
		hl:     make(map[int]Highlight),
		window: window,
		width:  width, height: height,
		cellsize: cellsize,
	}

	var err error
	ui.nvim, err = nvim.NewChildProcess(nvim.ChildProcessArgs("--embed"))
	if err != nil {
		panic(err)
	}

	ui.nvim.RegisterHandler("redraw", ui.HandleRedraw)
	ui.nvim.AttachUI(width, height, map[string]interface{}{
		"rgb":          true,
		"ext_linegrid": true,
	})

	return ui
}
