package font

import (
	"log"

	"edgaru089.ink/go/ygvim/internal/util/itype"
	"github.com/Zyko0/go-sdl3/sdl"
)

type page struct {
	rows    []row
	nextrow int // all the row's size added together
}

type row struct {
	width, height int
	top           int
}

// packGlyphRect packs a new glyph rect into the fontset's texture.
//
// It grows the texture's size if necessary.
func (f *Fontset) packGlyphRect(page *page, width, height int) itype.Recti {
	// find the line that fits the new glyph the best
	var bestRow *row
	var bestRatio float32 = 0
	for i := range page.rows {
		thisRow := &page.rows[i]
		ratio := float32(height) / float32(thisRow.height)

		// not rows either too small or too high
		if ratio < 0.7 || ratio > 1.0 {
			continue
		}

		// not rows that can't grow enough
		// (texture's width only grows when height grows)
		if thisRow.width+width >= int(f.texture.W) {
			continue
		}

		// not rows that are not better than what we already have
		if ratio < bestRatio {
			continue
		}

		// this row is better, use it
		bestRow = thisRow
		bestRatio = ratio
	}

	// no good rows, create a new one
	if bestRow == nil {
		rowHeight := height + height/10

		// grow the texture until it fits
		newtexWidth, newtexHeight := int(f.texture.W), int(f.texture.H)
		for page.nextrow+rowHeight >= newtexHeight || width >= newtexWidth {
			// the new row needs to fit at least this new glyph, hence this second condition
			newtexWidth *= 2
			newtexHeight *= 2
		}
		if newtexWidth != int(f.texture.W) || newtexHeight != int(f.texture.H) {
			// really create a new texture
			newtex, err := f.renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, newtexWidth, newtexHeight)
			if err != nil {
				log.Printf("failed to create new texture: %e", err)
				return itype.Recti{0, 0, 2, 2}
			}

			// set the new texture as render target and flip it
			f.renderer.SetRenderTarget(newtex)
			f.renderer.SetDrawColor(255, 255, 255, 0)
			f.renderer.Clear()
			f.renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)
			f.renderer.RenderTexture(
				f.texture,
				nil,
				&sdl.FRect{
					X: 0, Y: 0,
					W: float32(f.texture.W),
					H: float32(f.texture.H),
				},
			)
			f.renderer.Present()
			f.renderer.SetRenderTarget(nil)

			// sets new texture, destroys old one
			var oldtex *sdl.Texture
			f.texture, oldtex = newtex, f.texture
			oldtex.Destroy()
		}

		// appends the new row
		f.page.rows = append(f.page.rows, row{
			width:  width,
			height: rowHeight,
			top:    f.page.nextrow,
		})
		f.page.nextrow += rowHeight
		bestRow = &f.page.rows[len(f.page.rows)-1] // back
	}

	rect := itype.Recti{
		Left:   bestRow.width,
		Top:    bestRow.top,
		Width:  width,
		Height: height,
	}
	bestRow.width += width
	return rect
}
