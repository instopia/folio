// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package html

import (
	"testing"

	"golang.org/x/net/html"

	"github.com/carlos7ags/folio/layout"
)

// hintStyle returns the computed style of the first element with the given
// id, cascading from the root through every ancestor.
func hintStyle(t *testing.T, src, id string) computedStyle {
	t.Helper()
	c, doc, root := newGoldenConverter(t, src)
	var target *html.Node
	var find func(n *html.Node)
	find = func(n *html.Node) {
		if target != nil {
			return
		}
		if n.Type == html.ElementNode && getAttr(n, "id") == id {
			target = n
			return
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			find(ch)
		}
	}
	find(doc)
	if target == nil {
		t.Fatalf("no element with id %q", id)
	}
	var styleOf func(n *html.Node) computedStyle
	styleOf = func(n *html.Node) computedStyle {
		if n == nil || n.Type != html.ElementNode {
			return root
		}
		return c.computeElementStyle(n, styleOf(n.Parent))
	}
	return styleOf(target)
}

func TestPresentationalHintsTableCells(t *testing.T) {
	src := `<table id="t" width="600" bgcolor="#eeeeee" cellspacing="3" cellpadding="8"><tbody valign="bottom">
<tr id="r" bgcolor="ff6600" align="right"><td id="c" width="50%" height="40" align="center" valign="top" nowrap>x</td><td id="c2">y</td></tr>
</tbody></table>`

	tbl := hintStyle(t, src, "t")
	if tbl.Width == nil || tbl.Width.Value != 600 || tbl.Width.Unit != "px" {
		t.Errorf("table width = %+v, want 600px", tbl.Width)
	}
	if tbl.BackgroundColor == nil || *tbl.BackgroundColor != layout.Hex("#eeeeee") {
		t.Errorf("table background = %v, want #eeeeee", tbl.BackgroundColor)
	}
	if tbl.BorderSpacingH != 2.25 { // 3px
		t.Errorf("table border-spacing = %v, want 2.25pt", tbl.BorderSpacingH)
	}

	row := hintStyle(t, src, "r")
	if row.BackgroundColor == nil || *row.BackgroundColor != layout.Hex("#ff6600") {
		t.Errorf("row bgcolor without # = %v, want #ff6600", row.BackgroundColor)
	}

	cell := hintStyle(t, src, "c")
	if cell.Width == nil || cell.Width.Value != 50 || cell.Width.Unit != "%" {
		t.Errorf("cell width = %+v, want 50%%", cell.Width)
	}
	if cell.Height == nil || cell.Height.Value != 40 {
		t.Errorf("cell height = %+v, want 40px", cell.Height)
	}
	if cell.TextAlign != layout.AlignCenter {
		t.Errorf("cell align = %v, want center", cell.TextAlign)
	}
	if cell.VerticalAlign != "top" {
		t.Errorf("cell valign = %q, want top", cell.VerticalAlign)
	}
	if cell.WhiteSpace != "nowrap" {
		t.Errorf("cell nowrap = %q, want nowrap", cell.WhiteSpace)
	}
	if !cell.hasPadding() || cell.PaddingTopAt(500) != 6 { // 8px
		t.Errorf("cell padding = %v, want 6pt from cellpadding", cell.PaddingTopAt(500))
	}

	// A cell without its own attributes inherits valign from the row
	// group and align (text-align is inherited) from the row.
	c2 := hintStyle(t, src, "c2")
	if c2.VerticalAlign != "bottom" {
		t.Errorf("c2 valign = %q, want bottom from <tbody valign>", c2.VerticalAlign)
	}
	if c2.TextAlign != layout.AlignRight {
		t.Errorf("c2 align = %v, want right from <tr align>", c2.TextAlign)
	}
}

func TestPresentationalHintsLoseToCSS(t *testing.T) {
	src := `<style>#c { background-color: #000000; text-align: left }</style>
<table><tr><td id="c" bgcolor="#ff0000" align="center" width="100" style="width:200px">x</td></tr></table>`
	cell := hintStyle(t, src, "c")
	if cell.BackgroundColor == nil || *cell.BackgroundColor != layout.Hex("#000000") {
		t.Errorf("background = %v, want the stylesheet's #000000", cell.BackgroundColor)
	}
	if cell.TextAlign != layout.AlignLeft {
		t.Errorf("text-align = %v, want the stylesheet's left", cell.TextAlign)
	}
	if cell.Width == nil || cell.Width.Value != 200 {
		t.Errorf("width = %+v, want the inline 200px", cell.Width)
	}
}

func TestPresentationalHintsFontAndCenter(t *testing.T) {
	src := `<center id="ce"><p>t <font id="f" color="red" face="Georgia, serif" size="+2">big</font></p></center>`
	f := hintStyle(t, src, "f")
	if f.Color != layout.Hex("#ff0000") {
		t.Errorf("font color = %v, want red", f.Color)
	}
	if f.FontFamily != "georgia" {
		t.Errorf("font face = %q, want georgia", f.FontFamily)
	}
	if f.FontSize != 18 { // size=5 → x-large = 24px
		t.Errorf("font size = %v, want 18pt", f.FontSize)
	}
	if f.Display != "inline" {
		t.Errorf("font display = %q, want inline", f.Display)
	}
	if ce := hintStyle(t, src, "ce"); ce.TextAlign != layout.AlignCenter {
		t.Errorf("center text-align = %v, want center", ce.TextAlign)
	}
}

func TestPresentationalHintsTableCentering(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      bool
	}{
		{"align=center", `<table id="t" align="center" width="300"><tr><td>x</td></tr></table>`, true},
		{"inside td align=center", `<table><tr><td align="center"><table id="t" width="300"><tr><td>x</td></tr></table></td></tr></table>`, true},
		{"inside center", `<center><div><table id="t" width="300"><tr><td>x</td></tr></table></div></center>`, true},
		{"nearest align wins", `<div align="center"><div align="left"><table id="t"><tr><td>x</td></tr></table></div></div>`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := hintStyle(t, tc.src, "t")
			got := s.MarginLeftAuto && s.MarginRightAuto
			if got != tc.want {
				t.Errorf("centered = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseDimensionAndLegacyFontSize(t *testing.T) {
	dims := map[string]string{"600": "600px", " 50% ": "50%", "600px": "600px", "12.5": "12.5px", "0": "", "abc": "", "": ""}
	for in, want := range dims {
		got, _ := parseDimension(in)
		if got != want {
			t.Errorf("parseDimension(%q) = %q, want %q", in, got, want)
		}
	}
	sizes := map[string]string{"1": "x-small", "3": "medium", "7": "48px", "9": "48px", "+1": "large", "-2": "x-small", "-5": "x-small", "x": ""}
	for in, want := range sizes {
		got, _ := legacyFontSize(in)
		if got != want {
			t.Errorf("legacyFontSize(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestFontElementIsInline asserts <font> joins the surrounding text's
// paragraph instead of breaking out as a block of its own.
func TestFontElementIsInline(t *testing.T) {
	elems, err := Convert(`<div>Div <font color="red">font</font> text</div>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	var paras int
	var walk func(e layout.Element)
	walk = func(e layout.Element) {
		switch v := e.(type) {
		case *layout.Paragraph:
			paras++
		case *layout.Div:
			for _, ch := range v.Children() {
				walk(ch)
			}
		}
	}
	for _, e := range elems {
		walk(e)
	}
	if paras != 1 {
		t.Errorf("got %d paragraphs, want 1 (font must be inline)", paras)
	}
}
