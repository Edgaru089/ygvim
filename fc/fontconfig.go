// Package fc provides Go bindings for the fontconfig library.
//
// MatchFamily finds the best font file for a given family and optional style.
//
// MatchCodepoint finds a font file containing a Unicode codepoint, trying the
// given families in order and falling back to fontconfig's configured fallbacks.
package fc

/*
#cgo LDFLAGS: -lfontconfig
#include <fontconfig/fontconfig.h>
#include <stdlib.h>

extern const char *fc_object_family;
extern const char *fc_object_style;
extern const char *fc_object_file;
extern const char *fc_object_index;
extern const char *fc_object_charset;
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Font represents a matched font pattern.
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

// MatchFamily finds the best matching font file for the given family name and
// optional style description. style can be "" to match any style, or a
// description like "Bold", "Regular", "Light", "Italic", etc.
func MatchFamily(family, style string) (*Font, error) {
	if family == "" {
		return nil, fmt.Errorf("fc: family name must not be empty")
	}

	pattern := C.FcPatternCreate()
	if pattern == nil {
		return nil, fmt.Errorf("fc: failed to create pattern")
	}
	defer C.FcPatternDestroy(pattern)

	cFamily := C.CString(family)
	defer C.free(unsafe.Pointer(cFamily))
	C.FcPatternAddString(pattern, C.fc_object_family, (*C.FcChar8)(unsafe.Pointer(cFamily)))

	if style != "" {
		cStyle := C.CString(style)
		defer C.free(unsafe.Pointer(cStyle))
		C.FcPatternAddString(pattern, C.fc_object_style, (*C.FcChar8)(unsafe.Pointer(cStyle)))
	}

	config := C.FcConfigGetCurrent()
	C.FcConfigSubstitute(config, pattern, C.FcMatchPattern)
	C.FcDefaultSubstitute(pattern)

	var result C.FcResult
	matched := C.FcFontMatch(config, pattern, &result)
	if matched == nil {
		return nil, fmt.Errorf("fc: FcFontMatch returned nil pattern")
	}
	defer C.FcPatternDestroy(matched)

	return extractFont(matched)
}

// MatchCodepoint finds a font file that can render the given Unicode codepoint.
// families is an optional list of preferred font families, tried in the order
// given. If no family from the list contains the codepoint, fontconfig's
// configured fallback fonts are used.
func MatchCodepoint(codepoint rune, families ...string) (*Font, error) {
	config := C.FcConfigGetCurrent()

	for _, family := range families {
		f, err := matchCodepointFamily(config, codepoint, family)
		if err != nil {
			continue
		}
		return f, nil
	}

	return matchCodepointFallback(config, codepoint)
}

func matchCodepointFamily(config *C.FcConfig, codepoint rune, family string) (*Font, error) {
	pattern := C.FcPatternCreate()
	if pattern == nil {
		return nil, fmt.Errorf("fc: failed to create pattern")
	}
	defer C.FcPatternDestroy(pattern)

	charset := C.FcCharSetCreate()
	if charset == nil {
		return nil, fmt.Errorf("fc: failed to create char set")
	}
	defer C.FcCharSetDestroy(charset)

	C.FcCharSetAddChar(charset, C.FcChar32(codepoint))
	C.FcPatternAddCharSet(pattern, C.fc_object_charset, charset)

	cFamily := C.CString(family)
	defer C.free(unsafe.Pointer(cFamily))
	C.FcPatternAddString(pattern, C.fc_object_family, (*C.FcChar8)(unsafe.Pointer(cFamily)))

	C.FcConfigSubstitute(config, pattern, C.FcMatchPattern)
	C.FcDefaultSubstitute(pattern)

	var result C.FcResult
	matched := C.FcFontMatch(config, pattern, &result)
	if matched == nil {
		return nil, fmt.Errorf("fc: FcFontMatch returned nil pattern")
	}
	defer C.FcPatternDestroy(matched)

	return extractFont(matched)
}

func matchCodepointFallback(config *C.FcConfig, codepoint rune) (*Font, error) {
	pattern := C.FcPatternCreate()
	if pattern == nil {
		return nil, fmt.Errorf("fc: failed to create pattern")
	}
	defer C.FcPatternDestroy(pattern)

	charset := C.FcCharSetCreate()
	if charset == nil {
		return nil, fmt.Errorf("fc: failed to create char set")
	}
	defer C.FcCharSetDestroy(charset)

	C.FcCharSetAddChar(charset, C.FcChar32(codepoint))
	C.FcPatternAddCharSet(pattern, C.fc_object_charset, charset)

	C.FcConfigSubstitute(config, pattern, C.FcMatchPattern)
	C.FcDefaultSubstitute(pattern)

	var result C.FcResult
	matched := C.FcFontMatch(config, pattern, &result)
	if matched == nil {
		return nil, fmt.Errorf("fc: no font found containing U+%04X", codepoint)
	}
	defer C.FcPatternDestroy(matched)

	return extractFont(matched)
}

func extractFont(pattern *C.FcPattern) (*Font, error) {
	var (
		cFile   *C.FcChar8
		cFamily *C.FcChar8
		cStyle  *C.FcChar8
		index   C.int
	)

	if C.FcPatternGetString(pattern, C.fc_object_file, 0, &cFile) != C.FcResultMatch {
		return nil, fmt.Errorf("fc: pattern has no file")
	}
	if C.FcPatternGetString(pattern, C.fc_object_family, 0, &cFamily) != C.FcResultMatch {
		return nil, fmt.Errorf("fc: pattern has no family")
	}
	// Style and index are optional.
	C.FcPatternGetString(pattern, C.fc_object_style, 0, &cStyle)
	C.FcPatternGetInteger(pattern, C.fc_object_index, 0, &index)

	return &Font{
		File:   C.GoString((*C.char)(unsafe.Pointer(cFile))),
		Family: C.GoString((*C.char)(unsafe.Pointer(cFamily))),
		Style:  C.GoString((*C.char)(unsafe.Pointer(cStyle))),
		Index:  int(index),
	}, nil
}
