package layout

import (
	"reflect"
	"testing"
)

func TestSplitWordsKeepsNoBreakSpaces(t *testing.T) {
	got := splitWords("$5,000 each and every claim　x")
	want := []string{"$5,000 each", "and every claim", "x"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitWords = %q, want %q", got, want)
	}
}
