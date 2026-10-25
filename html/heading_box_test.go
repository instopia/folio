package html

import (
	"math"
	"testing"

	"github.com/carlos7ags/folio/layout"
)

func consumed(t *testing.T, src string) float64 {
	t.Helper()
	elems, err := Convert(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	total := 0.0
	for _, e := range elems {
		total += e.PlanLayout(layout.LayoutArea{Width: 400, Height: 1000}).Consumed
	}
	return total
}

// A heading's box is its CSS margins plus line-height × font-size, as for
// any block: no extra level-based space on top of the UA margins (which
// the converter already applies as CSS), and its line-height honoured.
func TestHeadingBoxIsMarginsPlusLineHeight(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want float64 // pt
	}{
		// 32px × 1.3 = 41.6px = 31.2pt
		{`<h1 style="margin:0; font-size:32px; line-height:1.3">b</h1>`, 31.2},
		// + 16px + 8px margins = 18pt
		{`<h1 style="margin:16px 0 8px 0; font-size:32px; line-height:1.3">b</h1>`, 49.2},
		// h2 at 24px, line-height 2 → 48px = 36pt
		{`<h2 style="margin:0; font-size:24px; line-height:2">b</h2>`, 36},
	} {
		if got := consumed(t, tc.src); math.Abs(got-tc.want) > 0.05 {
			t.Errorf("%s: %.2fpt, want %.2fpt", tc.src, got, tc.want)
		}
	}
	// With the UA margins (0.67em at 32px = 16.08pt each side) and the
	// default line-height, a bare <h1> is exactly the margins taller than
	// the same <h1> with margin:0.
	bare := consumed(t, `<h1>b</h1>`)
	flush := consumed(t, `<h1 style="margin:0">b</h1>`)
	if d := bare - flush; math.Abs(d-2*16.08) > 0.05 {
		t.Errorf("<h1> UA margins add %.2fpt, want %.2fpt", d, 2*16.08)
	}
}
