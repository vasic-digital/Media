package quality

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math/rand"
	"testing"
)

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return buf.Bytes()
}

func encodeJPEG(t *testing.T, img image.Image, quality int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatalf("encode JPEG: %v", err)
	}
	return buf.Bytes()
}

// noisyImage produces an image full of high-frequency noise. It decodes with
// high Laplacian variance and will not be flagged as blurry.
func noisyImage(w, h int, seed int64) image.Image {
	r := rand.New(rand.NewSource(seed))
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(r.Intn(256))
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 0xff})
		}
	}
	return img
}

// uniformImage produces a solid-colour image. Its Laplacian variance is zero,
// so it will always trip the blur gate.
func uniformImage(w, h int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// checkerboard produces a high-contrast alternating pattern with very high
// Laplacian variance, guaranteed to pass the blur threshold.
func checkerboard(w, h, cell int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(0xff)
			if ((x/cell)+(y/cell))%2 == 0 {
				v = 0x10
			}
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 0xff})
		}
	}
	return img
}
