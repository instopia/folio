package html

import (
	"encoding/binary"
	"sync"

	"github.com/carlos7ags/folio/font"
)

// Parsed caller-supplied fonts (Options.DefaultFont, Options.FallbackFonts)
// are shared across conversions. A parsed font.Face is safe for
// concurrent use (its lazy GSUB/GPOS/kern caches are sync.Once-guarded),
// and parsing — above all GPOS, built on first measurement — dominated
// the cost of converting a short document. Each conversion still wraps
// the shared Face in its own font.EmbeddedFont (per-document glyph
// usage and subsetting), so nothing mutable is shared.
//
// Entries are keyed by the identity of the byte slice (its first byte's
// address and its length), so callers should pass the same slice each
// time (e.g. a package-level []byte or go:embed variable) and must not
// modify it afterwards — the same contract Options already states for
// these trusted bytes. Holding the pointer keeps the bytes alive, which
// also means an address can never be reused by different data while it
// is cached. The cache is bounded: when full it is cleared.
const maxSharedFonts = 64

type sharedFontKey struct {
	data *byte
	n    int
	lang string // only for TrueType Collections (face choice depends on it)
}

type sharedFontEntry struct {
	once sync.Once
	face font.Face
	err  error
}

var sharedFonts struct {
	mu sync.Mutex
	m  map[sharedFontKey]*sharedFontEntry
}

const ttcMagic = 0x74746366 // "ttcf"

// parseSharedFont returns the parsed face for data, parsing it at most
// once per distinct slice (and language, for collections).
func parseSharedFont(data []byte, lang string) (font.Face, error) {
	if len(data) == 0 {
		return font.ParseFontForLanguage(data, lang)
	}
	key := sharedFontKey{data: &data[0], n: len(data)}
	if len(data) >= 4 && binary.BigEndian.Uint32(data) == ttcMagic {
		key.lang = lang
	}
	sharedFonts.mu.Lock()
	if sharedFonts.m == nil || (len(sharedFonts.m) >= maxSharedFonts && sharedFonts.m[key] == nil) {
		sharedFonts.m = make(map[sharedFontKey]*sharedFontEntry)
	}
	e := sharedFonts.m[key]
	if e == nil {
		e = &sharedFontEntry{}
		sharedFonts.m[key] = e
	}
	sharedFonts.mu.Unlock()
	e.once.Do(func() { e.face, e.err = font.ParseFontForLanguage(data, lang) })
	return e.face, e.err
}
