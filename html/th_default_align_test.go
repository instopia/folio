package html

import (
	"testing"

	"github.com/carlos7ags/folio/layout"
)

// firstTable converts src and returns its (possibly Div-wrapped) table
// and the wrapper, if any.
func firstTable(t *testing.T, src string) (*layout.Table, *layout.Div) {
	t.Helper()
	elems, err := Convert(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range elems {
		switch v := e.(type) {
		case *layout.Table:
			return v, nil
		case *layout.Div:
			if tbl := findTable(v.Children()); tbl != nil {
				return tbl, v
			}
		}
	}
	t.Fatalf("no table in %q", src)
	return nil, nil
}

func cellAlign(t *testing.T, c *layout.Cell) layout.Align {
	t.Helper()
	p, ok := c.Content().(*layout.Paragraph)
	if !ok {
		t.Fatalf("cell content is %T, want *layout.Paragraph", c.Content())
	}
	return p.Align()
}

func TestThCenteredByDefault(t *testing.T) {
	cases := []struct {
		name, src string
		want      layout.Align
	}{
		{"default", `<table><tr><th>H</th></tr><tr><td>a long cell</td></tr></table>`, layout.AlignCenter},
		{"in thead", `<table><thead><tr><th>H</th></tr></thead><tr><td>a long cell</td></tr></table>`, layout.AlignCenter},
		{"css on th", `<style>th{text-align:left}</style><table><tr><th>H</th></tr></table>`, layout.AlignLeft},
		{"align attr on th", `<table><tr><th align="right">H</th></tr></table>`, layout.AlignRight},
		// HTML §15.3.8: only a th whose parent's text-align is initial centers.
		{"align attr on tr", `<table><tr align="left"><th>H</th></tr></table>`, layout.AlignLeft},
		{"css on ancestor", `<div style="text-align:right"><table><tr><th>H</th></tr></table></div>`, layout.AlignRight},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tbl, _ := firstTable(t, tc.src)
			if got := cellAlign(t, tbl.Rows()[0].Cells()[0]); got != tc.want {
				t.Errorf("th align = %v, want %v", got, tc.want)
			}
		})
	}
	// td stays start-aligned.
	tbl, _ := firstTable(t, `<table><tr><td>d</td></tr></table>`)
	if got := cellAlign(t, tbl.Rows()[0].Cells()[0]); got != layout.AlignLeft {
		t.Errorf("td align = %v, want left", got)
	}
}
