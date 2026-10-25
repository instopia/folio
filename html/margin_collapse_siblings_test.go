package html

import (
	"math"
	"testing"

	"github.com/carlos7ags/folio/layout"
)

type spaced interface {
	GetSpaceBefore() float64
	GetSpaceAfter() float64
}

// Adjacent block siblings' vertical margins collapse to the larger of the
// two (CSS 2.1 §8.3.1) for every block-level element type the converter
// produces — not just in theory: the gap is max(a, b), not a + b.
func TestAdjacentSiblingMarginsCollapseForEveryBlockType(t *testing.T) {
	blocks := map[string]string{
		"div":  `<div style="margin:%s 0">x</div>`,
		"p":    `<p style="margin:%s 0">x</p>`,
		"h2":   `<h2 style="margin:%s 0">x</h2>`,
		"flex": `<div style="display:flex; margin:%s 0"><div>a</div><div>b</div></div>`,
		"grid": `<div style="display:grid; grid-template-columns:1fr 1fr; margin:%s 0"><div>a</div><div>b</div></div>`,
		"hr":   `<hr style="border:0; border-top:1px solid #000; margin:%s 0"/>`,
	}
	sprintf := func(f, m string) string {
		out := ""
		for i := 0; i < len(f); i++ {
			if f[i] == '%' && i+1 < len(f) && f[i+1] == 's' {
				out += m
				i++
				continue
			}
			out += string(f[i])
		}
		return out
	}
	for prevName, prev := range blocks {
		for nextName, next := range blocks {
			// prev's 32px bottom margin wins over next's 16px top margin:
			// next keeps no space of its own.
			elems, err := Convert(sprintf(prev, "32px")+sprintf(next, "16px"), nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(elems) != 2 {
				t.Fatalf("%s+%s: %d elements, want 2", prevName, nextName, len(elems))
			}
			n, ok := elems[1].(spaced)
			if !ok {
				t.Fatalf("%s: %T has no space accessors", nextName, elems[1])
			}
			if got := n.GetSpaceBefore(); math.Abs(got) > 0.01 {
				t.Errorf("%s then %s: next SpaceBefore = %.2fpt, want 0 (margins collapse to the larger)", prevName, nextName, got)
			}
			// next's 32px top margin wins over prev's 16px: next keeps
			// only the difference (16px = 12pt).
			elems, err = Convert(sprintf(prev, "16px")+sprintf(next, "32px"), nil)
			if err != nil {
				t.Fatal(err)
			}
			if got := elems[1].(spaced).GetSpaceBefore(); math.Abs(got-12) > 0.01 {
				t.Errorf("%s then %s: next SpaceBefore = %.2fpt, want 12", prevName, nextName, got)
			}
		}
	}
}

// The collapsed gap shows in layout: two 20px-margin paragraphs sit 20px
// apart, so the pair is shorter than with summed margins.
func TestCollapsedMarginsShortenLayout(t *testing.T) {
	elems, err := Convert(`<p style="margin:20px 0">a</p><p style="margin:20px 0">b</p>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	var total float64
	for _, e := range elems {
		total += e.PlanLayout(layout.LayoutArea{Width: 400, Height: 1000}).Consumed
	}
	one, _ := Convert(`<p style="margin:20px 0">a</p>`, nil)
	single := one[0].PlanLayout(layout.LayoutArea{Width: 400, Height: 1000}).Consumed
	// Two paragraphs with collapsed margins: 2×single − 15pt (one shared gap).
	if want := 2*single - 15; math.Abs(total-want) > 0.5 {
		t.Errorf("two paragraphs consume %.2fpt, want %.2fpt (margins collapsed)", total, want)
	}
}
