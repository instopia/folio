// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package html

import (
	"os"
	"testing"
	"testing/fstest"

	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/carlos7ags/folio/layout"
)

// goFamily is a complete in-memory family for the DefaultFont tests.
var goFamily = &FontFamily{
	Regular:    goregular.TTF,
	Bold:       gobold.TTF,
	Italic:     goitalic.TTF,
	BoldItalic: gobolditalic.TTF,
}

// wordFonts lays out the first paragraph of src and returns, for each
// word, the PostScript name of the embedded font or the standard font's
// name.
func wordFonts(t *testing.T, src string, opts *Options) map[string]string {
	t.Helper()
	elems, err := Convert(src, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(elems) == 0 {
		t.Fatal("no elements")
	}
	p, ok := elems[0].(*layout.Paragraph)
	if !ok {
		t.Fatalf("expected *layout.Paragraph, got %T", elems[0])
	}
	out := map[string]string{}
	for _, line := range p.Layout(500) {
		for _, w := range line.Words {
			switch {
			case w.Embedded != nil:
				out[w.Text] = w.Embedded.Face().PostScriptName()
			case w.Font != nil:
				out[w.Text] = w.Font.Name()
			}
		}
	}
	return out
}

func TestDefaultFontReplacesStandardFonts(t *testing.T) {
	src := `<p style="font-family:'Liberation Sans',Arial,sans-serif">plain <b>bold</b> <i>italic</i> <b><i>both</i></b> ` +
		`<span style="font-weight:600">semi</span> <span style="font-family:Georgia,serif">serif</span> ` +
		`<code>mono</code></p>`
	got := wordFonts(t, src, &Options{DefaultFont: goFamily, DisableSystemFonts: true})
	want := map[string]string{
		"plain":  "GoRegular",
		"bold":   "Go-Bold",
		"italic": "Go-Italic",
		"both":   "Go-BoldItalic",
		"semi":   "Go-Bold",   // 600 → nearest heavier face (CSS Fonts §5.2)
		"serif":  "GoRegular", // serif families use the default family too
		"mono":   "Courier",   // monospace keeps standard Courier
	}
	for word, ps := range want {
		if got[word] != ps {
			t.Errorf("%q set in %q, want %q (all: %v)", word, got[word], ps, got)
		}
	}
}

func TestDefaultFontMissingVariantsUseNearest(t *testing.T) {
	src := `<p>plain <b>bold</b> <i>italic</i> <b><i>both</i></b></p>`
	got := wordFonts(t, src, &Options{DefaultFont: &FontFamily{Regular: goregular.TTF, Bold: gobold.TTF}})
	want := map[string]string{"plain": "GoRegular", "bold": "Go-Bold", "italic": "GoRegular", "both": "Go-Bold"}
	for word, ps := range want {
		if got[word] != ps {
			t.Errorf("%q set in %q, want %q", word, got[word], ps)
		}
	}
}

func TestDefaultFontLosesToFontFace(t *testing.T) {
	src := `<style>@font-face{font-family:Brand;src:url(brand.ttf)}</style>` +
		`<p><span style="font-family:Brand">brand</span> other</p>`
	opts := &Options{
		DefaultFont: goFamily,
		BaseFS:      fstest.MapFS{"brand.ttf": &fstest.MapFile{Data: gobolditalic.TTF}},
	}
	got := wordFonts(t, src, opts)
	if got["brand"] != "Go-BoldItalic" || got["other"] != "GoRegular" {
		t.Errorf("got %v, want brand=Go-BoldItalic other=GoRegular", got)
	}
}

func TestFallbackFontsSplitByCoverage(t *testing.T) {
	cjk, err := os.ReadFile("../font/testdata/synthetic_cjk.ttf")
	if err != nil {
		t.Fatal(err)
	}
	src := `<p>Hello 中国 world</p>`
	opts := &Options{DefaultFont: goFamily, FallbackFonts: [][]byte{cjk}, DisableSystemFonts: true}
	got := wordFonts(t, src, opts)
	if got["Hello"] != "GoRegular" || got["world"] != "GoRegular" {
		t.Errorf("Latin words: got %v, want GoRegular", got)
	}
	found := false
	for word, ps := range got {
		for _, r := range word {
			if r == '中' {
				found = true
				if ps == "GoRegular" {
					t.Errorf("CJK word %q set in the default font; want the fallback", word)
				}
			}
		}
	}
	if !found {
		t.Errorf("no CJK word in output: %v", got)
	}

	// The same chain applies under the standard fonts.
	got = wordFonts(t, src, &Options{FallbackFonts: [][]byte{cjk}, DisableSystemFonts: true})
	if got["Hello"] != "Helvetica" {
		t.Errorf("Hello set in %q, want Helvetica", got["Hello"])
	}
}

func TestDisableSystemFontsSkipsSearch(t *testing.T) {
	c := &converter{opts: Options{DisableSystemFonts: true}}
	if fb := c.getFallbackFont(); fb != nil {
		t.Errorf("getFallbackFont = %s with DisableSystemFonts, want nil", fb.Face().PostScriptName())
	}
}

func TestClusterContinuationStaysInFont(t *testing.T) {
	for _, r := range []rune{0x200D, 0xFE0F, 0x0301, 0x1F3FD} {
		if !clusterContinuation(r) {
			t.Errorf("clusterContinuation(%U) = false, want true", r)
		}
	}
	for _, r := range []rune{'a', '中', 0x1F600} {
		if clusterContinuation(r) {
			t.Errorf("clusterContinuation(%U) = true, want false", r)
		}
	}
}
