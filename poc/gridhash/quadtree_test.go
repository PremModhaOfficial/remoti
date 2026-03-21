package gridhash

import (
	"testing"
)

func TestQuadFirstFrameAllDirty(t *testing.T) {
	d := NewQuadDetector(DefaultQuadConfig())
	img := makeFrame(1920, 1080)
	dirty := d.Update(img)
	// First frame returns full frame as one dirty rect.
	if len(dirty) != 1 {
		t.Fatalf("first frame should return 1 dirty rect (full frame), got %d", len(dirty))
	}
	if dirty[0].Bounds.Dx() != 1920 || dirty[0].Bounds.Dy() != 1080 {
		t.Errorf("expected full frame bounds, got %v", dirty[0].Bounds)
	}
}

func TestQuadIdenticalFrameNoDirty(t *testing.T) {
	d := NewQuadDetector(DefaultQuadConfig())
	a, b := makeIdenticalFrames(1920, 1080)
	d.Update(a)
	dirty := d.Update(b)
	if len(dirty) != 0 {
		t.Errorf("identical frames should have 0 dirty rects, got %d", len(dirty))
	}
}

func TestQuadSinglePixelChange(t *testing.T) {
	d := NewQuadDetector(DefaultQuadConfig())
	a, b := makeIdenticalFrames(1920, 1080)
	d.Update(a)

	// Flip one pixel near center.
	off := b.PixOffset(960, 540)
	b.Pix[off] ^= 0xFF

	dirty := d.Update(b)
	if len(dirty) == 0 {
		t.Fatal("expected at least 1 dirty rect for single pixel change")
	}
	// All dirty rects should contain the changed pixel.
	for _, dr := range dirty {
		if !dr.Bounds.Min.In(dr.Bounds) {
			t.Errorf("invalid bounds: %v", dr.Bounds)
		}
	}
	// The dirty rect should be a small leaf tile, not the full frame.
	for _, dr := range dirty {
		if dr.Bounds.Dx() > 960 || dr.Bounds.Dy() > 540 {
			t.Errorf("dirty rect too large for single pixel change: %v", dr.Bounds)
		}
	}
	t.Logf("single pixel change produced %d dirty rects, smallest: %dx%d",
		len(dirty), dirty[len(dirty)-1].Bounds.Dx(), dirty[len(dirty)-1].Bounds.Dy())
}

func TestQuadReset(t *testing.T) {
	d := NewQuadDetector(DefaultQuadConfig())
	img := makeFrame(1920, 1080)
	d.Update(img)
	d.Reset()
	dirty := d.Update(img)
	if len(dirty) != 1 {
		t.Errorf("after reset, should return full frame dirty, got %d rects", len(dirty))
	}
}

// --- Comparative Benchmarks: Quadtree vs Fixed Grid ---

// BenchmarkQuad_StaticScreen: Best case for quadtree — nothing changed.
// Quadtree should win: 1 hash vs 144 hashes.
func BenchmarkQuad_StaticScreen(b *testing.B) {
	d := NewQuadDetector(DefaultQuadConfig())
	img := makeFrame(1920, 1080)
	d.Update(img)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Update(img)
	}
}

// BenchmarkGrid_StaticScreen: Fixed grid on static screen — must hash all 144 sections.
func BenchmarkGrid_StaticScreen(b *testing.B) {
	d := NewDetector(DefaultConfig())
	img := makeFrame(1920, 1080)
	d.Update(img)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Update(img)
	}
}

// BenchmarkQuad_SmallChange: One pixel changed — quadtree recurses into one branch.
func BenchmarkQuad_SmallChange(b *testing.B) {
	d := NewQuadDetector(DefaultQuadConfig())
	a, img := makeIdenticalFrames(1920, 1080)
	d.Update(a)
	// Flip one pixel.
	off := img.PixOffset(960, 540)
	img.Pix[off] ^= 0xFF
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Re-prime with original to make the change detectable each iteration.
		d.hashes[toKey(a.Bounds())] = hashRect(a, a.Bounds())
		d.Update(img)
	}
}

// BenchmarkGrid_SmallChange: One pixel changed — grid still hashes all 144 sections.
func BenchmarkGrid_SmallChange(b *testing.B) {
	d := NewDetector(DefaultConfig())
	a, img := makeIdenticalFrames(1920, 1080)
	d.Update(a)
	off := img.PixOffset(960, 540)
	img.Pix[off] ^= 0xFF
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Re-prime with original.
		copy(d.hashes, make([]uint64, 144))
		d.Update(a)
		d.Update(img)
	}
}

// BenchmarkQuad_FullChange: Worst case for quadtree — everything changed.
func BenchmarkQuad_FullChange(b *testing.B) {
	d := NewQuadDetector(DefaultQuadConfig())
	img1 := makeFrame(1920, 1080)
	img2 := makeFrame(1920, 1080)
	d.Update(img1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			d.Update(img2)
		} else {
			d.Update(img1)
		}
	}
}

// BenchmarkGrid_FullChange: Full change — grid hashes all 144 sections.
func BenchmarkGrid_FullChange(b *testing.B) {
	d := NewDetector(DefaultConfig())
	img1 := makeFrame(1920, 1080)
	img2 := makeFrame(1920, 1080)
	d.Update(img1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			d.Update(img2)
		} else {
			d.Update(img1)
		}
	}
}
