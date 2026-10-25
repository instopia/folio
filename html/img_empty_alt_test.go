package html

import (
	"strings"
	"testing"

	"github.com/carlos7ags/folio/layout"
)

// An image that fails to load with alt="" is decorative (HTML §4.8.4.4):
// it renders nothing — never a "[image: <url>]" placeholder that prints
// the URL into the document. Without an alt attribute the placeholder
// stays (it is the only sign something is missing).
func TestBrokenImageWithEmptyAltRendersNothing(t *testing.T) {
	for _, src := range []string{"https://cdn.example/brand.png?x=1", "missing.png", "data:image/png;base64,iVBORw0KGgo=", "missing.svg"} {
		elems, err := Convert(`<p>before</p><img src="`+src+`" alt=""/><p>after</p>`, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := textOf(elems); strings.Contains(got, "[image") {
			t.Errorf("%s: rendered %q, want no placeholder", src, got)
		}
		elems, err = Convert(`<img src="`+src+`"/>`, nil)
		if err != nil {
			t.Fatal(err)
		}
		got := textOf(elems)
		if !strings.Contains(got, "[image") {
			t.Errorf("%s without alt: rendered %q, want the placeholder", src, got)
		}
		if strings.Contains(got, "base64") {
			t.Errorf("%s without alt: placeholder %q prints the data: payload", src, got)
		}
	}
}

func textOf(elems []layout.Element) string {
	var b strings.Builder
	for _, e := range elems {
		if p, ok := e.(*layout.Paragraph); ok {
			for _, r := range p.Runs() {
				b.WriteString(r.Text)
			}
		}
	}
	return b.String()
}
