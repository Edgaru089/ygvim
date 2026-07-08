package ui

import (
	"edgaru089.ink/go/ygvim/internal/util/itype"
	"github.com/Zyko0/go-sdl3/sdl"
)

func (ui *UI) ProcessEvent(event *sdl.Event) {
	switch event.Type {

	case sdl.EVENT_TEXT_INPUT:
		ui.InputText(event.TextInputEvent().Text)
	case sdl.EVENT_KEY_DOWN:
		ui.InputKey(event.KeyboardEvent().Key, event.KeyboardEvent().Mod)

	case sdl.EVENT_WINDOW_EXPOSED:
		ui.SetNeedFlip()
	case sdl.EVENT_WINDOW_RESIZED, sdl.EVENT_WINDOW_MAXIMIZED, sdl.EVENT_WINDOW_RESTORED:
		//w, h := event.WindowEvent().Data1, event.WindowEvent().Data2
		//window.SetSize(w, h)
		w, h, err := ui.window.SizeInPixels()
		if err == nil {
			nw, nh := ui.SetWindowSize(int(w), int(h))
			ui.window.SetSize(int32(nw), int32(nh))
		}

	case sdl.EVENT_MOUSE_BUTTON_DOWN:
		ui.MousePress(sdl.MouseButtonFlags(event.MouseButtonEvent().Button))
	case sdl.EVENT_MOUSE_BUTTON_UP:
		ui.MouseRelease(sdl.MouseButtonFlags(event.MouseButtonEvent().Button))
	case sdl.EVENT_MOUSE_MOTION:
		e := event.MouseMotionEvent()
		ui.MouseMove(int(e.X), int(e.Y))

	case sdl.EVENT_MOUSE_WHEEL:
		e := event.MouseWheelEvent()
		x, y := int(e.IntegerX), int(e.IntegerY)
		switch e.Direction {
		case sdl.MOUSEWHEEL_NORMAL:
		case sdl.MOUSEWHEEL_FLIPPED:
			x, y = -x, -y
		}

		// SDL's positive Y goes up.
		ui.MouseScroll(itype.Vec2i{x, -y})
	}
}
