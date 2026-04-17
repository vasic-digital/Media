package quality

import "testing"

func TestDefaultThresholds_CoversAllHints(t *testing.T) {
	d := DefaultThresholds()
	for _, h := range []Hint{
		HintMoviePoster, HintTvPoster, HintMusicAlbum, HintBookCover,
		HintGameCover, HintBackdrop, HintGeneric,
	} {
		if _, ok := d[h]; !ok {
			t.Errorf("default thresholds missing hint %q", h)
		}
	}
}

func TestLookup_FallsBackToGeneric(t *testing.T) {
	d := DefaultThresholds()
	got, ok := Lookup(d, "unrecognised_hint")
	if ok {
		t.Fatal("expected ok=false for unknown hint")
	}
	if got != d[HintGeneric] {
		t.Fatalf("expected fallback to HintGeneric, got %+v", got)
	}
}

func TestLookup_KnownHint(t *testing.T) {
	d := DefaultThresholds()
	got, ok := Lookup(d, HintMoviePoster)
	if !ok {
		t.Fatal("expected ok=true for known hint")
	}
	if got.MinWidth != 600 || got.MinHeight != 900 {
		t.Fatalf("movie_poster thresholds look wrong: %+v", got)
	}
}
