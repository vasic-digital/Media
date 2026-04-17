package quality

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"math"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

// ErrEmptyData is returned when Assess is called with zero-length input.
var ErrEmptyData = errors.New("quality: empty image data")

// Assessor assesses image bytes against a configurable set of Hint thresholds.
// The zero value is not usable; construct one with NewAssessor.
type Assessor struct {
	thresholds map[Hint]Threshold
}

// NewAssessor returns an Assessor initialised with DefaultThresholds.
// Callers may mutate the returned map via Override before using it.
func NewAssessor() *Assessor {
	return &Assessor{thresholds: DefaultThresholds()}
}

// NewAssessorWith returns an Assessor using the supplied threshold map. It
// copies the map so later edits by the caller do not race against Assess.
func NewAssessorWith(t map[Hint]Threshold) *Assessor {
	cp := make(map[Hint]Threshold, len(t))
	for k, v := range t {
		cp[k] = v
	}
	return &Assessor{thresholds: cp}
}

// Override sets or replaces the Threshold for a single Hint.
func (a *Assessor) Override(h Hint, t Threshold) { a.thresholds[h] = t }

// Assess decodes data and scores it against the thresholds for h. A nil error
// with a non-Pass Verdict means the image decoded but failed the gate. A
// non-nil error is only returned for truly undecidable inputs (empty data).
func (a *Assessor) Assess(data []byte, h Hint) (Score, error) {
	if len(data) == 0 {
		return Score{Verdict: FailEmpty, FailReason: "no bytes supplied"}, ErrEmptyData
	}

	cfg, _ := Lookup(a.thresholds, h)

	cfgHeader, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Score{IsCorrupt: true, Verdict: FailCorrupt, FailReason: fmt.Sprintf("decode config: %v", err)}, nil
	}
	if cfgHeader.Width*cfgHeader.Height > MaxDecodePixels {
		return Score{
			Width: cfgHeader.Width, Height: cfgHeader.Height,
			Format:     format,
			Verdict:    FailTooLarge,
			FailReason: fmt.Sprintf("decoded size %d*%d exceeds cap %d pixels", cfgHeader.Width, cfgHeader.Height, MaxDecodePixels),
		}, nil
	}

	img, format2, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Score{IsCorrupt: true, Verdict: FailCorrupt, FailReason: fmt.Sprintf("decode: %v", err)}, nil
	}
	if format2 != "" {
		format = format2
	}

	b := img.Bounds()
	w, ht := b.Dx(), b.Dy()
	if w <= 0 || ht <= 0 {
		return Score{IsCorrupt: true, Format: format, Verdict: FailCorrupt, FailReason: "zero-dimension image"}, nil
	}

	pixelCount := float64(w * ht)
	bpp := float64(len(data)) / pixelCount
	mp := pixelCount / 1_000_000
	aspect := float64(w) / float64(ht)

	score := Score{
		Width:         w,
		Height:        ht,
		Megapixels:    mp,
		BytesPerPixel: bpp,
		AspectRatio:   aspect,
		Format:        format,
	}

	if w < cfg.MinWidth || ht < cfg.MinHeight {
		score.Verdict = FailLowRes
		score.FailReason = fmt.Sprintf("%dx%d < required %dx%d", w, ht, cfg.MinWidth, cfg.MinHeight)
		return score, nil
	}
	if bpp < cfg.MinBytesPerPixel {
		score.Verdict = FailSmallBytes
		score.FailReason = fmt.Sprintf("bpp %.3f < required %.3f", bpp, cfg.MinBytesPerPixel)
		return score, nil
	}
	if cfg.AspectTolerance > 0 && math.Abs(aspect-cfg.AspectTarget) > cfg.AspectTolerance {
		score.Verdict = FailWrongAspect
		score.FailReason = fmt.Sprintf("aspect %.3f outside %.3f ± %.3f", aspect, cfg.AspectTarget, cfg.AspectTolerance)
		return score, nil
	}

	blurVar := laplacianVariance(img)
	score.BlurVariance = blurVar
	if blurVar < cfg.MinBlurVariance {
		score.Verdict = FailBlurry
		score.FailReason = fmt.Sprintf("blur variance %.2f < required %.2f", blurVar, cfg.MinBlurVariance)
		return score, nil
	}

	score.Verdict = Pass
	return score, nil
}

// Assess is a package-level convenience that uses a default assessor.
func Assess(data []byte, h Hint) (Score, error) {
	return defaultAssessor.Assess(data, h)
}

var defaultAssessor = NewAssessor()
