// Package quality provides deterministic image-quality assessment for cover art
// and similar media assets served to client applications.
//
// It decodes bytes, measures resolution, bytes-per-pixel, aspect ratio, and a
// Laplacian-variance blur estimate, and returns a Verdict keyed off per-Hint
// thresholds. Callers use the Verdict to decide whether to serve the image or
// request a replacement from an upstream provider.
//
// The package is pure Go and CGo-free so it runs identically in containerised
// builds on all supported platforms.
package quality
