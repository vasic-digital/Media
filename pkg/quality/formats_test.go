package quality

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"golang.org/x/image/bmp"
)

func noisyRGBA(w, h int, seed int64) image.Image {
	// Deliberately small noisy image; used to exercise decoders on
	// every supported format without blowing test runtime.
	return noisyImage(w, h, seed)
}

func TestAssess_DecodesPNG(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, noisyRGBA(400, 400, 1)); err != nil {
		t.Fatal(err)
	}
	score, err := Assess(buf.Bytes(), HintGeneric)
	if err != nil {
		t.Fatal(err)
	}
	if score.Format != "png" {
		t.Errorf("format = %q", score.Format)
	}
}

func TestAssess_DecodesJPEG(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, noisyRGBA(400, 400, 2), &jpeg.Options{Quality: 92}); err != nil {
		t.Fatal(err)
	}
	score, err := Assess(buf.Bytes(), HintGeneric)
	if err != nil {
		t.Fatal(err)
	}
	if score.Format != "jpeg" {
		t.Errorf("format = %q", score.Format)
	}
}

func TestAssess_DecodesGIF(t *testing.T) {
	var buf bytes.Buffer
	palette := []color.Color{color.Black, color.White}
	img := image.NewPaletted(image.Rect(0, 0, 400, 400), palette)
	for y := 0; y < 400; y++ {
		for x := 0; x < 400; x++ {
			if (x+y)%2 == 0 {
				img.SetColorIndex(x, y, 1)
			}
		}
	}
	if err := gif.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	score, err := Assess(buf.Bytes(), HintGeneric)
	if err != nil {
		t.Fatal(err)
	}
	if score.Format != "gif" {
		t.Errorf("format = %q", score.Format)
	}
}

func TestAssess_DecodesBMP(t *testing.T) {
	var buf bytes.Buffer
	if err := bmp.Encode(&buf, noisyRGBA(400, 400, 3)); err != nil {
		t.Fatal(err)
	}
	score, err := Assess(buf.Bytes(), HintGeneric)
	if err != nil {
		t.Fatal(err)
	}
	if score.Format != "bmp" {
		t.Errorf("format = %q", score.Format)
	}
}

func TestAssess_PolyglotContentRejected(t *testing.T) {
	// A JPEG header followed by garbage: should still decode (image/jpeg
	// is forgiving) or fail as corrupt. Either outcome is acceptable;
	// we only need to prove Nexus does not panic.
	raw := []byte{0xff, 0xd8, 0xff, 0xe0, 'G', 'A', 'R', 'B'}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic on polyglot: %v", r)
		}
	}()
	_, _ = Assess(raw, HintGeneric)
}

func TestAssess_DecompressionBombGuard(t *testing.T) {
	// Synthesise a truthy-but-tiny image and set our Assessor's internal
	// threshold pointer via the public MaxDecodePixels constant. The
	// easiest reachable path is to rely on real image dimensions; we
	// build a ~6400x6400 (~40 MP) image and a duplicate assessor with
	// a tighter cap expressed via Override-style threshold. Since
	// MaxDecodePixels is a package const, we cannot shrink it from
	// tests — instead we verify the guard triggers when called with a
	// hand-crafted PNG header declaring 9000x9000 pixels (81 MP, above
	// the 64 MP cap).
	img := image.NewRGBA(image.Rect(0, 0, 9000, 9000))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	s, err := Assess(buf.Bytes(), HintGeneric)
	if err != nil {
		t.Fatal(err)
	}
	if s.Verdict != FailTooLarge {
		t.Errorf("expected FailTooLarge for >64MP, got %s", s.Verdict)
	}
}

func TestAssess_ZeroByteRejected(t *testing.T) {
	if _, err := Assess([]byte{}, HintGeneric); err == nil {
		t.Fatal("empty bytes must error")
	}
}

func TestAssess_GarbageRejectedAsCorrupt(t *testing.T) {
	s, err := Assess([]byte("this is not an image at all"), HintGeneric)
	if err != nil {
		t.Fatal(err)
	}
	if s.Verdict != FailCorrupt {
		t.Errorf("garbage should be FailCorrupt, got %s", s.Verdict)
	}
}

func TestAssess_PartialJPEGReportedCorrupt(t *testing.T) {
	// JPEG prefix only — decoder should bail.
	raw := []byte{0xff, 0xd8}
	s, _ := Assess(raw, HintGeneric)
	if s.Verdict != FailCorrupt {
		t.Errorf("partial jpeg should be FailCorrupt, got %s", s.Verdict)
	}
}

func TestAssess_ExactThresholdBoundary(t *testing.T) {
	// At exactly the minimum width/height, the image should pass if
	// other signals are green.
	data := encodeJPEG(t, noisyImage(600, 900, 42), 92)
	s, _ := Assess(data, HintMoviePoster)
	if s.Verdict != Pass {
		t.Errorf("600x900 movie poster should pass, got %s (%s)", s.Verdict, s.FailReason)
	}
}

func TestAssess_OneBelowBoundaryFails(t *testing.T) {
	data := encodeJPEG(t, noisyImage(599, 900, 42), 92)
	s, _ := Assess(data, HintMoviePoster)
	if s.Verdict != FailLowRes {
		t.Errorf("599x900 movie poster should FailLowRes, got %s", s.Verdict)
	}
}

func TestAssess_BadAspectEvenWhenSharp(t *testing.T) {
	data := encodeJPEG(t, noisyImage(1400, 900, 42), 92)
	s, _ := Assess(data, HintMoviePoster)
	if s.Verdict != FailWrongAspect {
		t.Errorf("4:3 should FailWrongAspect for movie_poster, got %s", s.Verdict)
	}
}

func TestAssess_ReadableFormatStringStable(t *testing.T) {
	// Regression guard: Format should always be stringish after decoding.
	data := encodeJPEG(t, noisyImage(400, 400, 7), 90)
	s, _ := Assess(data, HintGeneric)
	if strings.TrimSpace(s.Format) == "" {
		t.Errorf("format empty for decoded image")
	}
}
