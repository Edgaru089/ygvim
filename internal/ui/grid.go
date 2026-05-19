package ui

import (
	"image/color"
	"slices"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Cell contains a single cell in the display grid.
type Cell struct {
	text string // text in utf8
	hlid int    // every cell has this number set, as opposed to grid_line line format
	wide bool   // whether this cell should be double-width? spans the next cell and skips it
}

// Highlight describes a color palette for a cell on screen.
type Highlight struct {
	name string // name associated with highlight id, in hl_group_set commands

	fg, bg  color.RGBA // alpha is always 1.0, alpha = 0 means use default values
	special color.RGBA // color to use for various underlines, when present.

	reverse      bool // reverse video, foreground / background are switched.
	italic, bold bool

	strikethrough, underline, undercurl bool
	underdotted, underdashed            bool
	underdouble                         bool // double underline

	url string // clickable hyperlink if len!=0
}

// Grid represents a Nvim grid.
type Grid struct {
	id int // grid index. 1 for the global default

	width, height int
	cells         [][]Cell

	cursor_row, cursor_col int

	vertices []sdl.Vertex
	indices  []int32
}

// Clear clears the grid, every cell
// is set to the Zero vaule.
func (g *Grid) Clear() {
	for _, l := range g.cells {
		for j := range l {
			l[j] = Cell{}
		}
	}
}

// Resize resizes the grid slices.
// It clears new cells if there are any.
func (g *Grid) Resize(width, height int) {

	if len(g.cells) < height {
		// grow to accommodate more lines if necessary
		g.cells = slices.Grow(g.cells, height-len(g.cells))
	}
	// more cap is guaranteed now, just reslice
	g.cells = g.cells[:height]

	// resize every line slice
	for i, line := range g.cells {
		if len(line) < width {
			// grow to accommodate more lines if necessary
			g.cells[i] = slices.Grow(g.cells[i], width-len(line))
		}
		// more cap is guaranteed now, just reslice
		g.cells[i] = g.cells[i][:width]
	}

	g.width = width
	g.height = height
}
