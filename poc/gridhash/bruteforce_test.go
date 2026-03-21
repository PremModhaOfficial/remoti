package gridhash

import "testing"

func TestBruteForceFirstFrame(t *testing.T) {
	d := NewBruteForceDetector()
	img := makeFrame(1920, 1080)
	if !d.Update(img) {
		t.Error("first frame should report changed")
	}
}

func TestBruteForceIdentical(t *testing.T) {
	d := NewBruteForceDetector()
	a, b := makeIdenticalFrames(1920, 1080)
	d.Update(a)
	if d.Update(b) {
		t.Error("identical frames should report unchanged")
	}
}

func TestBruteForceWithRegions(t *testing.T) {
	d := NewBruteForceDetector()
	cfg := DefaultConfig()
	a, b := makeIdenticalFrames(1920, 1080)
	d.UpdateWithRegions(a, cfg)

	// Change one pixel in section (3,2).
	off := b.PixOffset(400, 300)
	b.Pix[off] ^= 0xFF
	dirty := d.UpdateWithRegions(b, cfg)
	if len(dirty) != 1 {
		t.Fatalf("expected 1 dirty section, got %d", len(dirty))
	}
	if dirty[0].Col != 3 || dirty[0].Row != 2 {
		t.Errorf("expected (3,2), got (%d,%d)", dirty[0].Col, dirty[0].Row)
	}
}

// --- Three-way comparative benchmarks ---

// Static screen: nothing changed between frames.
func BenchmarkBrute_StaticScreen(b *testing.B) {
	d := NewBruteForceDetector()
	img := makeFrame(1920, 1080)
	d.Update(img)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Update(img)
	}
}

func BenchmarkBruteRegions_StaticScreen(b *testing.B) {
	d := NewBruteForceDetector()
	cfg := DefaultConfig()
	img := makeFrame(1920, 1080)
	d.UpdateWithRegions(img, cfg)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.UpdateWithRegions(img, cfg)
	}
}

// Full change: every pixel different.
func BenchmarkBrute_FullChange(b *testing.B) {
	d := NewBruteForceDetector()
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

func BenchmarkBruteRegions_FullChange(b *testing.B) {
	d := NewBruteForceDetector()
	cfg := DefaultConfig()
	img1 := makeFrame(1920, 1080)
	img2 := makeFrame(1920, 1080)
	d.UpdateWithRegions(img1, cfg)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			d.UpdateWithRegions(img2, cfg)
		} else {
			d.UpdateWithRegions(img1, cfg)
		}
	}
}
