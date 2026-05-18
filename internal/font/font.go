package font

import (
	"log"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"edgaru089.ink/go/ygvim/fc"
	"edgaru089.ink/go/ygvim/internal/util/itype"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

const (
	DefaultFontHeight = 12
)

// Glyph describes a glyph texture rect
// as well as basic typesetting info.
type Glyph struct {
	Bounds  itype.Recti // bounding rect of the glyph, relative to its baseline
	Advance int         // amount to move horizontally the next glyph in line

	TextureRect itype.Recti // texture coords of the glyph in its texture
}

// FontFamily is filled from the client's guifont option.
type FontFamily struct {
	FamilyName string
	Style      string

	Height, Width float32
	DrawBold      bool // not associated with the word Bold in family name
}

// Fontset maintains glyph textures rects on RenderTextures
// based on given list of font family names and other options.
//
// New fontsets must have SetFontFamilies called at least once before use.
type Fontset struct {
	fonts   []FontFamily
	fontfcs []fc.Font // slice for glyph lookups, only Family and Style are filled

	fontsttf map[fc.Font]*ttf.Font

	glyphs map[rune]Glyph // cached glyphs

	renderer   *sdl.Renderer
	texture    *sdl.Texture // only one texture for now
	maxtexsize int          // max texture size, presumably on both axis
	page       page         // for rectpacking
}

// NewFontset creates with the given renderer.
func NewFontset(renderer *sdl.Renderer) *Fontset {
	// retreive max texture size
	prop := renderer.Properties().NumberProperty("SDL.renderer.max_texture_size", 8192)

	log.Printf("NewFontset: SDL.renderer.max_texture_size = %d", prop)

	return &Fontset{
		fontsttf:   make(map[fc.Font]*ttf.Font),
		glyphs:     make(map[rune]Glyph),
		renderer:   renderer,
		maxtexsize: int(prop),
		page:       page{rows: make([]row, 1), nextrow: 0},
	}
}

func (f *Fontset) clear() {
	clear(f.glyphs)
	f.fonts = f.fonts[:0]
	f.fontfcs = f.fontfcs[:0]

	// close every Font in fontsttf
	for _, ft := range f.fontsttf {
		ft.Close()
	}
	clear(f.fontsttf)

	// create a invalid 2x2 texture and its row
	if f.texture != nil {
		f.texture.Destroy()
	}
	var err error
	f.texture, err = f.renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 2, 2)
	if err != nil {
		panic(err)
	}

	f.renderer.SetRenderTarget(f.texture)
	f.renderer.SetDrawColor(255, 255, 255, 255)
	f.renderer.Clear()
	f.renderer.Present()
	f.renderer.SetRenderTarget(nil)

	f.page.rows = f.page.rows[:1]
	f.page.rows[0] = row{width: 2, height: 2, top: 0}
	f.page.nextrow = 2
}

// SetFontFamilies resets the fontset using the given font family names.
//
// Every font family name can have multiple options trailing:  (:help guifont)
//
//   - :hXX    height is XX in pixels, default = 12, same as first font if not present
//   - :wXX    width is XX in pixels, autocomputed if not present (best only in monospace fonts!)
//   - :b      bold, this bypasses the Bold word in style, only use if Bold does not work
//   - :sBold  style, non standard, 's' followed by its style name like "Bold", "Light", "Italic"
func (f *Fontset) SetFontFamilies(fonts ...string) {
	f.clear()
	f.fonts = f.fonts[:0]
	f.fontfcs = f.fontfcs[:0]
	f.fonts = slices.Grow(f.fonts, len(fonts)+1)
	f.fontfcs = slices.Grow(f.fontfcs, len(fonts)+1)
	firstFontHeight := float32(DefaultFontHeight)

	for i, name := range fonts {
		parts := strings.Split(name, ":")
		familyname := parts[0]

		ff := FontFamily{
			FamilyName: familyname,
			Width:      0,
			Height:     firstFontHeight,
			DrawBold:   false,
		}

		// parse the flags
		for _, flag := range parts[1:] {
			if len(flag) < 1 {
				continue
			}

			if flag == "b" {
				ff.DrawBold = true
			} else if flag[0] == 'h' || flag[0] == 'w' {
				num, err := strconv.ParseFloat(flag[1:], 32)
				if err != nil {
					if flag[0] == 'h' {
						ff.Height = float32(num)
						if i == 0 {
							firstFontHeight = float32(num)
						}
					} else if flag[0] == 'w' {
						ff.Width = float32(num)
					}
				} else {
					log.Printf("SetFontFamilies: error parsing number in %s: %e", flag, err)
				}
			} else if flag[0] == 's' {
				ff.Style = flag[1:]
			} else {
				log.Printf("SetFontFamilies: unknown flag: %s", flag)
			}
		}

		f.fonts = append(f.fonts, ff)

		// prepare for lookup in fontconfig
		f.fontfcs = append(f.fontfcs, fc.Font{
			Family: ff.FamilyName,
			Style:  ff.Style,
		})
	}

}

// Glyph gets the glyph of a codepoint.
//
// Each font in the font family list is tried in order.
func (f *Fontset) Glyph(c rune) Glyph {
	if glyph, ok := f.glyphs[c]; ok {
		// fast path
		return glyph
	}

	// slow path, build the glyph
	g, err := f.newGlyph(c)
	if err != nil {
		if c == unicode.ReplacementChar {
			panic("Fontset.Glyph: cant rasterize replacement char")
		}
		return f.Glyph(unicode.ReplacementChar)
	}

	return g
}
