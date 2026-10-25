package html

import (
	"testing"

	"github.com/carlos7ags/folio/layout"
)

// Each in-flow child of a grid container is ONE grid item (CSS Grid §6.1):
// a column <div> holding several blocks keeps them together in its cell,
// and an empty column still occupies its cell.
func TestGridItemPerChildBox(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		items     int
	}{
		{"multi-block column", `<div><p>a1</p><p>a2</p><table><tr><td>t</td></tr></table></div> <div><p>b1</p></div>`, 2},
		{"empty first column", `<div></div> <div><p>b1</p></div>`, 2},
		{"empty columns", `<div></div><div></div><div></div>`, 3},
		{"text item", `text <div>b</div>`, 2},
		{"not in flow", `<div style="display:none">x</div><style>p{}</style><script>1</script><div>b</div>`, 1},
	} {
		elems, err := Convert(`<div style="display:grid; grid-template-columns:1fr 1fr 1fr; gap:24px">`+tc.src+`</div>`, nil)
		if err != nil {
			t.Fatal(err)
		}
		g, ok := elems[0].(*layout.Grid)
		if !ok {
			t.Fatalf("%s: %T, want *layout.Grid", tc.name, elems[0])
		}
		if got := len(g.Children()); got != tc.items {
			t.Errorf("%s: %d grid items, want %d", tc.name, got, tc.items)
		}
	}
}

// The second column's content lays out in the second column even when the
// first is empty.
func TestGridEmptyFirstColumnKeepsSecondInPlace(t *testing.T) {
	elems, err := Convert(`<div style="display:grid; grid-template-columns:1fr 1fr"><div></div><div><p>right</p></div></div>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	plan := elems[0].PlanLayout(layout.LayoutArea{Width: 400, Height: 1000})
	var maxX float64
	var walk func(bs []layout.PlacedBlock, x float64)
	walk = func(bs []layout.PlacedBlock, x float64) {
		for _, b := range bs {
			if b.Tag == "P" && x+b.X > maxX {
				maxX = x + b.X
			}
			walk(b.Children, x+b.X)
		}
	}
	walk(plan.Blocks, 0)
	if maxX < 199 {
		t.Errorf("paragraph at x=%.1f, want the second column (x≥200)", maxX)
	}
}
