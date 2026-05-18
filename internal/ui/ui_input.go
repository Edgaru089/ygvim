package ui

import (
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

func (ui *UI) InputText(str string) {
	var buf strings.Builder
	for _, c := range []byte(str) {
		if c == '<' {
			buf.WriteString("<LT>")
		} else {
			buf.WriteByte(c)
		}

	}
	ui.nvim.Input(buf.String())
}

func (ui *UI) InputKey(key sdl.Keycode, mod sdl.Keymod) {
	switch key {
	case sdl.K_UP:
		ui.nvim.Input("<Up>")
	case sdl.K_DOWN:
		ui.nvim.Input("<Down>")
	case sdl.K_LEFT:
		ui.nvim.Input("<Left>")
	case sdl.K_RIGHT:
		ui.nvim.Input("<Right>")
	case sdl.K_RETURN:
		ui.nvim.Input("<CR>")
	case sdl.K_ESCAPE:
		ui.nvim.Input("<Esc>")
	case sdl.K_BACKSPACE:
		ui.nvim.Input("<BS>")
	default:
		if (mod & sdl.KMOD_CTRL) != 0 {

			ch := byte(0)
			if key >= sdl.K_0 && key <= sdl.K_9 {
				ch = (byte)(key - sdl.K_0 + '0')
			} else if key >= sdl.K_A && key <= sdl.K_Z {
				ch = (byte)(key - sdl.K_A + 'A')
			}

			if ch != 0 {
				var buf strings.Builder
				buf.Grow(8)
				buf.WriteString("<C-")
				buf.WriteByte(ch)
				buf.WriteByte('>')
				ui.nvim.Input(buf.String())
			}
		}
	}
}
