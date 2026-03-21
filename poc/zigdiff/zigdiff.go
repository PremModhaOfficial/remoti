// Package zigdiff is a CGo bridge to the Zig grid-differ shared library.
//
// The Zig library (libgriddiff.so) must be built and present in the same
// directory before this package can be compiled:
//
//	cd poc/zigdiff
//	zig build-lib -dynamic -OReleaseFast -femit-bin=libgriddiff.so griddiff.zig
//
// Then build with CGo:
//
//	CGO_LDFLAGS="-L/path/to/zigdiff -lgriddiff -Wl,-rpath,/path/to/zigdiff" \
//	  go build remoti/poc/zigdiff
//
// Alternatively, copy libgriddiff.so to /usr/local/lib and run ldconfig.

package zigdiff

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR} -lgriddiff -Wl,-rpath,${SRCDIR}
#include "griddiff.h"
#include <stdlib.h>
*/
import "C"
import (
	"image"
	"unsafe"
)

const SectionCount = 144

// State holds the CGo-managed DiffState for a single Detector instance.
type State struct {
	raw *C.DiffState
}

// NewState allocates a fresh DiffState (all hashes zeroed).
func NewState() *State {
	s := &State{
		raw: (*C.DiffState)(C.malloc(C.size_t(C.griddiff_state_size()))),
	}
	C.griddiff_reset(s.raw)
	return s
}

// Free releases the CGo-allocated memory. Call when done.
func (s *State) Free() {
	C.free(unsafe.Pointer(s.raw))
	s.raw = nil
}

// Reset zeroes all hashes, forcing the next Update to report all sections dirty.
func (s *State) Reset() {
	C.griddiff_reset(s.raw)
}

// Update diffs img against the stored hashes.
// firstFrame=true primes the table (all 144 sections returned as dirty).
// Returns the indices of dirty sections (0..143).
func (s *State) Update(img *image.RGBA, firstFrame bool) []uint32 {
	var dirtyBuf [SectionCount]C.uint32_t
	n := C.griddiff_update(
		s.raw,
		(*C.uint8_t)(unsafe.Pointer(&img.Pix[0])),
		&dirtyBuf[0],
		C.bool(firstFrame),
	)
	result := make([]uint32, int(n))
	for i := range result {
		result[i] = uint32(dirtyBuf[i])
	}
	return result
}

// UpdateEq diffs imgNew against prevPix using SIMD byte comparison (no hashing).
// prevPix must be the same length as imgNew.Pix (1920*1080*4 = 8,294,400 bytes).
// The caller is responsible for copying imgNew.Pix into prevPix after each frame.
func UpdateEq(imgNew *image.RGBA, prevPix []byte) []uint32 {
	var dirtyBuf [SectionCount]C.uint32_t
	n := C.griddiff_eq(
		(*C.uint8_t)(unsafe.Pointer(&imgNew.Pix[0])),
		(*C.uint8_t)(unsafe.Pointer(&prevPix[0])),
		&dirtyBuf[0],
	)
	result := make([]uint32, int(n))
	for i := range result {
		result[i] = uint32(dirtyBuf[i])
	}
	return result
}

// SectionBounds returns the pixel rectangle for a section index (0..143)
// for a 1920x1080 16x9 grid.
func SectionBounds(idx uint32) image.Rectangle {
	secW := int(C.griddiff_sec_w())
	secH := int(C.griddiff_sec_h())
	cols := 16
	col := int(idx) % cols
	row := int(idx) / cols
	x0 := col * secW
	y0 := row * secH
	return image.Rect(x0, y0, x0+secW, y0+secH)
}
