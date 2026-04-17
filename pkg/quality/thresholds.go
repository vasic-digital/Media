package quality

// Hint names the class of image being assessed so per-type thresholds apply.
type Hint string

const (
	HintMoviePoster Hint = "movie_poster"
	HintTvPoster    Hint = "tv_poster"
	HintMusicAlbum  Hint = "music_album"
	HintBookCover   Hint = "book_cover"
	HintGameCover   Hint = "game_cover"
	HintBackdrop    Hint = "backdrop"
	HintGeneric     Hint = "generic"
)

// Threshold defines the minimum acceptable quality signals for a Hint.
//
// Width/Height are pixel dimensions, MinBlurVariance is the Laplacian
// variance floor below which an image is considered blurry, MinBytesPerPixel
// guards against empty pixels packed into a large dimension, and AspectTarget
// with AspectTolerance expresses the allowed width/height ratio (a tolerance
// of 0 skips the aspect check).
type Threshold struct {
	MinWidth         int
	MinHeight        int
	MinBlurVariance  float64
	MinBytesPerPixel float64
	AspectTarget     float64
	AspectTolerance  float64
}

// MaxDecodePixels caps the total pixel count we will decode. It guards against
// decompression-bomb inputs (many megabytes of encoded data expanding into
// gigabytes of pixels). 64 megapixels is enough for an 8K (7680x4320) backdrop.
const MaxDecodePixels = 64_000_000

// DefaultThresholds returns the built-in production thresholds. Callers may
// copy this map and override specific hints via config.
func DefaultThresholds() map[Hint]Threshold {
	return map[Hint]Threshold{
		HintMoviePoster: {MinWidth: 600, MinHeight: 900, MinBlurVariance: 80, MinBytesPerPixel: 0.40, AspectTarget: 2.0 / 3.0, AspectTolerance: 0.05},
		HintTvPoster:    {MinWidth: 600, MinHeight: 900, MinBlurVariance: 80, MinBytesPerPixel: 0.40, AspectTarget: 2.0 / 3.0, AspectTolerance: 0.05},
		HintMusicAlbum:  {MinWidth: 500, MinHeight: 500, MinBlurVariance: 70, MinBytesPerPixel: 0.35, AspectTarget: 1.0, AspectTolerance: 0.03},
		HintBookCover:   {MinWidth: 400, MinHeight: 600, MinBlurVariance: 60, MinBytesPerPixel: 0.30, AspectTarget: 2.0 / 3.0, AspectTolerance: 0.10},
		HintGameCover:   {MinWidth: 600, MinHeight: 800, MinBlurVariance: 80, MinBytesPerPixel: 0.40, AspectTarget: 3.0 / 4.0, AspectTolerance: 0.10},
		HintBackdrop:    {MinWidth: 1280, MinHeight: 720, MinBlurVariance: 100, MinBytesPerPixel: 0.50, AspectTarget: 16.0 / 9.0, AspectTolerance: 0.05},
		HintGeneric:     {MinWidth: 300, MinHeight: 300, MinBlurVariance: 60, MinBytesPerPixel: 0.25},
	}
}

// Lookup returns the Threshold for h, falling back to HintGeneric if h is
// unknown. The second return value reports whether the hint was recognised,
// so callers can log unusual values without failing.
func Lookup(thresholds map[Hint]Threshold, h Hint) (Threshold, bool) {
	if t, ok := thresholds[h]; ok {
		return t, true
	}
	return thresholds[HintGeneric], false
}
