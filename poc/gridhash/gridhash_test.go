package gridhash

import (
	"image"
	"math/rand"
	"testing"
)

// makeFrame creates a random RGBA image of the given size.
func makeFrame(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rand.Read(img.Pix)
	return img
}

// makeIdenticalFrames returns two identical frames.
func makeIdenticalFrames(w, h int) (*image.RGBA, *image.RGBA) {
	a := makeFrame(w, h)
	b := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(b.Pix, a.Pix)
	return a, b
}

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	if c.Cols != 16 || c.Rows != 9 {
		t.Errorf("expected 16x9, got %dx%d", c.Cols, c.Rows)
	}
	if c.SectionCount() != 144 {
		t.Errorf("expected 144 sections, got %d", c.SectionCount())
	}
}

func TestSectionBounds(t *testing.T) {
	c := DefaultConfig()
	// First section.
	b := c.SectionBounds(0, 0)
	if b.Min.X != 0 || b.Min.Y != 0 {
		t.Errorf("first section should start at (0,0), got (%d,%d)", b.Min.X, b.Min.Y)
	}
	if b.Dx() != 120 || b.Dy() != 120 {
		t.Errorf("first section should be 120x120, got %dx%d", b.Dx(), b.Dy())
	}
	// Last section should extend to frame edge.
	b = c.SectionBounds(c.Cols-1, c.Rows-1)
	if b.Max.X != 1920 || b.Max.Y != 1080 {
		t.Errorf("last section should end at (1920,1080), got (%d,%d)", b.Max.X, b.Max.Y)
	}
}

func TestFirstFrameAllDirty(t *testing.T) {
	d := NewDetector(DefaultConfig())
	img := makeFrame(1920, 1080)
	dirty := d.Update(img)
	if len(dirty) != 144 {
		t.Errorf("first frame should have all 144 sections dirty, got %d", len(dirty))
	}
}

func TestIdenticalFrameNoDirty(t *testing.T) {
	d := NewDetector(DefaultConfig())
	a, b := makeIdenticalFrames(1920, 1080)
	d.Update(a)
	dirty := d.Update(b)
	if len(dirty) != 0 {
		t.Errorf("identical frames should have 0 dirty sections, got %d", len(dirty))
	}
}

func TestSingleSectionChange(t *testing.T) {
	d := NewDetector(DefaultConfig())
	a, b := makeIdenticalFrames(1920, 1080)
	d.Update(a)

	// Modify a single pixel in section (3, 2) which covers x=360..479, y=240..359.
	off := b.PixOffset(400, 300)
	b.Pix[off] ^= 0xFF   // flip red channel
	b.Pix[off+1] ^= 0xFF // flip green channel

	dirty := d.Update(b)
	if len(dirty) != 1 {
		t.Fatalf("expected 1 dirty section, got %d", len(dirty))
	}
	if dirty[0].Col != 3 || dirty[0].Row != 2 {
		t.Errorf("expected dirty section at (3,2), got (%d,%d)", dirty[0].Col, dirty[0].Row)
	}
}

func TestReset(t *testing.T) {
	d := NewDetector(DefaultConfig())
	img := makeFrame(1920, 1080)
	d.Update(img)
	d.Reset()
	dirty := d.Update(img)
	if len(dirty) != 144 {
		t.Errorf("after reset, all sections should be dirty, got %d", len(dirty))
	}
}

func TestExtractSection(t *testing.T) {
	img := makeFrame(1920, 1080)
	bounds := image.Rect(120, 120, 240, 240)
	sec := ExtractSection(img, bounds)
	if sec.Bounds().Dx() != 120 || sec.Bounds().Dy() != 120 {
		t.Errorf("extracted section should be 120x120, got %dx%d", sec.Bounds().Dx(), sec.Bounds().Dy())
	}
	// Verify pixel values match.
	for y := 0; y < 120; y++ {
		for x := 0; x < 120; x++ {
			srcOff := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)
			dstOff := sec.PixOffset(x, y)
			for c := 0; c < 4; c++ {
				if img.Pix[srcOff+c] != sec.Pix[dstOff+c] {
					t.Fatalf("pixel mismatch at (%d,%d) channel %d", x, y, c)
				}
			}
		}
	}
}

func TestComposeDiffImage(t *testing.T) {
	img := makeFrame(1920, 1080)
	dirty := []DirtySection{
		{Col: 0, Row: 0, Index: 0, Bounds: image.Rect(0, 0, 120, 120)},
	}
	out := ComposeDiffImage(img, dirty)
	if out.Bounds().Dx() != 1920 || out.Bounds().Dy() != 1080 {
		t.Errorf("diff image should be full frame size")
	}
	// Check that dirty section has non-zero pixels (extremely likely with random data).
	hasNonZero := false
	for i := 0; i < 120*4; i++ {
		if out.Pix[i] != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Error("dirty section in diff image should have non-zero pixels")
	}
	// Check that a clean section is all zeros.
	// Section (8, 4) = middle of screen, index 8+4*16=72.
	cleanBounds := DefaultConfig().SectionBounds(8, 4)
	allZero := true
	for y := cleanBounds.Min.Y; y < cleanBounds.Max.Y && allZero; y++ {
		off := out.PixOffset(cleanBounds.Min.X, y)
		for x := 0; x < cleanBounds.Dx()*4; x++ {
			if out.Pix[off+x] != 0 {
				allZero = false
				break
			}
		}
	}
	if !allZero {
		t.Error("clean section in diff image should be all zeros")
	}
}

// --- Benchmarks ---

func BenchmarkUpdate_1080p_16x9(b *testing.B) {
	d := NewDetector(DefaultConfig())
	img := makeFrame(1920, 1080)
	// Prime with first frame.
	d.Update(img)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Update(img)
	}
}

func BenchmarkUpdate_1080p_32x18(b *testing.B) {
	cfg := GridConfig{FrameWidth: 1920, FrameHeight: 1080, Cols: 32, Rows: 18}
	d := NewDetector(cfg)
	img := makeFrame(1920, 1080)
	d.Update(img)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Update(img)
	}
}

func BenchmarkUpdate_4K_16x9(b *testing.B) {
	cfg := GridConfig{FrameWidth: 3840, FrameHeight: 2160, Cols: 16, Rows: 9}
	d := NewDetector(cfg)
	img := makeFrame(3840, 2160)
	d.Update(img)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Update(img)
	}
}

func BenchmarkHashSection(b *testing.B) {
	d := NewDetector(DefaultConfig())
	img := makeFrame(1920, 1080)
	bounds := DefaultConfig().SectionBounds(0, 0) // 120x120 section
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.hashSection(img, bounds)
	}
}

func BenchmarkExtractSection(b *testing.B) {
	img := makeFrame(1920, 1080)
	bounds := DefaultConfig().SectionBounds(0, 0)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ExtractSection(img, bounds)
	}
}
