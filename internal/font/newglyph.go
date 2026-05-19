package font

import (
	"fmt"
	"log"

	"edgaru089.ink/go/ygvim/fc"
	"edgaru089.ink/go/ygvim/internal/util/itype"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

// setNewTextureProps sets the scaling mode of new textures to nearest.
func setNewTextureProps(tex *sdl.Texture) {
	tex.SetScaleMode(sdl.SCALEMODE_NEAREST)
}

// newGlyph really rasterizes a new glyph and puts it into
// a new rectpack in the texture.
func (f *Fontset) newGlyph(c rune) (g Glyph, err error) {

	// match the codepoint
	fcFont, i, err := fc.MatchCodepoint(c, f.fontfcs)
	if err != nil {
		return
	}

	ft, err := f.getFontTTF(fcFont, i)
	if err != nil {
		return
	}

	// we have the TTF font instance, get the metrics first
	mt, err := ft.GlyphMetrics(uint32(c))
	if err != nil {
		return
	}
	g.Advance = int(mt.Advance)
	g.Bounds = itype.Recti{
		Left:   int(mt.MinX),
		Top:    int(mt.MinY),
		Width:  int(mt.MaxX - mt.MinX),
		Height: int(mt.MaxY - mt.MinY),
	}

	// draw the glyph
	// this returns a ARGB8888 image, colored if it's an emoji, white with alpha (non-premultiplied) otherwise
	// it actually tells us which one it returned in ttf.ImageType, but we don't care
	surface, _, err := ft.GlyphImage(uint32(c))
	if err != nil {
		return
	}

	// now that we have the image, then we allocate a rectpack
	rect := f.packGlyphRect(&f.page, int(surface.W), int(surface.H))

	// we draw
	surface.SetBlendMode(sdl.BLENDMODE_NONE)
	surface.Blit(nil, f.image, &sdl.Rect{X: int32(rect.Left), Y: int32(rect.Top)})

	g.TextureRect = rect

	f.glyphs[c] = g
	return g, nil
}

// getFontTTF obtains the TTF font instance, either from cache or from
// opening a new instance.
func (f *Fontset) getFontTTF(fn *fc.Font, i int) (ft *ttf.Font, err error) {
	if ft, ok := f.fontsttf[*fn]; ok {
		return ft, nil
	}

	// gets the font's point size.
	var ptsize float32
	if i < len(f.fonts) {
		ptsize = f.fonts[i].Height
	} else if len(f.fonts) > 0 {
		ptsize = f.fonts[0].Height
	} else {
		ptsize = DefaultFontHeight
	}

	// open.
	log.Printf("Opening new font: %v", fn)
	ft, err = ttf.OpenFont(fn.File, ptsize)
	if err != nil {
		return nil, fmt.Errorf("getFontTTF: error opening font: %e", err)
	}

	f.fontsttf[*fn] = ft
	return ft, nil
}
