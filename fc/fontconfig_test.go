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

	f, err := MatchFamily("DejaVu Sans", "")
	if err != nil {
		t.Fatalf("MatchFamily(DejaVu Sans): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if f.Family == "" {
		t.Error("expected non-empty Family")
	}
	t.Logf("DejaVu Sans -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
}

func TestMatchFamilyWithStyle(t *testing.T) {
	requireFontConfig(t)

	f, err := MatchFamily("DejaVu Sans", "Bold")
	if err != nil {
		t.Fatalf("MatchFamily(DejaVu Sans, Bold): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	t.Logf("DejaVu Sans Bold -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
}

func TestMatchFamilyNotFound(t *testing.T) {
	requireFontConfig(t)

	f, err := MatchFamily("NoSuchFontFamilyXYZ", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.File != "" {
		t.Errorf("expected empty File for non-existent family, got %q", f.File)
	}
}

func TestMatchFamilyLocalized(t *testing.T) {
	requireFontConfig(t)

	// DengXian is known as "等线" in Chinese. The font carries both names.
	// patternFamilyMatches checks all family names, so it should find the match.
	f, err := MatchFamily("DengXian", "")
	if err != nil {
		t.Fatalf("MatchFamily(DengXian): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File for DengXian (localized family)")
	}
	t.Logf("DengXian -> File=%s Family=%s Style=%q Index=%d", f.File, f.Family, f.Style, f.Index)
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
	f, idx, err := MatchCodepoint('A', nil)
	if err != nil {
		t.Fatalf("MatchCodepoint('A'): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if idx != 0 {
		t.Errorf("expected idx=0 (fallback for empty list = len(nil)), got %d", idx)
	}
	t.Logf("codepoint U+0041 (A) -> File=%s Family=%s Style=%q Index=%d familyIdx=%d", f.File, f.Family, f.Style, f.Index, idx)
}

func TestMatchCodepointWithFamilies(t *testing.T) {
	requireFontConfig(t)

	f, idx, err := MatchCodepoint('A', []Font{{Family: "DejaVu Sans"}})
	if err != nil {
		t.Fatalf("MatchCodepoint('A', DejaVu Sans): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if idx != 0 {
		t.Errorf("expected idx=0, got %d", idx)
	}
	t.Logf("codepoint U+0041 (A) with DejaVu Sans -> File=%s Family=%s Style=%q Index=%d familyIdx=%d", f.File, f.Family, f.Style, f.Index, idx)
}

func TestMatchCodepointWithStyle(t *testing.T) {
	requireFontConfig(t)

	f, idx, err := MatchCodepoint('A', []Font{{Family: "DejaVu Sans", Style: "Bold"}})
	if err != nil {
		t.Fatalf("MatchCodepoint('A', DejaVu Sans Bold): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if idx != 0 {
		t.Errorf("expected idx=0, got %d", idx)
	}
	t.Logf("codepoint U+0041 (A) with DejaVu Sans Bold -> File=%s Family=%s Style=%q Index=%d familyIdx=%d", f.File, f.Family, f.Style, f.Index, idx)
}

func TestMatchCodepointMultiFamily(t *testing.T) {
	requireFontConfig(t)

	f, idx, err := MatchCodepoint('A', []Font{
		{Family: "nosuchfamily"},
		{Family: "DejaVu Sans"},
		{Family: "DejaVu Serif"},
	})
	if err != nil {
		t.Fatalf("MatchCodepoint with multiple families: %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if idx != 1 {
		t.Errorf("expected idx=1 (DejaVu Sans), got %d", idx)
	}
	t.Logf("codepoint U+0041 (A) multi-family -> File=%s Family=%s Style=%q Index=%d familyIdx=%d", f.File, f.Family, f.Style, f.Index, idx)
}

func TestMatchCodepointFallback(t *testing.T) {
	requireFontConfig(t)

	// CJK codepoint — no families given, falls back to default.
	f, idx, err := MatchCodepoint('中', nil)
	if err != nil {
		t.Fatalf("MatchCodepoint('中'): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if idx != 0 {
		t.Errorf("expected idx=0 (fallback for empty list = len(nil)), got %d", idx)
	}
	t.Logf("codepoint U+4E2D (中) -> File=%s Family=%s Style=%q Index=%d familyIdx=%d", f.File, f.Family, f.Style, f.Index, idx)
}

func TestMatchCodepointMultiScript(t *testing.T) {
	requireFontConfig(t)

	families := []Font{
		{Family: "NoSuchFontXYZ"},       // 0: non-existent
		{Family: "Source Code Pro"},     // 1: Latin-only, no emoji
		{Family: "Noto Sans CJK TC"},    // 2: CJK + Latin
	}

	t.Run("Latin", func(t *testing.T) {
		f, idx, err := MatchCodepoint('A', families)
		if err != nil {
			t.Fatalf("MatchCodepoint('A'): %v", err)
		}
		if f.File == "" {
			t.Error("expected non-empty File")
		}
		if idx != 1 {
			t.Errorf("expected idx=1 (Source Code Pro), got %d", idx)
		}
		t.Logf("U+0041 (A) -> File=%s Family=%s Style=%q Index=%d familyIdx=%d", f.File, f.Family, f.Style, f.Index, idx)
	})

	t.Run("CJK", func(t *testing.T) {
		f, idx, err := MatchCodepoint('中', families)
		if err != nil {
			t.Fatalf("MatchCodepoint('中'): %v", err)
		}
		if f.File == "" {
			t.Error("expected non-empty File")
		}
		if idx != 2 {
			t.Errorf("expected idx=2 (Noto Sans CJK TC), got %d", idx)
		}
		t.Logf("U+4E2D (中) -> File=%s Family=%s Style=%q Index=%d familyIdx=%d", f.File, f.Family, f.Style, f.Index, idx)
	})

	t.Run("Emoji", func(t *testing.T) {
		f, idx, err := MatchCodepoint('😀', families)
	if err != nil {
		t.Fatalf("MatchCodepoint('😀'): %v", err)
	}
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if idx != 3 {
		t.Errorf("expected idx=3 (fallback, len=%d), got %d", len(families), idx)
	}
	t.Logf("U+1F600 (😀) -> File=%s Family=%s Style=%q Index=%d familyIdx=%d", f.File, f.Family, f.Style, f.Index, idx)
	})
}
