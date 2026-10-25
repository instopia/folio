package html

import "testing"

func TestNbspIsNotCollapsedOrBreakable(t *testing.T) {
	if got := collapseWhitespace("a  b  c"); got != "a  b c" {
		t.Errorf("collapseWhitespace = %q", got)
	}
	if got := collapseWhitespaceInline(" a b "); got != " a b " {
		t.Errorf("collapseWhitespaceInline = %q", got)
	}
}
