// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlos7ags/folio/document"
	"github.com/carlos7ags/folio/html"
)

// TestCIDKeyedCFFGlyphsRender pins that text set in a CID-keyed CFF
// (OpenType/CFF) font actually paints. The fixture is a 17-glyph
// fontTools subset of Noto Sans SC's OTF build whose charset keeps the
// source CIDs (GID 1 → CID 9544, …), as every CJK OTF in the wild
// does. The content stream addresses glyphs by GID through Identity-H,
// and for a /CIDFontType0 font a reader maps that CID to a glyph via
// the embedded CFF's charset — so unless the embedded CFF's charset is
// the identity, every glyph resolves to the wrong slot (here: an
// emptied one) and the page prints blank while text extraction, which
// only consults /ToUnicode, still looks right.
//
// The rasterising half needs Poppler's pdftoppm and skips without it;
// the structural assertions always run.
func TestCIDKeyedCFFGlyphsRender(t *testing.T) {
	fixture, err := filepath.Abs("../font/testdata/NotoSansSC-CIDSubset.otf")
	if err != nil {
		t.Fatal(err)
	}
	const text = "中华人民共和国保险经纪有限公司张伟"
	htmlStr := fmt.Sprintf(`<!DOCTYPE html>
<html><head><style>
@font-face { font-family: 'CJK'; src: url('%s'); }
body { font-family: 'CJK'; font-size: 28px; }
</style></head><body><p>%s</p></body></html>`, fixture, text)

	result, err := html.ConvertFull(htmlStr, &html.Options{AllowAbsolutePaths: true, StrictAssets: true})
	if err != nil {
		t.Fatalf("ConvertFull: %v", err)
	}
	doc := document.NewDocument(document.PageSizeA4)
	for _, e := range result.Elements {
		doc.Add(e)
	}
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	pdf := buf.Bytes()
	for _, want := range []string{"/FontFile3", "/CIDFontType0C", "/CIDFontType0"} {
		if !bytes.Contains(pdf, []byte(want)) {
			t.Fatalf("pdf lacks %s: fixture did not take the CFF embed path", want)
		}
	}
	if bytes.Contains(pdf, []byte("/CIDToGIDMap")) {
		t.Error("pdf carries /CIDToGIDMap, which is invalid for /CIDFontType0")
	}

	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "cff.pdf")
	if err := os.WriteFile(pdfPath, pdf, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := exec.LookPath("pdftotext"); err == nil {
		out, err := exec.Command("pdftotext", pdfPath, "-").Output()
		if err != nil {
			t.Fatalf("pdftotext: %v", err)
		}
		if got := strings.Join(strings.Fields(string(out)), ""); got != text {
			t.Errorf("pdftotext = %q, want %q", got, text)
		}
	}

	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("pdftoppm not installed; skipping the rasterised glyph check")
	}
	prefix := filepath.Join(dir, "page")
	if out, err := exec.Command("pdftoppm", "-gray", "-r", "50", "-singlefile", pdfPath, prefix).CombinedOutput(); err != nil {
		t.Fatalf("pdftoppm: %v\n%s", err, out)
	}
	dark, err := countDarkPixels(prefix + ".pgm")
	if err != nil {
		t.Fatal(err)
	}
	// 17 ideographs at 28px, 50 dpi paint a few hundred dark pixels;
	// the blank-glyph regression paints none.
	if dark < 100 {
		t.Errorf("rasterised page has %d dark pixels: the CFF glyphs printed blank", dark)
	}
}

// countDarkPixels reads a binary PGM (P5, 8-bit) and returns how many
// pixels are darker than mid-grey.
func countDarkPixels(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	var magic string
	var w, h, maxVal int
	if _, err := fmt.Fscan(r, &magic, &w, &h, &maxVal); err != nil {
		return 0, fmt.Errorf("pgm header: %w", err)
	}
	if magic != "P5" || maxVal > 255 {
		return 0, fmt.Errorf("pgm: unsupported %s maxval %d", magic, maxVal)
	}
	if _, err := r.ReadByte(); err != nil { // single whitespace after header
		return 0, err
	}
	px := make([]byte, w*h)
	if _, err := io.ReadFull(r, px); err != nil {
		return 0, fmt.Errorf("pgm pixels: %w", err)
	}
	n := 0
	for _, p := range px {
		if p < 128 {
			n++
		}
	}
	return n, nil
}
