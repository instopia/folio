package html

import (
	"testing"

	"github.com/carlos7ags/folio/layout"
)

func TestPreHasNoForcedBox(t *testing.T) {
	elems, err := Convert(`<pre>x  y</pre>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	div, ok := elems[0].(*layout.Div)
	if !ok {
		t.Fatalf("pre is %T", elems[0])
	}
	if div.Background() != nil || div.Padding() != (layout.Padding{}) {
		t.Errorf("pre got bg %v padding %+v, want none", div.Background(), div.Padding())
	}
	elems, err = Convert(`<pre style="background:#eee;padding:4pt">x</pre>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	div = elems[0].(*layout.Div)
	if div.Background() == nil || div.Padding().Top != 4 {
		t.Errorf("author pre styles lost: bg %v padding %+v", div.Background(), div.Padding())
	}
}
