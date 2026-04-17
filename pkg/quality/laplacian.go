package quality

import "image"

// laplacianVariance computes the variance of a 3x3 Laplacian convolution over
// the luminance channel of img. Sharp images have high variance (edges produce
// large kernel responses); blurry images have low variance because edges are
// smoothed away. This is the standard cheap blur-detection heuristic.
//
// The Laplacian kernel used is
//
//	0  1  0
//	1 -4  1
//	0  1  0
//
// Luminance uses the Rec. 601 coefficients (0.299 R + 0.587 G + 0.114 B).
func laplacianVariance(img image.Image) float64 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 3 || h < 3 {
		return 0
	}

	lum := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			lum[y*w+x] = 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bl>>8)
		}
	}

	var sum, sumSq float64
	var n int
	for y := 1; y < h-1; y++ {
		row := y * w
		for x := 1; x < w-1; x++ {
			idx := row + x
			v := lum[idx-w] + lum[idx+w] + lum[idx-1] + lum[idx+1] - 4*lum[idx]
			sum += v
			sumSq += v * v
			n++
		}
	}
	if n == 0 {
		return 0
	}
	mean := sum / float64(n)
	return sumSq/float64(n) - mean*mean
}
