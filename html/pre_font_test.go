// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package html

import (
	"testing"

	"github.com/carlos7ags/folio/layout"
)

// TestPreBlankLinesKeepHeight asserts a blank line inside <pre> still
// takes a line's height, and that <pre> text uses the default family.
func TestPreBlankLinesKeepHeight(t *testing.T) {
	elems, err := Convert("<pre style=\"font-family:inherit\">Hi,\n\nThere</pre>", &Options{DefaultFont: goFamily})
	if err != nil {
		t.Fatal(err)
	}
	div, ok := elems[0].(*layout.Div)
	if !ok {
		t.Fatalf("got %T, want *layout.Div", elems[0])
	}
	kids := div.Children()
	if len(kids) != 3 {
		t.Fatalf("got %d lines, want 3", len(kids))
	}
	area := layout.LayoutArea{Width: 400, Height: 1000}
	blank := kids[1].PlanLayout(area).Consumed
	first := kids[0].PlanLayout(area).Consumed
	if blank <= 0 || blank != first {
		t.Errorf("blank line height = %v, want %v", blank, first)
	}
	if emb := kids[0].(*layout.Paragraph).Runs()[0].Embedded; emb == nil || emb.Face().PostScriptName() != "GoRegular" {
		t.Errorf("pre text not set in the default family")
	}
}
