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
		if thisRow.width+width >= int(f.image.W) {
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
		newimgWidth, newimgHeight := int(f.image.W), int(f.image.H)
		for page.nextrow+rowHeight >= newimgHeight || width >= newimgWidth {
			// the new row needs to fit at least this new glyph, hence this second condition
			newimgWidth *= 2
			newimgHeight *= 2
		}
		if newimgWidth != int(f.image.W) || newimgHeight != int(f.image.H) {
			// really create a new image
			newimg, err := sdl.CreateSurface(newimgWidth, newimgHeight, sdl.PIXELFORMAT_ARGB8888)
			if err != nil {
				log.Printf("failed to create new image: %e", err)
				return itype.Recti{0, 0, 2, 2}
			}
			newimg.Clear(1, 1, 1, 0)

			// copy to new image
			f.image.SetBlendMode(sdl.BLENDMODE_NONE)
			f.image.Blit(nil, newimg, nil)

			// sets new texture, destroys old one
			var oldimg *sdl.Surface
			f.image, oldimg = newimg, f.image
			oldimg.Destroy()
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

// FlipTexture flips the internal image buffer into the real texture.
func (f *Fontset) FlipTexture() {
	if f.texture == nil || f.image.W != f.texture.W || f.image.H != f.texture.H {
		newtex, err := f.renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, int(f.image.W), int(f.image.H))
		if err != nil {
			log.Printf("failed to create new texture: %e", err)
			return
		}
		setNewTextureProps(newtex)

		if f.texture != nil {
			f.texture.Destroy()
		}
		f.texture = newtex
	}

	f.texture.Update(nil, f.image.Pixels(), f.image.Pitch)
}
