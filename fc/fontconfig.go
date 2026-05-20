//go:build !windows

package fc

/*
#cgo LDFLAGS: -lfontconfig
#include <fontconfig/fontconfig.h>
#include <stdlib.h>

const char *fc_object_family  = "family";
const char *fc_object_style   = "style";
const char *fc_object_file    = "file";
const char *fc_object_index   = "index";
const char *fc_object_charset = "charset";
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// MatchFamily finds the best matching font file for the given family name and
// optional style description. style can be "" to match any style, or a
// description like "Bold", "Regular", "Light", "Italic", etc.
//
// If no font matches the requested family, the returned Font will have an
// empty File field and no error.
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

	if !patternFamilyMatches(matched, family) {
		return &Font{}, nil
	}
	return extractFont(matched)
}

// patternFamilyMatches reports whether any FC_FAMILY value on pattern
// case-insensitively matches the requested family name. A single font may
// carry the family name in multiple languages (e.g. "DengXian" and "等线").
func patternFamilyMatches(pattern *C.FcPattern, requested string) bool {
	cReq := C.CString(requested)
	defer C.free(unsafe.Pointer(cReq))

	for i := C.int(0); ; i++ {
		var cFamily *C.FcChar8
		if C.FcPatternGetString(pattern, C.fc_object_family, i, &cFamily) != C.FcResultMatch {
			break
		}
		if cFamily == nil {
			continue
		}
		if C.FcStrCmp((*C.FcChar8)(unsafe.Pointer(cReq)), cFamily) == 0 {
			return true
		}
	}
	return false
}

// MatchCodepoint finds a font file that can render the given Unicode codepoint.
// families is a list of preferred fonts, each specifying a Family and
// optionally a Style (like "Bold", "Light"). The list is tried in order.
// If no font from the list contains the codepoint, fontconfig's configured
// fallback fonts are used.
//
// The returned int is the index of the matching family in families, or
// len(families) if a fallback font was selected.
func MatchCodepoint(codepoint rune, families []Font) (*Font, int, error) {
	config := C.FcConfigGetCurrent()

	for i, fam := range families {
		f, err := matchCodepointFamily(config, codepoint, fam.Family, fam.Style)
		if err != nil {
			continue
		}
		return f, i, nil
	}

	f, err := matchCodepointFallback(config, codepoint)
	return f, len(families), err
}

func matchCodepointFamily(config *C.FcConfig, codepoint rune, family, style string) (*Font, error) {
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

	if style != "" {
		cStyle := C.CString(style)
		defer C.free(unsafe.Pointer(cStyle))
		C.FcPatternAddString(pattern, C.fc_object_style, (*C.FcChar8)(unsafe.Pointer(cStyle)))
	}

	C.FcConfigSubstitute(config, pattern, C.FcMatchPattern)
	C.FcDefaultSubstitute(pattern)

	var result C.FcResult
	matched := C.FcFontMatch(config, pattern, &result)
	if matched == nil {
		return nil, fmt.Errorf("fc: FcFontMatch returned nil pattern")
	}
	defer C.FcPatternDestroy(matched)

	if !patternFamilyMatches(matched, family) {
		return nil, fmt.Errorf("fc: no match for family %q", family)
	}
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
