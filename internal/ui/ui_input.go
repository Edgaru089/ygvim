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
	case sdl.K_RETURN:
		ui.nvim.Input("<CR>")
	case sdl.K_ESCAPE:
		ui.nvim.Input("<Esc>")
	}
}
