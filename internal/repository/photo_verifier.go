package repository

import "context"

// PhotoVerdict is the result of verifying an uploaded image.
type PhotoVerdict struct {
	// Approved is the final decision; false means the upload should be rejected.
	Approved bool
	// HasFace / FaceCount describe face-presence detection.
	HasFace   bool
	FaceCount int
	// NSFWScore is the probability the image is explicit (0..1).
	NSFWScore float64
	// Reasons lists why it was rejected (e.g. "no_face", "nsfw").
	Reasons []string
}

// PhotoVerifier is the port for the external CV service that checks profile
// photos for a visible face and absence of explicit (NSFW) content. The
// implementation lives in internal/adapter/photoverify and calls a Python
// (FastAPI) service over HTTP.
type PhotoVerifier interface {
	// Verify inspects raw image bytes. An error means the service was
	// unreachable/timed out; callers should fail open (allow the upload).
	Verify(ctx context.Context, imageData []byte) (*PhotoVerdict, error)
}
