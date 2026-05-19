package ui

import (
	"unicode/utf8"

	"github.com/Zyko0/go-sdl3/sdl"
)

var (
	verts2   []sdl.Vertex
	indices2 []int32
)

func normalizeTexCoord(texX, texY int, texW, texH int32) sdl.FPoint {
	return sdl.FPoint{
		X: float32(texX) / float32(texW),
		Y: float32(texY) / float32(texH),
	}
}

// updateVertices re-generates vertices & indices for the grid.
func (grid *Grid) updateVertices(ui *UI) {

	verts2 = verts2[:0]
	indices2 = indices2[:0]
	grid.vertices = grid.vertices[:0]
	grid.indices = grid.indices[:0]

	// generate all the glyphs first before actually appending verts
	for _, line := range grid.cells {
		for _, cell := range line {
			// get glyph
			if len(cell.text) == 0 {
				continue
			}
			ucs, _ := utf8.DecodeRuneInString(cell.text)
			if ucs == utf8.RuneError {
				continue
			}
			ui.fontset.Glyph(ucs)
		}
	}

	ui.fontset.FlipTexture()
	texW, texH := ui.fontset.Texture().W, ui.fontset.Texture().H

	for row, line := range grid.cells {
		for col, cell := range line {
			// first vertex indice
			id := int32(len(grid.vertices))
			offx, offy := ui.cellsize[0]*col, ui.cellsize[1]*row

			// get background & foreground colors
			hl := ui.hl[cell.hlid]
			if hl.fg.A != 255 {
				hl = ui.hl[0]
			}
			if hl.bg.A != 255 {
				hl = ui.hl[0]
			}
			fg := sdl.FColor{
				R: float32(hl.fg.R) / 255.0,
				G: float32(hl.fg.G) / 255.0,
				B: float32(hl.fg.B) / 255.0,
				A: 1.0,
			}
			bg := sdl.FColor{
				R: float32(hl.bg.R) / 255.0,
				G: float32(hl.bg.G) / 255.0,
				B: float32(hl.bg.B) / 255.0,
				A: 1.0,
			}
			if row == grid.cursor_row && col == grid.cursor_col {
				fg, bg = bg, fg
			}

			// background
			grid.vertices = append(grid.vertices,
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx), Y: float32(offy)},
					Color:    bg,
					TexCoord: normalizeTexCoord(0, 0, texW, texH),
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx + ui.cellsize[0]), Y: float32(offy)},
					Color:    bg,
					TexCoord: normalizeTexCoord(2, 0, texW, texH),
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx), Y: float32(offy + ui.cellsize[1])},
					Color:    bg,
					TexCoord: normalizeTexCoord(0, 2, texW, texH),
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx + ui.cellsize[0]), Y: float32(offy + ui.cellsize[1])},
					Color:    bg,
					TexCoord: normalizeTexCoord(2, 2, texW, texH),
				},
			)

			grid.indices = append(grid.indices, id, id+1, id+2, id+1, id+2, id+3)

			// get glyph
			if len(cell.text) == 0 {
				continue
			}
			ucs, _ := utf8.DecodeRuneInString(cell.text)
			if ucs == utf8.RuneError {
				continue
			}
			glyph := ui.fontset.Glyph(ucs)

			// foreground
			fexX, fexY, fexW, fexH := glyph.TextureRect.Left, glyph.TextureRect.Top, glyph.TextureRect.Width, glyph.TextureRect.Height
			id = int32(len(verts2))
			verts2 = append(verts2,
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx), Y: float32(offy)},
					Color:    fg,
					TexCoord: normalizeTexCoord(fexX, fexY, texW, texH),
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx + fexW), Y: float32(offy)},
					Color:    fg,
					TexCoord: normalizeTexCoord(fexX+fexW, fexY, texW, texH),
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx), Y: float32(offy + fexH)},
					Color:    fg,
					TexCoord: normalizeTexCoord(fexX, fexY+fexH, texW, texH),
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx + fexW), Y: float32(offy + fexH)},
					Color:    fg,
					TexCoord: normalizeTexCoord(fexX+fexW, fexY+fexH, texW, texH),
				},
			)
			indices2 = append(indices2, id, id+1, id+2, id+1, id+2, id+3)
		}
	}

	for _, id := range indices2 {
		grid.indices = append(grid.indices, id+int32(len(grid.vertices)))
	}
	grid.vertices = append(grid.vertices, verts2...)

}

// ui.Lock should already be locked by caller (main)
func (ui *UI) Render(renderer *sdl.Renderer) {

	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	for _, grid := range ui.grids {
		renderer.SetDrawColor(255, 255, 255, 255)
		renderer.RenderGeometry(ui.fontset.Texture(), grid.vertices, grid.indices)

	}
}
