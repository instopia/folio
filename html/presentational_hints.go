// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package html

import (
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// presentationalHints maps legacy presentational attributes (bgcolor,
// width, align, valign, cellpadding, <font color face size>, ...) to CSS
// declarations, per the HTML Living Standard's rendering section (§15).
// Like browsers, computeElementStyle applies them as author-level
// declarations with zero specificity that precede every stylesheet and
// inline declaration, so any CSS for the same property wins.
//
// HTML email is still written largely with these attributes, so without
// them a typical newsletter loses its colours, column widths and
// alignment.
func presentationalHints(n *html.Node) [][2]string {
	if n.Type != html.ElementNode {
		return nil
	}
	switch n.DataAtom {
	case atom.Td, atom.Th, atom.Table, atom.Center:
		// May take hints from ancestors, or have a UA default.
	default:
		if len(n.Attr) == 0 {
			return nil
		}
	}
	var out [][2]string
	add := func(prop, val string) { out = append(out, [2]string{prop, val}) }

	switch n.DataAtom {
	case atom.Table:
		dimensionHint(n, "width", add)
		dimensionHint(n, "height", add)
		colorHint(n, "bgcolor", "background-color", add)
		if v, ok := attrPixels(n, "cellspacing"); ok {
			add("border-spacing", v)
		}
		switch strings.ToLower(strings.TrimSpace(getAttr(n, "align"))) {
		case "center":
			add("margin-left", "auto")
			add("margin-right", "auto")
		case "":
			// A table inside a centering container (<center>,
			// <td align=center>, <div align=center>) is centered by
			// browsers: the align hint behaves like -webkit-center.
			if centeredByAncestor(n) {
				add("margin-left", "auto")
				add("margin-right", "auto")
			}
		}
	case atom.Td, atom.Th:
		dimensionHint(n, "width", add)
		dimensionHint(n, "height", add)
		colorHint(n, "bgcolor", "background-color", add)
		alignHint(n, add)
		valignHint(n, add)
		if hasAttr(n, "nowrap") {
			add("white-space", "nowrap")
		}
		if tbl := cellTable(n); tbl != nil {
			if v, ok := attrPixels(tbl, "cellpadding"); ok {
				add("padding", v)
			}
		}
	case atom.Tr, atom.Thead, atom.Tbody, atom.Tfoot:
		dimensionHint(n, "height", add)
		colorHint(n, "bgcolor", "background-color", add)
		alignHint(n, add)
	case atom.Div, atom.P, atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Caption:
		alignHint(n, add)
	case atom.Center:
		// UA stylesheet: center { display: block; text-align: center }.
		add("text-align", "center")
	case atom.Body:
		colorHint(n, "bgcolor", "background-color", add)
		colorHint(n, "text", "color", add)
	case atom.Font:
		colorHint(n, "color", "color", add)
		if face := strings.TrimSpace(getAttr(n, "face")); face != "" {
			add("font-family", face)
		}
		if size, ok := legacyFontSize(getAttr(n, "size")); ok {
			add("font-size", size)
		}
	}
	return out
}

// dimensionHint maps a width/height attribute ("600", "600px", "50%") to
// the same-named CSS property. Zero and unparsable values are ignored,
// per the HTML "maps to the dimension property (ignoring zero)" rule.
func dimensionHint(n *html.Node, name string, add func(string, string)) {
	v, ok := parseDimension(getAttr(n, name))
	if ok {
		add(name, v)
	}
}

// parseDimension implements the HTML "rules for parsing dimension
// values": leading digits (with an optional fraction) are the value, a
// following "%" makes it a percentage, anything else is ignored. Returns
// a CSS length ("600px" / "50%").
func parseDimension(s string) (string, bool) {
	s = strings.TrimSpace(s)
	end := 0
	for end < len(s) && (s[end] >= '0' && s[end] <= '9' || s[end] == '.') {
		end++
	}
	if end == 0 {
		return "", false
	}
	v, err := strconv.ParseFloat(s[:end], 64)
	if err != nil || v <= 0 {
		return "", false
	}
	num := strconv.FormatFloat(v, 'f', -1, 64)
	if end < len(s) && s[end] == '%' {
		return num + "%", true
	}
	return num + "px", true
}

// attrPixels parses a non-negative integer attribute (cellpadding,
// cellspacing) as a CSS pixel length. Zero is kept: cellspacing="0" is
// meaningful.
func attrPixels(n *html.Node, name string) (string, bool) {
	s := strings.TrimSpace(getAttr(n, name))
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return "", false
	}
	v, err := strconv.Atoi(s[:end])
	if err != nil {
		return "", false
	}
	return strconv.Itoa(v) + "px", true
}

// colorHint maps a legacy colour attribute to prop. Besides CSS colour
// syntax it accepts bare hex digits ("ff6600"), which HTML's legacy
// colour parser treats as "#ff6600".
func colorHint(n *html.Node, attr, prop string, add func(string, string)) {
	v := strings.TrimSpace(getAttr(n, attr))
	if v == "" {
		return
	}
	if _, ok := parseColor(v); ok {
		add(prop, v)
		return
	}
	if len(v) == 6 || len(v) == 3 {
		if _, err := strconv.ParseUint(v, 16, 32); err == nil {
			add(prop, "#"+v)
		}
	}
}

// alignHint maps align=left|right|center|justify to text-align.
func alignHint(n *html.Node, add func(string, string)) {
	switch v := strings.ToLower(strings.TrimSpace(getAttr(n, "align"))); v {
	case "left", "right", "center", "justify":
		add("text-align", v)
	case "middle", "absmiddle":
		add("text-align", "center")
	}
}

// valignHint maps valign=top|middle|bottom|baseline on a cell to
// vertical-align. A cell without its own valign takes its row's, then
// its row group's: browsers' UA stylesheet makes cells inherit
// vertical-align from the row (td, th { vertical-align: inherit }).
func valignHint(n *html.Node, add func(string, string)) {
	for el := n; el != nil && el.Type == html.ElementNode; el = el.Parent {
		switch v := strings.ToLower(strings.TrimSpace(getAttr(el, "valign"))); v {
		case "top", "middle", "bottom", "baseline":
			add("vertical-align", v)
			return
		case "center":
			add("vertical-align", "middle")
			return
		}
		// Only climb cell → row → row group.
		switch el.DataAtom {
		case atom.Td, atom.Th:
			if el.Parent == nil || el.Parent.DataAtom != atom.Tr {
				return
			}
		case atom.Tr:
			if p := el.Parent; p == nil || (p.DataAtom != atom.Tbody && p.DataAtom != atom.Thead && p.DataAtom != atom.Tfoot) {
				return
			}
		default:
			return
		}
	}
}

// cellTable returns the <table> a cell belongs to (cell → tr →
// [thead|tbody|tfoot] → table), or nil.
func cellTable(n *html.Node) *html.Node {
	p := n.Parent
	if p == nil || p.DataAtom != atom.Tr {
		return nil
	}
	p = p.Parent
	if p != nil && (p.DataAtom == atom.Tbody || p.DataAtom == atom.Thead || p.DataAtom == atom.Tfoot) {
		p = p.Parent
	}
	if p == nil || p.DataAtom != atom.Table {
		return nil
	}
	return p
}

// centeredByAncestor reports whether the nearest ancestor carrying a
// horizontal alignment — an align attribute, or <center> — centers.
func centeredByAncestor(n *html.Node) bool {
	for el := n.Parent; el != nil; el = el.Parent {
		if el.Type != html.ElementNode {
			continue
		}
		if el.DataAtom == atom.Center {
			return true
		}
		if v := strings.ToLower(strings.TrimSpace(getAttr(el, "align"))); v != "" {
			return v == "center" || v == "middle"
		}
		if el.DataAtom == atom.Table {
			return false
		}
	}
	return false
}

// legacyFontSize maps <font size> (1–7, or +N/-N relative to 3) to a CSS
// font size, per the HTML "rules for parsing a legacy font size".
func legacyFontSize(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	mode := byte(0)
	if s[0] == '+' || s[0] == '-' {
		mode = s[0]
		s = s[1:]
	}
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return "", false
	}
	v, err := strconv.Atoi(s[:end])
	if err != nil {
		return "", false
	}
	switch mode {
	case '+':
		v = 3 + v
	case '-':
		v = 3 - v
	}
	v = min(max(v, 1), 7)
	return [...]string{"", "x-small", "small", "medium", "large", "x-large", "xx-large", "48px"}[v], true
}
