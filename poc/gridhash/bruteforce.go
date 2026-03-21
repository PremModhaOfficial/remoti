package gridhash

import "image"

// BruteForceDetector does a direct byte comparison of the entire frame.
// This represents the simplest possible approach: memcmp the full pixel buffer.
type BruteForceDetector struct {
	prevPix []byte
}

// NewBruteForceDetector creates a brute-force pixel-diff detector.
func NewBruteForceDetector() *BruteForceDetector {
	return &BruteForceDetector{}
}

// Update compares the frame against the previous one byte-by-byte.
// Returns true if any pixel changed.
func (b *BruteForceDetector) Update(img *image.RGBA) bool {
	pix := img.Pix
	if b.prevPix == nil {
		b.prevPix = make([]byte, len(pix))
		copy(b.prevPix, pix)
		return true // first frame
	}

	changed := false
	for i := 0; i < len(pix); i++ {
		if pix[i] != b.prevPix[i] {
			changed = true
			break
		}
	}
	copy(b.prevPix, pix)
	return changed
}

// UpdateWithRegions does brute-force diff but reports which grid sections changed.
// This gives the same output as the grid detector but without hashing.
func (b *BruteForceDetector) UpdateWithRegions(img *image.RGBA, config GridConfig) []DirtySection {
	pix := img.Pix
	if b.prevPix == nil {
		b.prevPix = make([]byte, len(pix))
		copy(b.prevPix, pix)
		// First frame: all dirty.
		var dirty []DirtySection
		for row := 0; row < config.Rows; row++ {
			for col := 0; col < config.Cols; col++ {
				dirty = append(dirty, DirtySection{
					Col:    col,
					Row:    row,
					Index:  row*config.Cols + col,
					Bounds: config.SectionBounds(col, row),
				})
			}
		}
		return dirty
	}

	var dirty []DirtySection
	for row := 0; row < config.Rows; row++ {
		for col := 0; col < config.Cols; col++ {
			bounds := config.SectionBounds(col, row)
			sectionChanged := false
			for y := bounds.Min.Y; y < bounds.Max.Y && !sectionChanged; y++ {
				start := img.PixOffset(bounds.Min.X, y)
				end := img.PixOffset(bounds.Max.X, y)
				for i := start; i < end; i++ {
					if pix[i] != b.prevPix[i] {
						sectionChanged = true
						break
					}
				}
			}
			if sectionChanged {
				dirty = append(dirty, DirtySection{
					Col:    col,
					Row:    row,
					Index:  row*config.Cols + col,
					Bounds: bounds,
				})
			}
		}
	}
	copy(b.prevPix, pix)
	return dirty
}
