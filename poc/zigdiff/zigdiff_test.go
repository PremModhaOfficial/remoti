package zigdiff_test

// To run these benchmarks you need the shared library:
//   cd poc/zigdiff && zig build-lib -dynamic -OReleaseFast -femit-bin=libgriddiff.so griddiff.zig
//   go test -bench=. -benchtime=5s -count=1 ./poc/zigdiff/

import (
	"image"
	"math/rand"
	"testing"

	"remoti/poc/zigdiff"
)

func makeFrame(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rand.Read(img.Pix)
	return img
}

func makeIdenticalFrames(w, h int) (*image.RGBA, *image.RGBA) {
	a := makeFrame(w, h)
	b := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(b.Pix, a.Pix)
	return a, b
}

// TestCorrectness verifies the Zig differ matches expected behavior.
func TestCorrectness(t *testing.T) {
	s := zigdiff.NewState()
	defer s.Free()

	a, b := makeIdenticalFrames(1920, 1080)

	// First frame: all 144 sections dirty.
	dirty := s.Update(a, true)
	if len(dirty) != 144 {
		t.Fatalf("first frame: want 144 dirty, got %d", len(dirty))
	}

	// Identical frame: 0 dirty.
	dirty = s.Update(b, false)
	if len(dirty) != 0 {
		t.Fatalf("identical frame: want 0 dirty, got %d", len(dirty))
	}

	// Flip one pixel in section (col=3, row=2) → index = 2*16+3 = 35.
	off := b.PixOffset(400, 300)
	b.Pix[off] ^= 0xFF

	dirty = s.Update(b, false)
	if len(dirty) != 1 {
		t.Fatalf("single pixel flip: want 1 dirty, got %d", len(dirty))
	}
	if dirty[0] != 35 {
		t.Fatalf("wrong dirty section: want 35, got %d", dirty[0])
	}
}

func TestCorrectnessEq(t *testing.T) {
	a, b := makeIdenticalFrames(1920, 1080)

	// Identical: 0 dirty.
	dirty := zigdiff.UpdateEq(a, b.Pix)
	if len(dirty) != 0 {
		t.Fatalf("identical: want 0 dirty, got %d", len(dirty))
	}

	// Flip pixel in section 35.
	off := b.PixOffset(400, 300)
	b.Pix[off] ^= 0xFF
	dirty = zigdiff.UpdateEq(b, a.Pix)
	if len(dirty) != 1 {
		t.Fatalf("single flip: want 1 dirty, got %d", len(dirty))
	}
	if dirty[0] != 35 {
		t.Fatalf("wrong section: want 35, got %d", dirty[0])
	}
}

// --- Benchmarks ---

// BenchmarkZigHash_Static: steady-state hash diff, 0 sections changed.
func BenchmarkZigHash_Static(b *testing.B) {
	s := zigdiff.NewState()
	defer s.Free()
	img := makeFrame(1920, 1080)
	s.Update(img, true) // prime
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Update(img, false)
	}
}

// BenchmarkZigHash_1Dirty: hash diff with 1 section changed per frame.
func BenchmarkZigHash_1Dirty(b *testing.B) {
	s := zigdiff.NewState()
	defer s.Free()
	a, bImg := makeIdenticalFrames(1920, 1080)
	off := bImg.PixOffset(400, 300)
	bImg.Pix[off] ^= 0xFF
	s.Update(a, true)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Alternate between a and b to keep 1 section always dirty.
		if i%2 == 0 {
			s.Update(bImg, false)
		} else {
			s.Update(a, false)
		}
	}
}

// BenchmarkZigEq_Static: SIMD eq diff, 0 sections changed.
func BenchmarkZigEq_Static(b *testing.B) {
	img := makeFrame(1920, 1080)
	prev := make([]byte, len(img.Pix))
	copy(prev, img.Pix)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		zigdiff.UpdateEq(img, prev)
	}
}

// BenchmarkZigEq_1Dirty: SIMD eq diff, 1 section changed.
func BenchmarkZigEq_1Dirty(b *testing.B) {
	a, bImg := makeIdenticalFrames(1920, 1080)
	off := bImg.PixOffset(400, 300)
	bImg.Pix[off] ^= 0xFF
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		zigdiff.UpdateEq(bImg, a.Pix)
	}
}
