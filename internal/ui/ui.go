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

	drawXoff, drawXoffWide, drawYoff, drawYoffWide float32 // reset to default as -999

	mouseX, mouseY int // mouse position on screen, in cells
	mouses         sdl.MouseButtonFlags
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
//
// It returns the real window size the cells really cover.
func (ui *UI) SetWindowSize(width, height int) (w, h int) {
	ui.SetSize(
		width/ui.cellsize[0],
		height/ui.cellsize[1],
	)

	w = width - width%ui.cellsize[0]
	h = height - height%ui.cellsize[1]
	return
}

func mouseKeyName(key sdl.MouseButtonFlags) string {
	switch key {
	case sdl.BUTTON_LEFT:
		return "left"
	case sdl.BUTTON_RIGHT:
		return "right"
	case sdl.BUTTON_MIDDLE:
		return "middle"
	}
	return ""
}

// MouseMove moves the mouse in screen coords
func (ui *UI) MouseMove(x, y int) {
	newX, newY := x/ui.cellsize[0], y/ui.cellsize[1]

	if ui.mouseX != newX || ui.mouseY != newY {

		ui.mouseX = newX
		ui.mouseY = newY

		// send mouse drag events
		if ui.mouses != 0 {
			for key := sdl.BUTTON_LEFT; key <= sdl.BUTTON_RIGHT; key++ {
				if ui.mouses&sdl.ButtonMask(key) != 0 {
					if keyname := mouseKeyName(key); keyname != "" {
						ui.nvim.InputMouse(keyname, "drag", "", 0, ui.mouseY, ui.mouseX)
					}
				}
			}
		}
	}
}

// MousePress marks the mouse to begin being pressed.
//
// Does nothing if the mouse is already down.
func (ui *UI) MousePress(key sdl.MouseButtonFlags) {
	if ui.mouses&sdl.ButtonMask(key) == 0 {
		ui.mouses |= sdl.ButtonMask(key)
		// send a press event
		if keyname := mouseKeyName(key); keyname != "" {
			ui.nvim.InputMouse(keyname, "press", "", 0, ui.mouseY, ui.mouseX)
		}
	}
}

// MousePress marks the mouse to stop being pressed.
//
// Does nothing if the mouse is already up.
func (ui *UI) MouseRelease(key sdl.MouseButtonFlags) {
	if ui.mouses&sdl.ButtonMask(key) != 0 {
		ui.mouses &= ^sdl.ButtonMask(key)
		// send a release event
		if keyname := mouseKeyName(key); keyname != "" {
			ui.nvim.InputMouse(keyname, "release", "", 0, ui.mouseY, ui.mouseX)
		}
	}
}

// MouseScroll sends scroll events.
// 1 for down/right, -1 for up/left.
func (ui *UI) MouseScroll(dir itype.Vec2i) {
	if dir[0] > 0 {
		ui.nvim.InputMouse("wheel", "right", "", 0, ui.mouseY, ui.mouseX)
	}
	if dir[0] < 0 {
		ui.nvim.InputMouse("wheel", "left", "", 0, ui.mouseY, ui.mouseX)
	}
	if dir[1] > 0 {
		ui.nvim.InputMouse("wheel", "down", "", 0, ui.mouseY, ui.mouseX)
	}
	if dir[1] < 0 {
		ui.nvim.InputMouse("wheel", "up", "", 0, ui.mouseY, ui.mouseX)
	}
}
