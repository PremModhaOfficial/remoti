// griddiff.h — C interface for the Zig griddiff shared library.
// Generated manually to match the export fn declarations in griddiff.zig.
#ifndef GRIDDIFF_H
#define GRIDDIFF_H

#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

// Opaque diff state — allocate sizeof(DiffState) bytes and pass a pointer.
// Use griddiff_state_size() to get the required allocation size.
typedef struct {
    uint64_t hashes[144]; // one hash per section (Grid1080p.SECTION_COUNT)
} DiffState;

// Returns the byte size of DiffState (for dynamic allocation from Go).
size_t griddiff_state_size(void);

// Reset all stored hashes to zero (forces next diff to report all 144 dirty).
void griddiff_reset(DiffState *state);

// Hash-based diff. Compares pix against stored hashes in state.
// On first_frame=true: primes the hash table, returns 144 (all sections dirty).
// On first_frame=false: returns count of changed sections, indices in dirty_out.
// dirty_out must hold at least 144 uint32_t slots.
uint32_t griddiff_update(DiffState *state, const uint8_t *pix,
                         uint32_t *dirty_out, bool first_frame);

// SIMD equality diff — no hashing, compare pix_new vs pix_old directly.
// Returns dirty section count. Caller must maintain pix_old buffer.
uint32_t griddiff_eq(const uint8_t *pix_new, const uint8_t *pix_old,
                     uint32_t *dirty_out);

uint32_t griddiff_section_count(void);
uint32_t griddiff_sec_w(void);
uint32_t griddiff_sec_h(void);

#endif // GRIDDIFF_H
