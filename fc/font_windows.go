//go:build windows

package fc

import (
	"fmt"
	"syscall"
	"unsafe"
)

// DirectWrite font matching via dwrite.dll COM interfaces.
// Vtable indices verified against dwrite.h (mingw-w64).

var (
	moddwrite                      = syscall.NewLazyDLL("dwrite.dll")
	procDWriteCreateFactory        = moddwrite.NewProc("DWriteCreateFactory")
	iidIDWriteFactory              = windowsGUID{0xb859ee5a, 0xd838, 0x4b5b, [8]byte{0xa2, 0xe8, 0x1a, 0xdc, 0x7d, 0x93, 0xdb, 0x48}}
	iidIDWriteLocalFontFileLoader = windowsGUID{0xb2d9f3ec, 0xc9fe, 0x4a11, [8]byte{0xa2, 0xec, 0xd8, 0x62, 0x08, 0xf7, 0xc0, 0xa2}}
)

type windowsGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

const (
	vtbl_QueryInterface = 0
	vtbl_AddRef         = 1
	vtbl_Release        = 2
)

// --- dwriteObject: base COM pointer ---

type dwriteObject struct {
	p    uintptr
	vtbl *uintptr
}

func newDWriteObject(p uintptr) dwriteObject {
	// p is a COM interface pointer; the first field is a pointer to the vtable.
	vtbl := *(**uintptr)(unsafe.Pointer(p))
	return dwriteObject{p: p, vtbl: vtbl}
}

func (o *dwriteObject) call(methodIndex int, args ...uintptr) (uintptr, uintptr, uintptr) {
	all := make([]uintptr, 0, len(args)+1)
	all = append(all, o.p)
	all = append(all, args...)
	fn := *(*uintptr)(unsafe.Add(unsafe.Pointer(o.vtbl), uintptr(methodIndex)*unsafe.Sizeof(uintptr(0))))
	a, b, c := syscall.SyscallN(fn, all...)
	return a, b, uintptr(c)
}

// --- MatchFamily ---

func MatchFamily(family, style string) (*Font, error) {
	if family == "" {
		return nil, fmt.Errorf("fc: family name must not be empty")
	}

	dw, err := createFactory()
	if err != nil {
		return nil, err
	}
	defer dw.release()

	coll := dw.getSystemFontCollection()
	if coll.p == 0 {
		return nil, fmt.Errorf("fc: failed to get system font collection")
	}
	defer coll.release()

	idx, exists := coll.findFamilyName(family)
	if !exists {
		return &Font{}, nil
	}

	fam := coll.getFontFamily(idx)
	if fam.p == 0 {
		return nil, fmt.Errorf("fc: failed to get font family")
	}
	defer fam.release()

	font, err := fam.matchFont(style)
	if err != nil {
		return nil, err
	}
	if font.p == 0 {
		return &Font{}, nil
	}
	defer font.release()

	face := font.createFontFace()
	if face.p == 0 {
		return nil, fmt.Errorf("fc: failed to create font face")
	}
	defer face.release()

	return fontFaceToFont(face, family, font.weight(), font.styleStr())
}

// --- MatchCodepoint ---

func MatchCodepoint(codepoint rune, families []Font) (*Font, int, error) {
	dw, err := createFactory()
	if err != nil {
		return nil, 0, err
	}
	defer dw.release()

	coll := dw.getSystemFontCollection()
	if coll.p == 0 {
		return nil, 0, fmt.Errorf("fc: failed to get system font collection")
	}
	defer coll.release()

	for i, fam := range families {
		f, err := matchCodepointInFamily(coll, codepoint, fam.Family, fam.Style)
		if err != nil {
			continue
		}
		if f != nil {
			return f, i, nil
		}
	}

	f, err := matchCodepointFallback(coll, codepoint)
	return f, len(families), err
}

func matchCodepointInFamily(coll *dwriteFontCollection, codepoint rune, family, style string) (*Font, error) {
	idx, exists := coll.findFamilyName(family)
	if !exists {
		return nil, fmt.Errorf("fc: family %q not found", family)
	}

	fam := coll.getFontFamily(idx)
	if fam.p == 0 {
		return nil, fmt.Errorf("fc: failed to get font family")
	}
	defer fam.release()

	font, err := fam.matchFont(style)
	if err != nil {
		return nil, err
	}
	if font.p == 0 {
		return nil, fmt.Errorf("fc: no matching font in family")
	}
	defer font.release()

	if !font.hasCharacter(codepoint) {
		return nil, fmt.Errorf("fc: font does not contain U+%04X", codepoint)
	}

	face := font.createFontFace()
	if face.p == 0 {
		return nil, fmt.Errorf("fc: failed to create font face")
	}
	defer face.release()

	return fontFaceToFont(face, family, font.weight(), font.styleStr())
}

func matchCodepointFallback(coll *dwriteFontCollection, codepoint rune) (*Font, error) {
	count := coll.fontFamilyCount()
	for i := uint32(0); i < count; i++ {
		fam := coll.getFontFamily(i)
		if fam.p == 0 {
			continue
		}
		fc := fam.fontCount()
		if fc == 0 {
			fam.release()
			continue
		}
		font := fam.getFont(0)
		if font.p == 0 {
			fam.release()
			continue
		}
		if font.hasCharacter(codepoint) {
			face := font.createFontFace()
			if face.p != 0 {
				f, err := fontFaceToFont(face, "", font.weight(), font.styleStr())
				face.release()
				if err != nil {
					font.release()
					fam.release()
					continue
				}
				names := fam.familyNames()
				font.release()
				fam.release()
				if names.p != 0 {
					if s := names.localeString(0); s != "" {
						f.Family = s
					}
					names.release()
				}
				return f, nil
			}
		}
		font.release()
		fam.release()
	}
	return nil, fmt.Errorf("fc: no font found containing U+%04X", codepoint)
}

// --- dwriteFactory (vtable start: 3) ---

type dwriteFactory struct{ dwriteObject }

func createFactory() (*dwriteFactory, error) {
	var p uintptr
	r, _, _ := procDWriteCreateFactory.Call(
		0, // DWRITE_FACTORY_TYPE_SHARED
		uintptr(unsafe.Pointer(&iidIDWriteFactory)),
		uintptr(unsafe.Pointer(&p)),
	)
	if r != 0 {
		return nil, fmt.Errorf("fc: DWriteCreateFactory failed with HRESULT 0x%X", r)
	}
	return &dwriteFactory{newDWriteObject(p)}, nil
}

func (f *dwriteFactory) release()                     { f.call(vtbl_Release) }
func (f *dwriteFactory) getSystemFontCollection() *dwriteFontCollection {
	var coll uintptr
	f.call(3, uintptr(unsafe.Pointer(&coll)), 0)
	return &dwriteFontCollection{newDWriteObject(coll)}
}

// --- dwriteFontCollection (vtable start: 3) ---

type dwriteFontCollection struct{ dwriteObject }

func (c *dwriteFontCollection) release()                         { c.call(vtbl_Release) }
func (c *dwriteFontCollection) fontFamilyCount() uint32           { r, _, _ := c.call(3); return uint32(r) }
func (c *dwriteFontCollection) getFontFamily(index uint32) *dwriteFontFamily {
	var fam uintptr
	c.call(4, uintptr(index), uintptr(unsafe.Pointer(&fam)))
	return &dwriteFontFamily{newDWriteObject(fam)}
}
func (c *dwriteFontCollection) findFamilyName(name string) (index uint32, exists bool) {
	u16, _ := syscall.UTF16PtrFromString(name)
	var idx uint32
	var ex int32
	c.call(5, uintptr(unsafe.Pointer(u16)), uintptr(unsafe.Pointer(&idx)), uintptr(unsafe.Pointer(&ex)))
	return idx, ex != 0
}

// --- dwriteFontFamily (vtable start: 6, inherits IDWriteFontList at 3-5) ---

type dwriteFontFamily struct{ dwriteObject }

func (f *dwriteFontFamily) release()       { f.call(vtbl_Release) }
func (f *dwriteFontFamily) fontCount() uint32 { r, _, _ := f.call(4); return uint32(r) }

func (f *dwriteFontFamily) getFont(index uint32) *dwriteFont {
	var font uintptr
	f.call(5, uintptr(index), uintptr(unsafe.Pointer(&font)))
	return &dwriteFont{newDWriteObject(font)}
}

func (f *dwriteFontFamily) matchFont(style string) (*dwriteFont, error) {
	weight, s, stretch := parseStyle(style)
	return f.getFirstMatchingFont(weight, s, stretch)
}

func (f *dwriteFontFamily) getFirstMatchingFont(weight, s, stretch uint32) (*dwriteFont, error) {
	var font uintptr
	f.call(7, uintptr(weight), uintptr(stretch), uintptr(s), uintptr(unsafe.Pointer(&font)))
	if font == 0 {
		return nil, fmt.Errorf("fc: no matching font in family for style")
	}
	return &dwriteFont{newDWriteObject(font)}, nil
}

func (f *dwriteFontFamily) familyNames() *dwriteLocalizedStrings {
	var names uintptr
	f.call(6, uintptr(unsafe.Pointer(&names)))
	return &dwriteLocalizedStrings{newDWriteObject(names)}
}

// --- dwriteFont (vtable start: 3) ---

type dwriteFont struct{ dwriteObject }

func (f *dwriteFont) release()               { f.call(vtbl_Release) }
func (f *dwriteFont) weight() uint32          { r, _, _ := f.call(4); return uint32(r) }
func (f *dwriteFont) stretch() uint32         { r, _, _ := f.call(5); return uint32(r) }
func (f *dwriteFont) style() uint32           { r, _, _ := f.call(6); return uint32(r) }
func (f *dwriteFont) faceNames() *dwriteLocalizedStrings {
	var names uintptr
	f.call(8, uintptr(unsafe.Pointer(&names)))
	return &dwriteLocalizedStrings{newDWriteObject(names)}
}
func (f *dwriteFont) hasCharacter(codepoint rune) bool {
	var exists int32
	f.call(12, uintptr(codepoint), uintptr(unsafe.Pointer(&exists)))
	return exists != 0
}
func (f *dwriteFont) createFontFace() *dwriteFontFace {
	var face uintptr
	f.call(13, uintptr(unsafe.Pointer(&face)))
	return &dwriteFontFace{newDWriteObject(face)}
}

func (f *dwriteFont) styleStr() string {
	switch f.style() {
	case 1:
		return "Oblique"
	case 2:
		return "Italic"
	default:
		return "Regular"
	}
}

// --- dwriteFontFace (vtable start: 3) ---

type dwriteFontFace struct{ dwriteObject }

func (f *dwriteFontFace) release()  { f.call(vtbl_Release) }
func (f *dwriteFontFace) index() uint32 { r, _, _ := f.call(5); return uint32(r) }

func (f *dwriteFontFace) getFiles() *dwriteFontFile {
	var n uint32 = 1
	var file uintptr
	f.call(4, uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&file)))
	if n == 0 || file == 0 {
		return nil
	}
	return &dwriteFontFile{newDWriteObject(file)}
}

// --- dwriteFontFile (vtable start: 3) ---

type dwriteFontFile struct{ dwriteObject }

func (f *dwriteFontFile) release() { f.call(vtbl_Release) }

func (f *dwriteFontFile) referenceKey() []byte {
	var key unsafe.Pointer
	var size uint32
	f.call(3, uintptr(unsafe.Pointer(&key)), uintptr(unsafe.Pointer(&size)))
	if key == nil || size == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(key), size)
}

func (f *dwriteFontFile) getLoader() *dwriteLocalFontFileLoader {
	var loader uintptr
	r, _, _ := f.call(4, uintptr(unsafe.Pointer(&loader)))
	if r != 0 || loader == 0 {
		return nil
	}
	var localLoader uintptr
	obj := newDWriteObject(loader)
	obj.call(vtbl_QueryInterface, uintptr(unsafe.Pointer(&iidIDWriteLocalFontFileLoader)), uintptr(unsafe.Pointer(&localLoader)))
	obj.call(vtbl_Release)
	if localLoader == 0 {
		return nil
	}
	return &dwriteLocalFontFileLoader{newDWriteObject(localLoader)}
}

// --- dwriteLocalFontFileLoader (vtable start: 4, inherits CreateStreamFromKey at 3) ---

type dwriteLocalFontFileLoader struct{ dwriteObject }

func (l *dwriteLocalFontFileLoader) release() { l.call(vtbl_Release) }

func (l *dwriteLocalFontFileLoader) getFilePathFromKey(key []byte) string {
	var pathLen uint32
	l.call(4, uintptr(unsafe.Pointer(unsafe.SliceData(key))), uintptr(len(key)), uintptr(unsafe.Pointer(&pathLen)))
	if pathLen == 0 {
		return ""
	}
	buf := make([]uint16, pathLen+1)
	l.call(5, uintptr(unsafe.Pointer(unsafe.SliceData(key))), uintptr(len(key)), uintptr(unsafe.Pointer(&buf[0])), uintptr(pathLen+1))
	return syscall.UTF16ToString(buf)
}

// --- dwriteLocalizedStrings (vtable start: 3) ---

type dwriteLocalizedStrings struct{ dwriteObject }

func (s *dwriteLocalizedStrings) release() { s.call(vtbl_Release) }

func (s *dwriteLocalizedStrings) localeString(index uint32) string {
	var length uint32
	s.call(7, uintptr(index), uintptr(unsafe.Pointer(&length)))
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	s.call(8, uintptr(index), uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	return syscall.UTF16ToString(buf)
}

// --- helpers ---

func fontFaceToFont(face *dwriteFontFace, familyHint string, weight uint32, styleStr string) (*Font, error) {
	file := face.getFiles()
	if file == nil {
		return nil, fmt.Errorf("fc: failed to get font file")
	}
	defer file.release()

	key := file.referenceKey()
	if key == nil {
		return nil, fmt.Errorf("fc: failed to get reference key")
	}

	loader := file.getLoader()
	if loader == nil {
		return nil, fmt.Errorf("fc: failed to get local font file loader")
	}
	defer loader.release()

	path := loader.getFilePathFromKey(key)
	if path == "" {
		return nil, fmt.Errorf("fc: failed to get font file path")
	}

	familyName := familyHint
	if familyName == "" {
		familyName = "Unknown"
	}

	return &Font{
		File:   path,
		Family: familyName,
		Style:  styleStr,
		Index:  int(face.index()),
	}, nil
}

func parseStyle(style string) (weight, dwriteStyle, stretch uint32) {
	weight = 400  // DWRITE_FONT_WEIGHT_REGULAR
	stretch = 5   // DWRITE_FONT_STRETCH_NORMAL

	switch style {
	case "Thin", "Hairline":
		weight = 100
	case "ExtraLight", "UltraLight", "Extra Light", "Ultra Light":
		weight = 200
	case "Light":
		weight = 300
	case "SemiLight", "Semi Light":
		weight = 350
	case "Regular", "Normal", "":
		weight = 400
	case "Medium":
		weight = 500
	case "SemiBold", "Semi Bold", "DemiBold", "Demi Bold":
		weight = 600
	case "Bold":
		weight = 700
	case "ExtraBold", "Extra Bold", "UltraBold", "Ultra Bold":
		weight = 800
	case "Black", "Heavy":
		weight = 900
	}

	switch style {
	case "Italic":
		dwriteStyle = 2
	case "Oblique":
		dwriteStyle = 1
	default:
		dwriteStyle = 0
	}

	return
}
