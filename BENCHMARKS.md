# json2xml Benchmark Results

Comprehensive performance comparison between all json2xml implementations.

## Test Environment

- **Machine**: Apple Silicon (M-series, aarch64)
- **OS**: macOS
- **Date**: January 28, 2026

### Implementations Tested

| Implementation | Type | Notes |
|----------------|------|-------|
| Python | Library | Pure Python (json2xml) |
| Rust | Library | Native extension via PyO3 (json2xml-rs) |
| Go | CLI | Standalone binary (json2xml-go v1.0.0) |
| Zig | CLI | Standalone binary (json2xml-zig) |

## Test Data

| Size | Description | Bytes |
|------|-------------|-------|
| Small | Simple object `{"name": "John", "age": 30, "city": "New York"}` | 47 |
| Medium | 10 generated records with nested structures | ~3,208 |
| bigexample.json | Real-world patent data | 2,018 |
| Large | 100 generated records with nested structures | ~32,205 |
| Very Large | 1,000 generated records with nested structures | ~323,119 |

## Results

### Performance Summary

| Test Case | Python | Rust | Go | Zig |
|-----------|--------|------|-----|-----|
| Small (47B) | 41.88µs | 1.66µs | 4.52ms | 2.80ms |
| Medium (3.2KB) | 2.19ms | 71.85µs | 4.33ms | 2.18ms |
| bigexample (2KB) | 854.38µs | 30.89µs | 4.28ms | 2.12ms |
| Large (32KB) | 21.57ms | 672.96µs | 4.47ms | 2.48ms |
| Very Large (323KB) | 216.52ms | 6.15ms | 4.44ms | 5.54ms |

### Speedup vs Pure Python

| Test Case | Rust | Go | Zig |
|-----------|------|-----|-----|
| Small (47B) | **25.2x** | 0.0x* | 0.0x* |
| Medium (3.2KB) | **30.5x** | 0.5x* | 1.0x* |
| bigexample (2KB) | **27.7x** | 0.2x* | 0.4x* |
| Large (32KB) | **32.1x** | 4.8x | **8.7x** |
| Very Large (323KB) | **35.2x** | **48.8x** | **39.1x** |

*CLI tools have process spawn overhead (~2-4ms) which dominates for small inputs

## Key Observations

### 1. Go Excels for Very Large CLI Workloads 🚀

For very large inputs (323KB+):
- **48.8x faster** than Python
- Consistent ~4ms overhead regardless of input size
- Best throughput for batch processing large files

### 2. Rust Extension is Best for Python Users 🦀

The Rust extension (json2xml-rs) provides:
- **~25-35x faster** than pure Python consistently across all input sizes
- **Zero process overhead** - called directly from Python
- **Easy install**: `pip install json2xml[fast]`

### 3. Zig is Highly Competitive ⚡

After recent optimizations:
- **39.1x faster** than Python for very large files
- **Faster startup** than Go (~2ms vs ~4ms)
- Best for small-to-medium CLI workloads

### 4. Process Spawn Overhead Matters

CLI tools (Go, Zig) have process spawn overhead:
- Go: ~4ms startup overhead
- Zig: ~2ms startup overhead
- Dominates for small inputs
- Negligible for large inputs where actual work dominates

## When to Use Each Implementation

| Use Case | Recommended | Why |
|----------|-------------|-----|
| Python library calls | **Rust** (`pip install json2xml[fast]`) | 25-35x faster, no overhead |
| Small files via CLI | **Zig** | Fastest startup (~2ms) |
| Very large files via CLI | **Go** | Best throughput (48.8x faster) |
| Batch processing | **Go** or **Rust** | Both excellent |
| Pure Python required | **Python** (json2xml) | Always available |

## Installation

```bash
# Go CLI (this repo)
go install github.com/vinitkumar/json2xml-go@latest

# Or build from source
go build -o json2xml-go ./cmd/json2xml-go

# Python with Rust acceleration
pip install json2xml[fast]

# Zig CLI
# See: github.com/vinitkumar/json2xml-zig
```

## Running the Benchmarks

```bash
# From the Python json2xml repo
cd ~/projects/python/json2xml
python benchmark_all.py
```

## Related Projects

- **Python + Rust**: [github.com/vinitkumar/json2xml](https://github.com/vinitkumar/json2xml)
- **Zig version**: [github.com/vinitkumar/json2xml-zig](https://github.com/vinitkumar/json2xml-zig)
