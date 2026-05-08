package util

import "unicode/utf8"

// IsCharWide tells if the first rune in the text falls under
// East Asian Width class Wide (W) or Fullwidth (F).
func IsCharWide(text string) bool {
	// get the first rune
	ucs, _ := utf8.DecodeRuneInString(text)
	if ucs == utf8.RuneError {
		return false
	}

	// this part is copied from Markus Kuhn's wcwidth function.
	// https://www.cl.cam.ac.uk/~mgk25/ucs/wcwidth.c
	//
	// in case you want to make do the right way, here's these:
	// https://www.unicode.org/reports/tr11/
	// https://www.unicode.org/Public/UCD/latest/ucd/EastAsianWidth.txt
	return (ucs >= 0x1100 &&
		(ucs <= 0x115f || /* Hangul Jamo init. consonants */
			ucs == 0x2329 || ucs == 0x232a ||
			(ucs >= 0x2e80 && ucs <= 0xa4cf && ucs != 0x303f) || /* CJK ... Yi */
			(ucs >= 0xac00 && ucs <= 0xd7a3) || /* Hangul Syllables */
			(ucs >= 0xf900 && ucs <= 0xfaff) || /* CJK Compatibility Ideographs */
			(ucs >= 0xfe10 && ucs <= 0xfe19) || /* Vertical forms */
			(ucs >= 0xfe30 && ucs <= 0xfe6f) || /* CJK Compatibility Forms */
			(ucs >= 0xff00 && ucs <= 0xff60) || /* Fullwidth Forms */
			(ucs >= 0xffe0 && ucs <= 0xffe6) ||
			(ucs >= 0x20000 && ucs <= 0x2fffd) ||
			(ucs >= 0x30000 && ucs <= 0x3fffd)))
}
