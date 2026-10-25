package html

import (
	"encoding/base64"
	"testing"

	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gosmallcaps"
)

// A font-family list falls back family by family, as in a browser: with
// no @font-face for Georgia, "Georgia, 'Times New Roman', serif" is a
// serif list (standard Times), not Helvetica because its first name is
// unknown.
func TestFontFamilyListFallsBackToItsGenericClass(t *testing.T) {
	src := `<p><span style="font-family:Georgia,'Times New Roman',serif">serif</span> ` +
		`<span style="font-family:Rockwell,'Courier New',monospace">mono</span> ` +
		`<span style="font-family:'Segoe UI',Candara,sans-serif">sans</span> ` +
		`<span style="font-family:Unknown">unknown</span></p>`
	got := wordFonts(t, src, nil)
	want := map[string]string{"serif": "Times-Roman", "mono": "Courier", "sans": "Helvetica", "unknown": "Helvetica"}
	for word, name := range want {
		if got[word] != name {
			t.Errorf("%q set in %q, want %q (all: %v)", word, got[word], name, got)
		}
	}
}

// Options.FontFamilies supplies in-memory families by name and for the
// generic classes; DefaultFont still covers everything else, and an
// @font-face rule for a name wins over the option.
func TestOptionsFontFamiliesByNameAndGenericClass(t *testing.T) {
	opts := &Options{
		DefaultFont:        goFamily,
		DisableSystemFonts: true,
		FontFamilies: map[string]*FontFamily{
			"serif": {Regular: gomono.TTF},
			"Brand": {Regular: gosmallcaps.TTF},
		},
	}
	src := `<p><span style="font-family:Georgia,Cambria,'Times New Roman',serif">georgia</span> ` +
		`<span style="font-family:Missing,brand,sans-serif">brand</span> ` +
		`<span style="font-family:Inter,Helvetica,sans-serif">sans</span> ` +
		`<b style="font-family:serif">heavy</b></p>`
	got := wordFonts(t, src, opts)
	want := map[string]string{
		"georgia": "GoMono",      // serif class → FontFamilies["serif"]
		"brand":   "GoSmallcaps", // second family in the list, by name
		"sans":    "GoRegular",   // no entry → DefaultFont
		"heavy":   "GoMono",      // nearest variant of the serif family
	}
	for word, name := range want {
		if got[word] != name {
			t.Errorf("%q set in %q, want %q (all: %v)", word, got[word], name, got)
		}
	}

	// An @font-face rule for the same name is matched first.
	face := `<style>@font-face { font-family: Brand; src: url(data:font/ttf;base64,` + b64(gomedium.TTF) + `) }</style>`
	got = wordFonts(t, face+`<p style="font-family:Brand">word</p>`, opts)
	if got["word"] != "GoMedium" {
		t.Errorf("@font-face Brand: set in %q, want GoMedium", got["word"])
	}
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
