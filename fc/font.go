// Package fc provides Go bindings for font matching.
//
// MatchFamily finds the best font file for a given family and optional style.
//
// MatchCodepoint finds a font file containing a Unicode codepoint, trying the
// given families in order and falling back to system fallback fonts.
package fc

// Font represents a matched font.
type Font struct {
	// File is the path to the font file.
	File string
	// Family is the font family name.
	Family string
	// Style is the font style name (e.g. "Bold", "Italic", "Regular").
	Style string
	// Index is the face index within the font file (usually 0).
	Index int
}
