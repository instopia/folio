package html_test

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/carlos7ags/folio/document"
	foliohtml "github.com/carlos7ags/folio/html"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
)

// Conversions sharing DefaultFont/FallbackFonts run concurrently
// (go test -race catches shared mutable state).
func TestSharedFontsConcurrentConversions(t *testing.T) {
	opts := &foliohtml.Options{DefaultFont: &foliohtml.FontFamily{Regular: goregular.TTF, Bold: gobold.TTF, Italic: goitalic.TTF}, FallbackFonts: [][]byte{goregular.TTF}, DisableSystemFonts: true}
	src := `<h1>Title</h1><p>Hello <b>bold</b> <i>italic</i> AVA To Wa — 漢</p><table><tr><th>H</th><td>d</td></tr></table>`
	var wg sync.WaitGroup
	sizes := make([]int, 8)
	for i := range sizes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			doc := document.NewDocument(document.PageSizeA4)
			if err := doc.AddHTMLWithContext(context.Background(), src, opts); err != nil {
				t.Error(err)
				return
			}
			var buf bytes.Buffer
			if _, err := doc.WriteTo(&buf); err != nil {
				t.Error(err)
				return
			}
			sizes[i] = buf.Len()
		}()
	}
	wg.Wait()
	for i, n := range sizes {
		if n == 0 || n != sizes[0] {
			t.Errorf("conversion %d wrote %d bytes, want %d (identical, non-empty)", i, n, sizes[0])
		}
	}
}
