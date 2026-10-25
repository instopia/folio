package html

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"

	"github.com/carlos7ags/folio/layout"
)

func pngDataURI(t *testing.T, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		img.Set(x, 0, color.RGBA{255, 0, 0, 255})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// width: 50% on an <img> is half the containing block's width (CSS 2.1
// §10.2), resolved at layout time — not "auto", which filled the line.
func TestImagePercentWidthResolvesAgainstContainer(t *testing.T) {
	uri := pngDataURI(t, 200, 100)
	for _, tc := range []struct {
		style      string
		area, want float64
	}{
		{"width:50%", 400, 200},
		{"width:50%", 300, 150},
		{"width:25%; max-width:100%", 400, 100},
		{"width:100%", 400, 400},
	} {
		elems, err := Convert(`<img src="`+uri+`" style="`+tc.style+`"/>`, nil)
		if err != nil {
			t.Fatal(err)
		}
		ie, ok := elems[0].(*layout.ImageElement)
		if !ok {
			t.Fatalf("%s: %T, want *layout.ImageElement", tc.style, elems[0])
		}
		lines := ie.Layout(tc.area)
		if got := lines[0].Width; math.Abs(got-tc.want) > 0.01 {
			t.Errorf("%s in %vpt: width %.2f, want %.2f", tc.style, tc.area, got, tc.want)
		}
		if got := lines[0].Height; math.Abs(got-tc.want/2) > 0.01 {
			t.Errorf("%s in %vpt: height %.2f, want %.2f (aspect ratio kept)", tc.style, tc.area, got, tc.want/2)
		}
	}
}

// In a column flex container with align-items:center (a <figure> laid out
// as a centred column), a 50% image is half the container wide and
// centred — the percentage is of the container, applied once.
func TestPercentImageCentredInColumnFlex(t *testing.T) {
	uri := pngDataURI(t, 200, 100)
	elems, err := Convert(`<div style="display:flex; flex-direction:column; align-items:center"><img src="`+uri+`" style="width:50%"/></div>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	plan := elems[0].PlanLayout(layout.LayoutArea{Width: 400, Height: 1000})
	var found bool
	var walk func(bs []layout.PlacedBlock, x float64)
	walk = func(bs []layout.PlacedBlock, x float64) {
		for _, b := range bs {
			if b.Tag == "Figure" || (b.Width > 0 && len(b.Children) == 0 && b.Height > 50) {
				found = true
				if got := x + b.X; math.Abs(got-100) > 0.5 || math.Abs(b.Width-200) > 0.5 {
					t.Errorf("image at x=%.1f width %.1f, want x=100 width 200", got, b.Width)
				}
			}
			walk(b.Children, x+b.X)
		}
	}
	walk(plan.Blocks, 0)
	if !found {
		t.Fatal("image block not found")
	}
}
