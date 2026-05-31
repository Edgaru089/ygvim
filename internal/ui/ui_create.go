package ui

import (
	"log"

	"edgaru089.ink/go/ygvim/internal/font"
	"edgaru089.ink/go/ygvim/internal/util/itype"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/neovim/go-client/nvim"
)

func InvokeNvim(window *sdl.Window, renderer *sdl.Renderer, width, height int, cellsize itype.Vec2i) (ui *UI) {

	ui = &UI{
		grids:  make(map[int]*Grid),
		hl:     make(map[int]Highlight),
		window: window,
		width:  width, height: height,
		cellsize: cellsize,
		fontset:  font.NewFontset(renderer),
		drawXoff: -999,
		needren:  make(chan struct{}, 1),
		needflip: make(chan struct{}, 1),
	}

	var err error
	ui.nvim, err = nvim.NewChildProcess(nvim.ChildProcessArgs("--embed"), nvim.ChildProcessServe(false))
	if err != nil {
		panic(err)
	}

	return ui
}

func (ui *UI) Reconnect(target string) (err error) {
	if ui.nvim != nil {
		ui.nvim.DetachUI() // errors are ignored
		ui.nvim.Close()
	}

	ui.nvim, err = nvim.Dial(target)
	if err != nil {
		return err
	}
	return nil
}

func (ui *UI) TryConnectNextAddr() bool {
	if ui.nextaddr != "" {
		err := ui.Reconnect(ui.nextaddr)
		ui.nextaddr = ""
		if err != nil {
			log.Printf("TryConnectNextAddr: dial error: %s", err.Error())
			return false
		}
		return true
	}
	return false
}

func (ui *UI) Serve() error {
	return ui.nvim.Serve()
}

func (ui *UI) AttachUI() {
	ui.nvim.RegisterHandler("redraw", ui.HandleRedraw)
	ui.nvim.AttachUI(ui.width, ui.height, map[string]interface{}{
		"rgb":          true,
		"ext_linegrid": true,
	})

}
