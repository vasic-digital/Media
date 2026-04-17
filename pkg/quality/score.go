package quality

import "fmt"

// Verdict summarises the outcome of a quality assessment. Pass means the image
// is suitable to serve; every other value is a specific failure reason so
// callers can log it or surface it as structured telemetry.
type Verdict int

const (
	Pass Verdict = iota
	FailLowRes
	FailBlurry
	FailSmallBytes
	FailCorrupt
	FailWrongAspect
	FailTooLarge
	FailEmpty
)

// String returns the canonical lower_snake_case name for v so it can be stored
// in the database and logged consistently.
func (v Verdict) String() string {
	switch v {
	case Pass:
		return "pass"
	case FailLowRes:
		return "fail_lowres"
	case FailBlurry:
		return "fail_blurry"
	case FailSmallBytes:
		return "fail_small_bytes"
	case FailCorrupt:
		return "fail_corrupt"
	case FailWrongAspect:
		return "fail_wrong_aspect"
	case FailTooLarge:
		return "fail_too_large"
	case FailEmpty:
		return "fail_empty"
	default:
		return fmt.Sprintf("verdict_%d", int(v))
	}
}

// Score carries the measurements and final Verdict for an assessed image. The
// zero value is not useful; callers should only consult fields on a Score
// returned from Assess.
type Score struct {
	Width         int
	Height        int
	Megapixels    float64
	BytesPerPixel float64
	BlurVariance  float64
	AspectRatio   float64
	Format        string
	IsCorrupt     bool
	Verdict       Verdict
	FailReason    string
}

// OK reports whether the image passed the gate.
func (s Score) OK() bool { return s.Verdict == Pass }
