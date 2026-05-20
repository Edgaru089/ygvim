package ui

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"edgaru089.ink/go/ygvim/internal/util"
	"edgaru089.ink/go/ygvim/internal/util/itype"
	"github.com/Zyko0/go-sdl3/sdl"
)

var (
	handler  map[string]func(ui *UI, args []any)
	stackbuf [4096]byte
)

// shorthand to cast MsgPack ints (that are sometimes int64, sometimes uint64) into plain int
//
// Yes i know int is shorter, no i don't care F you
func intf64(i any) int {
	v := reflect.ValueOf(i)
	if v.CanInt() {
		return (int)(v.Int())
	}
	if v.CanUint() {
		return (int)(v.Uint())
	}

	panic(fmt.Errorf("intf64: %T is not int64/uint64", i))
}

func init() {
	handler = make(map[string]func(ui *UI, args []any))

	//////// global events ////////

	handler["set_icon"] = func(ui *UI, args []any) {}
	handler["set_title"] = func(ui *UI, args []any) {
		ui.window.SetTitle(args[0].(string))
	}
	handler["busy_start"] = func(ui *UI, args []any) { sdl.HideCursor() }
	handler["busy_stop"] = func(ui *UI, args []any) { sdl.ShowCursor() }
	handler["bell"] = func(ui *UI, args []any) {}
	handler["visual_bell"] = func(ui *UI, args []any) { /* TODO bells */ }

	handler["flush"] = func(ui *UI, args []any) {
		ui.SetNeedRender()
	}

	handler["option_set"] = func(ui *UI, args []any) {
		if len(args) < 2 {
			return
		}
		switch args[0].(string) {
		case "guifont":
			var cellSize itype.Vec2i
			// custom flags:
			//  - xXX:  cell width XX, default ceiling(first font size)/2
			//  - yXX:  cell height XX, default ceiling(first font size)
			ui.fontset.SetFontFamiliesEx(func(flag string) {
				if len(flag) < 1 {
					return
				}
				switch flag[0] {
				case 'x':
					num, err := strconv.ParseInt(flag[1:], 10, 32)
					if err != nil {
						log.Printf("invalid flag %s: %e", flag, err)
					}
					cellSize[0] = int(num)
				case 'y':
					num, err := strconv.ParseInt(flag[1:], 10, 32)
					if err != nil {
						log.Printf("invalid flag %s: %e", flag, err)
					}
					cellSize[1] = int(num)
				default:
					log.Printf("option_set callback: unknown flag: %s", flag)
				}
			}, strings.Split(args[1].(string), ",")...)

			if cellSize[1] == 0 {
				cellSize[1] = int(math.Ceil(float64(ui.fontset.FirstFontHeight()) * 1.05))
				if cellSize[1]%2 == 1 {
					cellSize[1]++
				}
			}
			if cellSize[0] == 0 {
				cellSize[0] = cellSize[1] / 2
			}

			if ui.cellsize != cellSize {
				ui.cellsize = cellSize
				// resize the window
				ui.window.SetSize(int32(ui.cellsize[0]*ui.width), int32(ui.cellsize[1]*ui.height))
			}

			ui.drawXoff = -999
			ui.SetNeedRender()
		}
	}

	//////// grid events (line based) ////////

	handler["grid_resize"] = func(ui *UI, args []any) {
		grid_id := intf64(args[0])
		grid := ui.grids[grid_id]
		if grid == nil {
			grid = &Grid{}
			ui.grids[grid_id] = grid
		}

		grid.Resize(intf64(args[1]), intf64(args[2]))
	}

	handler["default_colors_set"] = func(ui *UI, args []any) {
		rgb_fg, rgb_bg, rgb_sp := intf64(args[0]), intf64(args[1]), intf64(args[2])
		ui.hl[0] = Highlight{
			fg:      util.RGB1(rgb_fg),
			bg:      util.RGB1(rgb_bg),
			special: util.RGB1(rgb_sp),
			// every other flag is false.
		}
	}

	handler["hl_attr_define"] = func(ui *UI, args []any) {
		hl := Highlight{}
		id := intf64(args[0])
		rgb_attr := args[1].(map[string]any)

		// every key is optional, so iterate and set each key
		for key, val := range rgb_attr {
			switch key {
			case "foreground":
				hl.fg = util.RGB1(intf64(val))
			case "background":
				hl.bg = util.RGB1(intf64(val))
			case "special":
				hl.special = util.RGB1(intf64(val))

			case "reverse":
				hl.reverse = val.(bool)
			case "italic":
				hl.italic = val.(bool)
			case "bold":
				hl.bold = val.(bool)
			case "strikethrough":
				hl.strikethrough = val.(bool)
			case "underline":
				hl.underline = val.(bool)
			case "undercurl":
				hl.undercurl = val.(bool)
			case "underdouble":
				hl.underdouble = val.(bool)
			case "underdotted":
				hl.underdotted = val.(bool)
			case "underdashed":
				hl.underdashed = val.(bool)
			case "url":
				hl.url = val.(string)
			}
		}

		ui.hl[id] = hl
	}

	//handler["hl_group_set"] = func(ui *UI, args []any) {}

	handler["grid_line"] = func(ui *UI, args []any) {
		grid := ui.grids[intf64(args[0])]

		row := intf64(args[1])
		col := intf64(args[2])
		cells := args[3].([]any)
		hlid := ui.last_hlid

		for _, cell_any := range cells {
			cell := cell_any.([]any)
			text := cell[0].(string)

			// set hl_id if present
			if len(cell) >= 2 {
				hlid = intf64(cell[1])
			}

			// get repeat times
			repeat := 1
			if len(cell) >= 3 {
				repeat = intf64(cell[2])
			}

			// write the cells
			for i := 0; i < repeat; i++ {
				grid.cells[row][col] = Cell{
					text: text,
					hlid: hlid,
					wide: util.IsCharWide(text),
				}
				col++
			}
		}
	}

	handler["grid_clear"] = func(ui *UI, args []any) {
		ui.grids[intf64(args[0])].Clear()
	}
	handler["grid_cursor_goto"] = func(ui *UI, args []any) {
		grid := ui.grids[intf64(args[0])]

		grid.cursor_row = intf64(args[1])
		grid.cursor_col = intf64(args[2])

		ui.window.SetTextInputArea(
			&sdl.Rect{
				X: int32(grid.cursor_col * ui.cellsize[0]),
				Y: int32(grid.cursor_row * ui.cellsize[1]),
				W: int32(ui.cellsize[0]),
				H: int32(ui.cellsize[1]),
			}, 0)
	}

	handler["grid_scroll"] = func(ui *UI, args []any) {
		grid := ui.grids[intf64(args[0])]
		top, bot := intf64(args[1]), intf64(args[2])
		left, right := intf64(args[3]), intf64(args[4])
		rows, _ := intf64(args[5]), intf64(args[6])

		if rows > 0 {
			// move rectangle up
			for line := top + rows; line < bot; line++ {
				copy(grid.cells[line-rows][left:right], grid.cells[line][left:right])
			}
		} else if rows < 0 {
			// move rectangle down
			for line := bot + rows - 1; line >= top; line-- {
				copy(grid.cells[line-rows][left:right], grid.cells[line][left:right])
			}
		}
	}

	//////// others ////////
	handler["win_viewport"] = func(ui *UI, args []any) {}
	handler["mode_info_set"] = func(ui *UI, args []any) {}
	handler["mode_change"] = func(ui *UI, args []any) {}
	handler["mouse_on"] = func(ui *UI, args []any) {}
	handler["mouse_off"] = func(ui *UI, args []any) {}
}

// callRedrawEvent calls the given message, recovering if paniced.
//
// args should be []any, but just in case it panics the cast is done here to recover.
//
// Every event comes in this format: [ 'message_type', [args...], [args...] ]
// where every [args...] slice following 'message_type' calls the message RPC
// once with the given arguments, therefore the RPC (and also this function)
// sometimes get called multiple times in a single event, once for every [args...] slice.
func (ui *UI) callRedrawEvent(msg string, args any, f func(*UI, []any)) {
	defer func() {
		err := recover()
		len := runtime.Stack(stackbuf[:], false)
		if err != nil {
			log.Printf("HandleRedraw: panic('%s') processing message('%s') %v\n     stack: %s", err, msg, args, stackbuf[:len])
		}
	}()

	ui.Lock()
	defer ui.Unlock()

	f(ui, args.([]any))
}

// HandleRedraw handles a redraw message from Nvim.
func (ui *UI) HandleRedraw(args ...[]any) {

	for _, e := range args {
		msg, ok := e[0].(string)
		if !ok {
			log.Printf("HandleRedraw: message has no type string: %v", e)
			continue
		}

		f, ok := handler[msg]
		if ok {
			// multiple calls in same event
			for _, args := range e[1:] {
				ui.callRedrawEvent(msg, args, f)
			}
		} else {
			log.Printf("HandleRedraw: ignoring message type %s", msg)
		}
	}
}
