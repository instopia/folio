// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package layout

import (
	"testing"

	"github.com/carlos7ags/folio/font"
)

// TestLeadingPunctKeepsFollowingSpace covers a run that starts with
// punctuation abutting the previous run, then a space: the punctuation is
// absorbed into the previous word, and the space after it must survive.
// "<a>policy</a>? It's" used to render "policy?It's" because the glue step
// zeroed the previous word's SpaceAfter and the merge never restored it.
func TestLeadingPunctKeepsFollowingSpace(t *testing.T) {
	cases := []struct {
		name string
		runs []TextRun
		want []string // word text, with "_" appended when a space follows
	}{
		{
			name: "punct then space",
			runs: []TextRun{
				NewRun("for our policy", font.Helvetica, 10).WithLinkURI("https://example.com"),
				NewRun("? It's needed", font.Helvetica, 10),
			},
			want: []string{"for_", "our_", "policy?_", "It's_", "needed"},
		},
		{
			name: "punct only run, next run spaced",
			runs: []TextRun{
				NewRun("policy", font.Helvetica, 10).WithLinkURI("https://example.com"),
				NewRun(".", font.Helvetica, 10),
				NewRun(" Next", font.Helvetica, 10).WithLinkURI("https://example.org"),
			},
			want: []string{"policy._", "Next"},
		},
		{
			name: "punct only run, next run abutting",
			runs: []TextRun{
				NewRun("policy", font.Helvetica, 10).WithLinkURI("https://example.com"),
				NewRun(")", font.Helvetica, 10),
				NewRun("s", font.Helvetica, 10).WithLinkURI("https://example.org"),
			},
			want: []string{"policy)", "s"},
		},
		{
			name: "punct then word, no space",
			runs: []TextRun{
				NewRun("policy", font.Helvetica, 10).WithLinkURI("https://example.com"),
				NewRun("'s cover", font.Helvetica, 10),
			},
			want: []string{"policy'", "s_", "cover"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewStyledParagraph(tc.runs...)
			words, _ := p.measureWords(500)
			var got []string
			for i, w := range words {
				s := w.Text
				if w.SpaceAfter > 0 && i < len(words)-1 {
					s += "_"
				}
				got = append(got, s)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("words = %q, want %q", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("words = %q, want %q", got, tc.want)
				}
			}
		})
	}
}
