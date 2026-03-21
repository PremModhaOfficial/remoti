// Package gridhash implements spatial hash-based screen change detection.
//
// Instead of diffing every pixel on a 1920x1080 screen (2M+ pixels),
// it divides the screen into a grid of sections and hashes each section
// using xxHash64. Only sections where the hash changed are reported as dirty.
//
// Recommended configuration: 16x9 grid (144 sections of 120x120 pixels)
// Total change detection time: ~0.5ms on modern CPUs.
package gridhash

import (
	"image"
	"sync"

	"github.com/cespare/xxhash/v2"
)

// GridConfig defines the grid dimensions and frame size.
type GridConfig struct {
	// FrameWidth and FrameHeight are the full screen dimensions.
	FrameWidth  int
	FrameHeight int
	// Cols and Rows define the grid. Default: 16x9 for 1920x1080.
	Cols int
	Rows int
}

// DefaultConfig returns the recommended 16x9 grid for 1080p.
func DefaultConfig() GridConfig {
	return GridConfig{
		FrameWidth:  1920,
		FrameHeight: 1080,
		Cols:        16,
		Rows:        9,
	}
}

// SectionBounds returns the pixel rectangle for grid cell (col, row).
func (c GridConfig) SectionBounds(col, row int) image.Rectangle {
	secW := c.FrameWidth / c.Cols
	secH := c.FrameHeight / c.Rows
	x0 := col * secW
	y0 := row * secH
	// Last column/row absorbs remainder pixels.
	x1 := x0 + secW
	y1 := y0 + secH
	if col == c.Cols-1 {
		x1 = c.FrameWidth
	}
	if row == c.Rows-1 {
		y1 = c.FrameHeight
	}
	return image.Rect(x0, y0, x1, y1)
}

// SectionCount returns the total number of grid sections.
func (c GridConfig) SectionCount() int {
	return c.Cols * c.Rows
}

// DirtySection represents a grid cell whose content changed.
type DirtySection struct {
	Col    int
	Row    int
	Index  int // Col + Row*Cols
	Bounds image.Rectangle
}

// Detector performs frame-to-frame change detection using spatial hashing.
type Detector struct {
	config GridConfig
	// hashes stores the previous frame's hash for each section.
	hashes []uint64
	// initialized tracks whether we have a previous frame to compare against.
	initialized bool
	// pool of hashers to reduce allocation.
	pool sync.Pool
}

// NewDetector creates a change detector with the given grid configuration.
func NewDetector(config GridConfig) *Detector {
	return &Detector{
		config: config,
		hashes: make([]uint64, config.SectionCount()),
		pool: sync.Pool{
			New: func() any {
				return xxhash.New()
			},
		},
	}
}

// hashSection computes xxHash64 of a section's pixel data in-place.
// It feeds row segments directly to the hasher, avoiding any pixel copy.
// This works because image.RGBA.Pix is a contiguous byte slice with stride.
func (d *Detector) hashSection(img *image.RGBA, bounds image.Rectangle) uint64 {
	h := d.pool.Get().(*xxhash.Digest)
	h.Reset()

	// For each row in the section, hash the contiguous RGBA byte range.
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		start := img.PixOffset(bounds.Min.X, y)
		end := img.PixOffset(bounds.Max.X, y)
		h.Write(img.Pix[start:end])
	}

	sum := h.Sum64()
	d.pool.Put(h)
	return sum
}

// Update processes a new frame and returns the list of changed sections.
// On the first call, all sections are considered dirty.
//
// The image must be *image.RGBA with dimensions matching the config.
// This method is NOT safe for concurrent use on the same Detector.
func (d *Detector) Update(img *image.RGBA) []DirtySection {
	totalSections := d.config.SectionCount()

	// Compute all section hashes. Use parallel goroutines for sections.
	newHashes := make([]uint64, totalSections)

	// For 144 sections, parallelism with a WaitGroup is straightforward.
	// Each goroutine handles one row of sections to reduce goroutine overhead.
	var wg sync.WaitGroup
	for row := 0; row < d.config.Rows; row++ {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			for col := 0; col < d.config.Cols; col++ {
				idx := row*d.config.Cols + col
				bounds := d.config.SectionBounds(col, row)
				newHashes[idx] = d.hashSection(img, bounds)
			}
		}(row)
	}
	wg.Wait()

	// Compare against previous hashes and collect dirty sections.
	var dirty []DirtySection
	for row := 0; row < d.config.Rows; row++ {
		for col := 0; col < d.config.Cols; col++ {
			idx := row*d.config.Cols + col
			if !d.initialized || newHashes[idx] != d.hashes[idx] {
				dirty = append(dirty, DirtySection{
					Col:    col,
					Row:    row,
					Index:  idx,
					Bounds: d.config.SectionBounds(col, row),
				})
			}
		}
	}

	// Store hashes for next comparison.
	copy(d.hashes, newHashes)
	d.initialized = true

	return dirty
}

// Reset clears the stored hashes, causing the next Update to treat all
// sections as dirty (equivalent to a VNC I-frame).
func (d *Detector) Reset() {
	d.initialized = false
	for i := range d.hashes {
		d.hashes[i] = 0
	}
}

// Config returns the detector's grid configuration.
func (d *Detector) Config() GridConfig {
	return d.config
}

// ExtractSection copies the pixel data for a single dirty section into
// a new *image.RGBA. Use this to send only changed regions downstream.
func ExtractSection(src *image.RGBA, bounds image.Rectangle) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		srcStart := src.PixOffset(bounds.Min.X, bounds.Min.Y+y)
		srcEnd := srcStart + bounds.Dx()*4
		dstStart := dst.PixOffset(0, y)
		copy(dst.Pix[dstStart:], src.Pix[srcStart:srcEnd])
	}
	return dst
}

// ComposeDiffImage creates a full-frame image with only dirty sections visible
// and all other sections black. Useful for sending a "diff view" to an AI.
func ComposeDiffImage(src *image.RGBA, dirty []DirtySection) *image.RGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	// Allocate black image (zero-initialized = black in RGBA).
	out := image.NewRGBA(image.Rect(0, 0, w, h))

	for _, sec := range dirty {
		b := sec.Bounds
		for y := b.Min.Y; y < b.Max.Y; y++ {
			srcOff := src.PixOffset(b.Min.X, y)
			dstOff := out.PixOffset(b.Min.X, y)
			rowBytes := b.Dx() * 4
			copy(out.Pix[dstOff:dstOff+rowBytes], src.Pix[srcOff:srcOff+rowBytes])
		}
	}
	return out
}
