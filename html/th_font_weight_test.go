// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package html

import (
	"testing"

	"github.com/carlos7ags/folio/font"
	"github.com/carlos7ags/folio/layout"
)

// TestThBoldIsOverridable asserts <th> is bold by default but an author's
// font-weight on it wins, as with the UA stylesheet rule in browsers.
func TestThBoldIsOverridable(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      *font.Standard
	}{
		{"default", `<table><tr><th>H</th></tr></table>`, font.HelveticaBold},
		{"font-weight normal", `<style>th{font-weight:normal}</style><table><tr><th>H</th></tr></table>`, font.Helvetica},
	} {
		t.Run(tc.name, func(t *testing.T) {
			elems, err := Convert(tc.src, nil)
			if err != nil {
				t.Fatal(err)
			}
			tbl, ok := elems[0].(*layout.Table)
			if !ok {
				t.Fatalf("got %T, want *layout.Table", elems[0])
			}
			p, ok := tbl.Rows()[0].Cells()[0].Content().(*layout.Paragraph)
			if !ok {
				t.Fatalf("cell content is %T, want *layout.Paragraph", tbl.Rows()[0].Cells()[0].Content())
			}
			if got := p.Runs()[0].Font; got != tc.want {
				t.Errorf("font = %v, want %v", got.Name(), tc.want.Name())
			}
		})
	}
}
