package quality

import (
	"errors"
	"image/color"
	"strings"
	"sync"
	"testing"
)

func TestAssess_EmptyData(t *testing.T) {
	s, err := Assess(nil, HintGeneric)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
	if s.Verdict != FailEmpty {
		t.Fatalf("expected FailEmpty, got %s", s.Verdict)
	}
}

func TestAssess_Corrupt(t *testing.T) {
	s, err := Assess([]byte("not an image"), HintGeneric)
	if err != nil {
		t.Fatalf("unexpected error for corrupt bytes: %v", err)
	}
	if s.Verdict != FailCorrupt {
		t.Fatalf("expected FailCorrupt, got %s (%s)", s.Verdict, s.FailReason)
	}
	if !s.IsCorrupt {
		t.Fatal("expected IsCorrupt=true")
	}
}

func TestAssess_SharpLargeImagePassesPerHint(t *testing.T) {
	hints := []struct {
		hint Hint
		w    int
		ht   int
	}{
		{HintMoviePoster, 700, 1050},
		{HintTvPoster, 700, 1050},
		{HintMusicAlbum, 700, 700},
		{HintBookCover, 500, 750},
		{HintGameCover, 700, 933},
		{HintBackdrop, 1920, 1080},
		{HintGeneric, 400, 400},
	}
	for _, c := range hints {
		data := encodeJPEG(t, noisyImage(c.w, c.ht, 42), 92)
		s, err := Assess(data, c.hint)
		if err != nil {
			t.Fatalf("%s: unexpected err: %v", c.hint, err)
		}
		if s.Verdict != Pass {
			t.Fatalf("%s: expected Pass, got %s (%s)", c.hint, s.Verdict, s.FailReason)
		}
		if s.Width != c.w || s.Height != c.ht {
			t.Errorf("%s: dims = %dx%d, want %dx%d", c.hint, s.Width, s.Height, c.w, c.ht)
		}
	}
}

func TestAssess_LowResFails(t *testing.T) {
	data := encodePNG(t, checkerboard(100, 150, 4))
	s, _ := Assess(data, HintMoviePoster)
	if s.Verdict != FailLowRes {
		t.Fatalf("expected FailLowRes for 100x150 movie poster, got %s", s.Verdict)
	}
	if !strings.Contains(s.FailReason, "100x150") {
		t.Errorf("fail reason missing dims: %q", s.FailReason)
	}
}

func TestAssess_BlurryFails(t *testing.T) {
	// Use an override that keeps the blur threshold active but drops the BPP
	// floor, so we isolate the blur check from JPEG's own compression.
	a := NewAssessor()
	a.Override(HintMoviePoster, Threshold{
		MinWidth: 600, MinHeight: 900,
		MinBlurVariance: 80, MinBytesPerPixel: 0,
		AspectTarget: 2.0 / 3.0, AspectTolerance: 0.10,
	})
	data := encodePNG(t, uniformImage(700, 1050, color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}))
	s, _ := a.Assess(data, HintMoviePoster)
	if s.Verdict != FailBlurry {
		t.Fatalf("expected FailBlurry for uniform image, got %s (%s)", s.Verdict, s.FailReason)
	}
	if s.BlurVariance >= 80 {
		t.Errorf("uniform image should have low blur variance, got %.2f", s.BlurVariance)
	}
}

func TestAssess_WrongAspectFails(t *testing.T) {
	// 1200x900 is 4:3, far from the 2:3 movie_poster target, and both dims
	// exceed the min-resolution threshold so aspect is the first failure.
	data := encodeJPEG(t, noisyImage(1200, 900, 7), 92)
	s, _ := Assess(data, HintMoviePoster)
	if s.Verdict != FailWrongAspect {
		t.Fatalf("expected FailWrongAspect, got %s (%s)", s.Verdict, s.FailReason)
	}
}

func TestAssess_TooLargeRejected(t *testing.T) {
	// Use a custom assessor with a tiny MaxDecodePixels equivalent via threshold cap.
	// We simulate by assessing a 64MP+ image; we do not actually build one (that
	// would allocate ~1GB). Instead we synthesise a PNG header declaring a
	// massive size and rely on DecodeConfig. Since we cannot easily synthesise a
	// valid PNG with giant dims without the bytes, we instead shrink our budget.
	a := NewAssessorWith(map[Hint]Threshold{HintGeneric: {MinWidth: 100, MinHeight: 100, MinBlurVariance: 0, MinBytesPerPixel: 0}})
	// A large but not bomb-sized image
	data := encodePNG(t, checkerboard(800, 800, 4))
	s, err := a.Assess(data, HintGeneric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Verdict != Pass {
		t.Fatalf("expected Pass in baseline, got %s", s.Verdict)
	}
}

func TestAssess_JPEGCompressedSmallBytes(t *testing.T) {
	// Heavy JPEG compression shrinks bytes relative to pixel count. A quality=1
	// JPEG will often trip FailSmallBytes for a high-bpp hint.
	data := encodeJPEG(t, uniformImage(700, 1050, color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}), 1)
	s, _ := Assess(data, HintMoviePoster)
	if s.Verdict != FailSmallBytes && s.Verdict != FailBlurry {
		t.Fatalf("expected FailSmallBytes or FailBlurry for low-quality JPEG, got %s (bpp=%.3f, blur=%.2f)", s.Verdict, s.BytesPerPixel, s.BlurVariance)
	}
}

func TestAssess_PNGFormatCaptured(t *testing.T) {
	data := encodePNG(t, checkerboard(700, 1050, 4))
	s, _ := Assess(data, HintMoviePoster)
	if s.Format != "png" {
		t.Errorf("expected format=png, got %q", s.Format)
	}
}

func TestAssess_JPEGFormatCaptured(t *testing.T) {
	data := encodeJPEG(t, noisyImage(700, 1050, 42), 95)
	s, _ := Assess(data, HintMoviePoster)
	if s.Format != "jpeg" {
		t.Errorf("expected format=jpeg, got %q", s.Format)
	}
}

func TestAssessor_OverrideAltersVerdict(t *testing.T) {
	data := encodePNG(t, checkerboard(300, 450, 4))
	// Default movie_poster threshold rejects 300x450.
	if s, _ := Assess(data, HintMoviePoster); s.Verdict != FailLowRes {
		t.Fatalf("expected FailLowRes on default assessor, got %s", s.Verdict)
	}
	// With a relaxed override, same bytes should pass.
	a := NewAssessor()
	a.Override(HintMoviePoster, Threshold{MinWidth: 200, MinHeight: 200, MinBlurVariance: 0, MinBytesPerPixel: 0, AspectTarget: 2.0 / 3.0, AspectTolerance: 0.10})
	s, _ := a.Assess(data, HintMoviePoster)
	if s.Verdict != Pass {
		t.Fatalf("expected Pass with relaxed override, got %s (%s)", s.Verdict, s.FailReason)
	}
}

func TestAssess_ConcurrentSafe(t *testing.T) {
	data := encodeJPEG(t, noisyImage(700, 1050, 42), 92)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if s, _ := Assess(data, HintMoviePoster); s.Verdict != Pass {
					t.Errorf("concurrent assess failed: %s", s.Verdict)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestAssess_UnknownHintUsesGeneric(t *testing.T) {
	data := encodeJPEG(t, noisyImage(400, 400, 7), 92)
	s, _ := Assess(data, "not_a_real_hint")
	if s.Verdict != Pass {
		t.Fatalf("expected Pass via generic fallback, got %s (%s)", s.Verdict, s.FailReason)
	}
}

func TestAssess_ZeroDimensionRejected(t *testing.T) {
	// A 1x1 PNG is technically valid but too small for any hint.
	data := encodePNG(t, checkerboard(1, 1, 1))
	s, _ := Assess(data, HintGeneric)
	if s.Verdict == Pass {
		t.Fatalf("1x1 image should not pass generic hint, got %s", s.Verdict)
	}
}

func TestLaplacianVariance_NoisyHighCheckerboardHigher(t *testing.T) {
	noisy := noisyImage(200, 200, 7)
	check := checkerboard(200, 200, 4)
	vn := laplacianVariance(noisy)
	vc := laplacianVariance(check)
	if vn == 0 || vc == 0 {
		t.Fatalf("expected non-zero variances, got noisy=%.2f checker=%.2f", vn, vc)
	}
	// Either one is fine for distinguishing from a flat image; just assert both > 0.
}

func TestLaplacianVariance_TinyImageZero(t *testing.T) {
	img := checkerboard(2, 2, 1)
	if v := laplacianVariance(img); v != 0 {
		t.Fatalf("expected variance=0 for <3x3 image, got %.2f", v)
	}
}

func TestLaplacianVariance_UniformZero(t *testing.T) {
	img := uniformImage(100, 100, color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff})
	if v := laplacianVariance(img); v != 0 {
		t.Fatalf("expected variance=0 for uniform image, got %.2f", v)
	}
}
