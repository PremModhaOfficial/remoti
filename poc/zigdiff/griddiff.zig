// griddiff.zig — comptime-parameterized screen grid differ.
//
// All grid geometry (section counts, byte offsets, row strides) is resolved
// at compile time via the Grid comptime struct.  The exported C functions are
// thin wrappers so CGo can call them directly.
//
// Build as shared library:
//   zig build-lib -dynamic -OReleaseFast -femit-bin=libgriddiff.so griddiff.zig
//
// Build as benchmark binary:
//   zig build-exe -OReleaseFast -femit-bin=bench_zig griddiff.zig

const std = @import("std");
const builtin = @import("builtin");

// ---------------------------------------------------------------------------
// Comptime grid geometry
// ---------------------------------------------------------------------------

/// Grid comptime type: all arithmetic is resolved by the compiler.
pub fn Grid(
    comptime frame_w: u32,
    comptime frame_h: u32,
    comptime cols: u32,
    comptime rows: u32,
) type {
    comptime {
        if (frame_w % cols != 0) @compileError("frame_w must be divisible by cols");
        if (frame_h % rows != 0) @compileError("frame_h must be divisible by rows");
    }

    const sec_w: u32 = frame_w / cols;
    const sec_h: u32 = frame_h / rows;
    const stride: u32 = frame_w * 4; // RGBA stride in bytes
    const section_count: u32 = cols * rows;
    const section_bytes: u32 = sec_w * 4; // bytes per row-slice within a section

    // Pre-compute the starting byte offset for each section (top-left pixel).
    // This is a comptime [section_count]u32 array — zero runtime cost.
    const offsets: [section_count]u32 = blk: {
        var arr: [section_count]u32 = undefined;
        var idx: u32 = 0;
        while (idx < section_count) : (idx += 1) {
            const col = idx % cols;
            const row = idx / cols;
            arr[idx] = (row * sec_h * stride) + (col * sec_w * 4);
        }
        break :blk arr;
    };

    return struct {
        pub const FRAME_W = frame_w;
        pub const FRAME_H = frame_h;
        pub const COLS = cols;
        pub const ROWS = rows;
        pub const SEC_W = sec_w;
        pub const SEC_H = sec_h;
        pub const STRIDE = stride;
        pub const SECTION_COUNT = section_count;
        pub const SECTION_BYTES = section_bytes;
        pub const OFFSETS = offsets;

        /// Hash a single section using Zig's XxHash3 (SIMD @Vector(8,u64)).
        /// Uses comptime offset — no multiply at runtime for the base address.
        pub fn hashSection(pix: [*]const u8, section_idx: u32) u64 {
            const base = OFFSETS[section_idx];
            var h = std.hash.XxHash3.init(0);
            var row: u32 = 0;
            while (row < SEC_H) : (row += 1) {
                const off = base + row * STRIDE;
                h.update(pix[off .. off + SECTION_BYTES]);
            }
            return h.final();
        }

        /// Single-section hash with XxHash64 (for comparison).
        pub fn hashSection64(pix: [*]const u8, section_idx: u32) u64 {
            const base = OFFSETS[section_idx];
            var h = std.hash.XxHash64.init(0);
            var row: u32 = 0;
            while (row < SEC_H) : (row += 1) {
                const off = base + row * STRIDE;
                h.update(pix[off .. off + SECTION_BYTES]);
            }
            return h.final();
        }

        /// SIMD 32-byte equality check for a row slice starting at `off`.
        /// Returns true if ANY byte differs. Fully inlined and unrolled by compiler.
        pub inline fn rowDirty(a: [*]const u8, b: [*]const u8, off: u32) bool {
            const Vec32 = @Vector(32, u8);
            // SECTION_BYTES = 480 for 1080p (120px * 4 bytes). 480/32 = 15 chunks.
            comptime var chunk_off: u32 = 0;
            inline while (chunk_off + 32 <= SECTION_BYTES) : (chunk_off += 32) {
                const va: Vec32 = a[off + chunk_off ..][0..32].*;
                const vb: Vec32 = b[off + chunk_off ..][0..32].*;
                if (@reduce(.Or, va != vb)) return true;
            }
            // Scalar tail.
            comptime var tail: u32 = chunk_off;
            inline while (tail < SECTION_BYTES) : (tail += 1) {
                if (a[off + tail] != b[off + tail]) return true;
            }
            return false;
        }

        /// Hash-based diff. Runtime loop over sections — keeps code size small
        /// so the hot path stays in i-cache. The inner hash loop (per-row) is
        /// where the work is, not the section dispatch.
        pub fn diff(
            pix: [*]const u8,
            prev_hashes: *[SECTION_COUNT]u64,
            dirty_out: [*]u32,
        ) u32 {
            var dirty_count: u32 = 0;
            var i: u32 = 0;
            while (i < SECTION_COUNT) : (i += 1) {
                const h = hashSection(pix, i);
                if (h != prev_hashes[i]) {
                    dirty_out[dirty_count] = i;
                    dirty_count += 1;
                    prev_hashes[i] = h;
                }
            }
            return dirty_count;
        }

        /// SIMD equality diff — no hashing, compare buffers directly.
        /// Section loop is runtime (not inline) to keep code size sane.
        /// The per-row SIMD comparison is what matters for throughput.
        pub fn diffEq(
            pix_new: [*]const u8,
            pix_old: [*]const u8,
            dirty_out: [*]u32,
        ) u32 {
            var dirty_count: u32 = 0;
            var i: u32 = 0;
            while (i < SECTION_COUNT) : (i += 1) {
                const base = OFFSETS[i];
                var dirty = false;
                var row: u32 = 0;
                while (row < SEC_H) : (row += 1) {
                    const off = base + row * STRIDE;
                    if (rowDirty(pix_new, pix_old, off)) {
                        dirty = true;
                        break;
                    }
                }
                if (dirty) {
                    dirty_out[dirty_count] = i;
                    dirty_count += 1;
                }
            }
            return dirty_count;
        }
    };
}

// ---------------------------------------------------------------------------
// Concrete 1920x1080 16x9 instance
// ---------------------------------------------------------------------------

pub const Grid1080p = Grid(1920, 1080, 16, 9);

pub const DiffState = extern struct {
    hashes: [Grid1080p.SECTION_COUNT]u64,
};

// ---------------------------------------------------------------------------
// C-exported API (used by CGo)
// ---------------------------------------------------------------------------

export fn griddiff_state_size() usize {
    return @sizeOf(DiffState);
}

export fn griddiff_reset(state: *DiffState) void {
    @memset(&state.hashes, 0);
}

/// Hash-based diff. Returns number of dirty section indices written to dirty_out.
/// dirty_out must hold at least SECTION_COUNT (144) u32 slots.
/// On first_frame=true, primes the hash table and returns all 144 as dirty.
export fn griddiff_update(
    state: *DiffState,
    pix: [*]const u8,
    dirty_out: [*]u32,
    first_frame: bool,
) u32 {
    if (first_frame) {
        // Prime: hash all sections, mark all dirty.
        var i: u32 = 0;
        while (i < Grid1080p.SECTION_COUNT) : (i += 1) {
            state.hashes[i] = Grid1080p.hashSection(pix, i);
            dirty_out[i] = i;
        }
        return Grid1080p.SECTION_COUNT;
    }
    return Grid1080p.diff(pix, &state.hashes, dirty_out);
}

/// Direct SIMD equality diff. Caller manages prev_pix buffer externally.
export fn griddiff_eq(
    pix_new: [*]const u8,
    pix_old: [*]const u8,
    dirty_out: [*]u32,
) u32 {
    return Grid1080p.diffEq(pix_new, pix_old, dirty_out);
}

export fn griddiff_section_count() u32 {
    return Grid1080p.SECTION_COUNT;
}

export fn griddiff_sec_w() u32 {
    return Grid1080p.SEC_W;
}

export fn griddiff_sec_h() u32 {
    return Grid1080p.SEC_H;
}

// ---------------------------------------------------------------------------
// Standalone benchmark
// ---------------------------------------------------------------------------

pub fn main() !void {
    const args = try std.process.argsAlloc(std.heap.page_allocator);
    defer std.process.argsFree(std.heap.page_allocator, args);

    const do_bench = args.len > 1 and std.mem.eql(u8, args[1], "bench");

    const frame_bytes = Grid1080p.FRAME_W * Grid1080p.FRAME_H * 4;
    const pix_a = try std.heap.page_allocator.alloc(u8, frame_bytes);
    const pix_b = try std.heap.page_allocator.alloc(u8, frame_bytes);
    defer std.heap.page_allocator.free(pix_a);
    defer std.heap.page_allocator.free(pix_b);

    var prng = std.Random.DefaultPrng.init(42);
    prng.fill(pix_a);
    @memcpy(pix_b, pix_a);

    // Flip one pixel in section (col=3, row=2) → index = 2*16+3 = 35.
    const flip_off = 300 * Grid1080p.STRIDE + 400 * 4;
    pix_b[flip_off] ^= 0xFF;

    var state: DiffState = undefined;
    griddiff_reset(&state);
    var dirty_buf: [Grid1080p.SECTION_COUNT]u32 = undefined;

    // Correctness check (always).
    _ = griddiff_update(&state, pix_a.ptr, &dirty_buf, true); // prime
    const n_hash = griddiff_update(&state, pix_b.ptr, &dirty_buf, false);
    const n_eq = griddiff_eq(pix_b.ptr, pix_a.ptr, &dirty_buf);

    const stdout = std.io.getStdOut().writer();
    try stdout.print("Correctness: hash_dirty={d} eq_dirty={d} (expected 1 each)\n", .{ n_hash, n_eq });
    if (n_hash >= 1) {
        try stdout.print("  Hash dirty[0]={d} (expected 35)\n", .{dirty_buf[0]});
    }

    if (!do_bench) return;

    const iterations: u64 = 5_000;

    // --- Bench 1: steady-state hash diff (identical frames, 0 dirty) ---
    // Prime once, then repeatedly diff pix_a vs itself (no changes).
    griddiff_reset(&state);
    _ = griddiff_update(&state, pix_a.ptr, &dirty_buf, true);
    var t0 = try std.time.Timer.start();
    var i: u64 = 0;
    while (i < iterations) : (i += 1) {
        _ = griddiff_update(&state, pix_a.ptr, &dirty_buf, false);
    }
    const hash_static_ns = t0.read();

    // --- Bench 2: steady-state hash diff (1 dirty section alternating) ---
    // Alternate between pix_a and pix_b so section 35 is always dirty.
    // Each iteration is exactly one griddiff_update call.
    griddiff_reset(&state);
    _ = griddiff_update(&state, pix_a.ptr, &dirty_buf, true);
    t0 = try std.time.Timer.start();
    i = 0;
    while (i < iterations) : (i += 1) {
        if (i % 2 == 0) {
            _ = griddiff_update(&state, pix_b.ptr, &dirty_buf, false);
        } else {
            _ = griddiff_update(&state, pix_a.ptr, &dirty_buf, false);
        }
    }
    const hash_1dirty_ns = t0.read();

    // --- Bench 3: eq diff static (identical frames) ---
    t0 = try std.time.Timer.start();
    i = 0;
    while (i < iterations) : (i += 1) {
        _ = griddiff_eq(pix_a.ptr, pix_a.ptr, &dirty_buf);
    }
    const eq_static_ns = t0.read();

    // --- Bench 4: eq diff with 1 dirty section ---
    t0 = try std.time.Timer.start();
    i = 0;
    while (i < iterations) : (i += 1) {
        _ = griddiff_eq(pix_b.ptr, pix_a.ptr, &dirty_buf);
    }
    const eq_1dirty_ns = t0.read();

    const ns_per = struct {
        fn fmt(total_ns: u64, iters: u64) f64 {
            return @as(f64, @floatFromInt(total_ns)) / @as(f64, @floatFromInt(iters)) / 1000.0;
        }
    }.fmt;

    try stdout.print("\n=== Zig benchmark ({d} iterations) ===\n", .{iterations});
    try stdout.print("Hash diff (0 dirty / static):  {d:.2}µs/op\n", .{ns_per(hash_static_ns, iterations)});
    try stdout.print("Hash diff (1 dirty, full hash): {d:.2}µs/op  [reset+prime+diff]\n", .{ns_per(hash_1dirty_ns, iterations)});
    try stdout.print("Eq diff   (0 dirty / static):  {d:.2}µs/op\n", .{ns_per(eq_static_ns, iterations)});
    try stdout.print("Eq diff   (1 dirty section):   {d:.2}µs/op\n", .{ns_per(eq_1dirty_ns, iterations)});
    try stdout.print("\nGo baseline (from Go bench): ~500µs/op (hash, 16x9 1080p, 9 goroutines)\n", .{});

    // --- Bench 5: single section hash (XxHash3) ---
    t0 = try std.time.Timer.start();
    i = 0;
    var sink: u64 = 0;
    while (i < iterations * 100) : (i += 1) {
        sink +%= Grid1080p.hashSection(pix_a.ptr, 0);
    }
    const sec_hash3_ns = t0.read();

    // --- Bench 6: single section hash (XxHash64) ---
    t0 = try std.time.Timer.start();
    i = 0;
    var sink2: u64 = 0;
    while (i < iterations * 100) : (i += 1) {
        sink2 +%= Grid1080p.hashSection64(pix_a.ptr, 0);
    }
    const sec_hash64_ns = t0.read();

    const sec_iters = iterations * 100;
    try stdout.print("\nPer-section hash (XxHash3):  {d:.3}µs  sink={d}  [Go BenchmarkHashSection ~7µs]\n", .{ ns_per(sec_hash3_ns, sec_iters), sink });
    try stdout.print("Per-section hash (XxHash64): {d:.3}µs  sink={d}\n", .{ ns_per(sec_hash64_ns, sec_iters), sink2 });
}
