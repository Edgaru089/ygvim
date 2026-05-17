package fc

import (
	"os"
	"testing"
)

func requireFontConfig(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/etc/fonts/fonts.conf"); os.IsNotExist(err) {
		t.Skip("fontconfig configuration not found, skipping test")
	}
}

func TestMatchFamily(t *testing.T) {
	requireFontConfig(t)

	f, err := MatchFamily("sans-serif", "")
	if err != nil {
		t.Fatalf("MatchFamily(sans-serif): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if f.Family == "" {
		t.Error("expected non-empty Family")
	}
	t.Logf("sans-serif -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
}

func TestMatchFamilyWithStyle(t *testing.T) {
	requireFontConfig(t)

	f, err := MatchFamily("serif", "Bold")
	if err != nil {
		t.Fatalf("MatchFamily(serif, Bold): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	t.Logf("serif Bold -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
}

func TestMatchFamilyEmptyName(t *testing.T) {
	_, err := MatchFamily("", "")
	if err == nil {
		t.Error("expected error for empty family name")
	}
}

func TestMatchCodepoint(t *testing.T) {
	requireFontConfig(t)

	// U+0041 is 'A' — every font should have it.
	f, err := MatchCodepoint('A')
	if err != nil {
		t.Fatalf("MatchCodepoint('A'): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	t.Logf("codepoint U+0041 (A) -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
}

func TestMatchCodepointWithFamilies(t *testing.T) {
	requireFontConfig(t)

	// U+0041 ('A') with sans-serif preferred.
	f, err := MatchCodepoint('A', "sans-serif")
	if err != nil {
		t.Fatalf("MatchCodepoint('A', sans-serif): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	t.Logf("codepoint U+0041 (A) with sans-serif -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
}

func TestMatchCodepointMultiFamily(t *testing.T) {
	requireFontConfig(t)

	// Try families in order. U+0041 is common.
	f, err := MatchCodepoint('A', "nosuchfamily", "sans-serif", "serif")
	if err != nil {
		t.Fatalf("MatchCodepoint with multiple families: %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	t.Logf("codepoint U+0041 (A) multi-family -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
}

func TestMatchCodepointFallback(t *testing.T) {
	requireFontConfig(t)

	// CJK codepoint — may fall back to a CJK font.
	f, err := MatchCodepoint('中')
	if err != nil {
		t.Fatalf("MatchCodepoint('中'): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	t.Logf("codepoint U+4E2D (中) -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
}
