package html

import (
	"testing"

	"github.com/carlos7ags/folio/layout"
)

func TestTableCSSBorderPaintsTableBoxOnly(t *testing.T) {
	none := layout.CellBorders{}
	tbl, div := firstTable(t, `<table style="border:2px solid #ddd"><tr><td>a</td><td>b</td></tr></table>`)
	for i, c := range tbl.Rows()[0].Cells() {
		if c.Borders() != none {
			t.Errorf("cell %d borders = %+v, want none (CSS border is not inherited)", i, c.Borders())
		}
	}
	if div == nil || div.Borders().Top.Width != 1.5 || div.Borders().Left.Width != 1.5 {
		t.Fatalf("table box border missing: div=%v", div)
	}

	// The border attribute still borders every cell.
	tbl, _ = firstTable(t, `<table border="1"><tr><td>a</td><td>b</td></tr></table>`)
	for i, c := range tbl.Rows()[0].Cells() {
		if c.Borders().Top.Width == 0 {
			t.Errorf("border=1: cell %d has no border", i)
		}
	}

	// Collapsed table with bordered cells: no extra box around them.
	_, div = firstTable(t, `<table border="1" style="border-collapse:collapse;border:1px solid"><tr><td>a</td></tr></table>`)
	if div != nil && div.Borders() != none {
		t.Errorf("collapsed bordered-cell table got a second box border: %+v", div.Borders())
	}
}
