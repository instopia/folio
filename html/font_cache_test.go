package html

import (
	"bytes"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func TestParseSharedFontParsesOncePerSlice(t *testing.T) {
	a, err := parseSharedFont(goregular.TTF, "")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := parseSharedFont(goregular.TTF, "en")
	if a != b {
		t.Error("same slice parsed twice")
	}
	cp := bytes.Clone(goregular.TTF)
	c, _ := parseSharedFont(cp, "")
	if c == a {
		t.Error("a different slice must not share the cached face")
	}
}
