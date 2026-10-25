package html

import (
	"testing"

	"github.com/carlos7ags/folio/layout"
)

// An <hr> drawn by its border paints in the border's colour, width and
// style, never the inherited text colour: a body { color } must not
// recolour a page's rules.
func TestHrBorderColourWinsOverInheritedTextColour(t *testing.T) {
	cases := []struct {
		name, html string
		color      layout.Color
		width      float64
		style      layout.BorderStyle
	}{
		{"border-top", `<div style="color:#1e293b"><hr style="border:0; border-top:2px solid #ff0000; height:0"/></div>`,
			layout.RGB(1, 0, 0), 1.5, layout.BorderSolid},
		{"var fallback", `<div style="color:#1e293b; --pb-border:#00ff00"><hr style="border:0; border-top:12px solid var(--pb-border, #e2e8f0)"/></div>`,
			layout.RGB(0, 1, 0), 9, layout.BorderSolid},
		{"dashed", `<div style="color:#1e293b"><hr style="border:0; border-top:1px dashed #0000ff"/></div>`,
			layout.RGB(0, 0, 1), 0.75, layout.BorderDashed},
	}
	for _, tc := range cases {
		elems, err := Convert(tc.html, nil)
		if err != nil {
			t.Fatal(err)
		}
		var hr *layout.LineSeparator
		for _, e := range elems {
			if ls, ok := e.(*layout.LineSeparator); ok {
				hr = ls
			}
		}
		if hr == nil {
			t.Fatalf("%s: no LineSeparator in %v", tc.name, elems)
		}
		if hr.Color() != tc.color {
			t.Errorf("%s: colour %+v, want %+v", tc.name, hr.Color(), tc.color)
		}
		if hr.Width() != tc.width {
			t.Errorf("%s: width %v, want %v", tc.name, hr.Width(), tc.width)
		}
		if hr.Style() != tc.style {
			t.Errorf("%s: style %v, want %v", tc.name, hr.Style(), tc.style)
		}
	}
}
