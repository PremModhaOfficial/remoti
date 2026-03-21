package gridhash

import (
	"image"

	"github.com/cespare/xxhash/v2"
)

// QuadConfig controls the quadtree-based change detector.
type QuadConfig struct {
	FrameWidth  int
	FrameHeight int
	// MinTileSize is the smallest tile dimension before recursion stops.
	// Recommended: 64 (64x64 tiles, matching HEVC CTU minimum).
	MinTileSize int
	// MaxDepth limits recursion. 0 = unlimited (bounded by MinTileSize).
	MaxDepth int
}

// DefaultQuadConfig returns recommended settings for 1080p.
func DefaultQuadConfig() QuadConfig {
	return QuadConfig{
		FrameWidth:  1920,
		FrameHeight: 1080,
		MinTileSize: 60, // Divides evenly: 1920/32=60, 1080/18=60
		MaxDepth:    0,
	}
}

// QuadDetector performs frame-to-frame change detection using a quadtree.
//
// On a mostly-static screen, it hashes the full frame first. If unchanged,
// it returns immediately (one hash comparison). If changed, it recursively
// subdivides into quadrants, only descending into changed regions.
//
// Best case (static screen):  O(1) — one hash
// Typical case (small change): O(log N) — a few subdivisions
// Worst case (full change):    O(N) — equivalent to fixed grid
type QuadDetector struct {
	config QuadConfig
	// hashes maps rectangle keys to their previous hash value.
	hashes map[rectKey]uint64
	// initialized tracks first-frame state.
	initialized bool
}

// rectKey is a compact representation of a rectangle for map lookups.
type rectKey struct {
	x0, y0, x1, y1 int
}

func toKey(r image.Rectangle) rectKey {
	return rectKey{r.Min.X, r.Min.Y, r.Max.X, r.Max.Y}
}

// NewQuadDetector creates a quadtree-based change detector.
func NewQuadDetector(config QuadConfig) *QuadDetector {
	return &QuadDetector{
		config: config,
		hashes: make(map[rectKey]uint64, 256),
	}
}

// hashRect computes xxHash64 of a rectangular region in-place.
func hashRect(img *image.RGBA, bounds image.Rectangle) uint64 {
	h := xxhash.New()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		start := img.PixOffset(bounds.Min.X, y)
		end := img.PixOffset(bounds.Max.X, y)
		h.Write(img.Pix[start:end])
	}
	return h.Sum64()
}

// DirtyRect is a changed rectangle found by the quadtree detector.
// Unlike fixed-grid DirtySection, these can be different sizes depending
// on how deep the recursion went.
type DirtyRect struct {
	Bounds image.Rectangle
	Depth  int // 0 = full frame, higher = more refined
}

// Update processes a new frame and returns dirty rectangles at varying
// granularities. Static regions are detected with minimal work.
func (q *QuadDetector) Update(img *image.RGBA) []DirtyRect {
	fullBounds := image.Rect(0, 0, q.config.FrameWidth, q.config.FrameHeight)

	if !q.initialized {
		// First frame: hash everything, return full frame as dirty.
		q.hashAll(img, fullBounds, 0)
		q.initialized = true
		return []DirtyRect{{Bounds: fullBounds, Depth: 0}}
	}

	var dirty []DirtyRect
	q.detect(img, fullBounds, 0, &dirty)
	return dirty
}

// detect recursively checks a region and subdivides if changed.
func (q *QuadDetector) detect(img *image.RGBA, bounds image.Rectangle, depth int, dirty *[]DirtyRect) {
	key := toKey(bounds)
	newHash := hashRect(img, bounds)

	oldHash, exists := q.hashes[key]
	if exists && newHash == oldHash {
		// This entire region is unchanged — prune.
		return
	}

	// Region changed. Can we subdivide further?
	w, h := bounds.Dx(), bounds.Dy()
	canSubdivide := w > q.config.MinTileSize && h > q.config.MinTileSize
	if q.config.MaxDepth > 0 && depth >= q.config.MaxDepth {
		canSubdivide = false
	}

	if !canSubdivide {
		// Leaf node: report as dirty, update hash.
		*dirty = append(*dirty, DirtyRect{Bounds: bounds, Depth: depth})
		q.hashes[key] = newHash
		return
	}

	// Subdivide into 4 quadrants.
	midX := bounds.Min.X + w/2
	midY := bounds.Min.Y + h/2

	quads := [4]image.Rectangle{
		image.Rect(bounds.Min.X, bounds.Min.Y, midX, midY), // top-left
		image.Rect(midX, bounds.Min.Y, bounds.Max.X, midY), // top-right
		image.Rect(bounds.Min.X, midY, midX, bounds.Max.Y), // bottom-left
		image.Rect(midX, midY, bounds.Max.X, bounds.Max.Y), // bottom-right
	}

	for _, quad := range quads {
		if quad.Dx() > 0 && quad.Dy() > 0 {
			q.detect(img, quad, depth+1, dirty)
		}
	}

	// Update hash for this level so future frames can prune here.
	q.hashes[key] = newHash
}

// hashAll recursively stores hashes for all levels of the quadtree.
func (q *QuadDetector) hashAll(img *image.RGBA, bounds image.Rectangle, depth int) {
	key := toKey(bounds)
	q.hashes[key] = hashRect(img, bounds)

	w, h := bounds.Dx(), bounds.Dy()
	canSubdivide := w > q.config.MinTileSize && h > q.config.MinTileSize
	if q.config.MaxDepth > 0 && depth >= q.config.MaxDepth {
		canSubdivide = false
	}

	if !canSubdivide {
		return
	}

	midX := bounds.Min.X + w/2
	midY := bounds.Min.Y + h/2
	quads := [4]image.Rectangle{
		image.Rect(bounds.Min.X, bounds.Min.Y, midX, midY),
		image.Rect(midX, bounds.Min.Y, bounds.Max.X, midY),
		image.Rect(bounds.Min.X, midY, midX, bounds.Max.Y),
		image.Rect(midX, midY, bounds.Max.X, bounds.Max.Y),
	}
	for _, quad := range quads {
		if quad.Dx() > 0 && quad.Dy() > 0 {
			q.hashAll(img, quad, depth+1)
		}
	}
}

// Reset clears the quadtree hash state.
func (q *QuadDetector) Reset() {
	q.initialized = false
	q.hashes = make(map[rectKey]uint64, 256)
}

// Config returns the detector's configuration.
func (q *QuadDetector) Config() QuadConfig {
	return q.config
}
