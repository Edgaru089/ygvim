package ui

import (
	"sync"

	"edgaru089.ink/go/ygvim/internal/font"
	"edgaru089.ink/go/ygvim/internal/util/itype"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/neovim/go-client/nvim"
)

// UI contains all client side data for the Nvim UI client
type UI struct {
	sync.Mutex

	width, height int // only used for now cuz single grid mode
	cellsize      itype.Vec2i

	grids map[int]*Grid

	hl map[int]Highlight // hl[0] is always the default palette

	last_hlid int // the last hl_id seen by grid_line events

	window *sdl.Window
	nvim   *nvim.Nvim

	needren  chan struct{}
	needflip chan struct{}

	fontset *font.Fontset
}

// SetSize resizes the UI. Callbacks are invoked only when
// this really changes.
func (ui *UI) SetSize(width, height int) {
	if ui.width != width || ui.height != height {
		ui.width, ui.height = width, height
		ui.nvim.TryResizeUI(width, height)
	}
}

// SetWindowSize divides the new window size and calls SetSize.
func (ui *UI) SetWindowSize(width, height int) {
	ui.SetSize(
		width/ui.cellsize[0],
		height/ui.cellsize[1],
	)
}
