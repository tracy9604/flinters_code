# Benchmark

## Environment

- Machine: Apple Silicon (darwin/arm64)
- Go: go1.24
- Build: `go build -ldflags="-s -w" -o aggregator .`

## Input

The real assignment dataset, `testdata/ad_data.csv`:

- File size: 995 MiB
- Rows: 26,843,544 data rows
- Distinct campaigns: 50

## Command

```
/usr/bin/time -l ./aggregator --input testdata/ad_data.csv --output results --stats
```

## Results

Median of several warm runs:

| Metric                      | Value                    |
| --------------------------- | ------------------------ |
| Wall-clock time             | ~6.1 s                   |
| Throughput                  | ~4.4 M rows/sec          |
| Throughput (bytes)          | ~163 MiB/sec             |
| Peak RSS (`time -l`)        | 10,682,368 B (~10.2 MiB) |
| Go heap in use (end)        | ~3.1 MiB                 |
| Go total alloc (cumulative) | ~956 MiB                 |
| Rows skipped                | 0                        |

Peak memory is flat (~10 MiB) and independent of input size because the file is
streamed row-by-row and only the per-campaign aggregates (50 small structs) are
held in memory. The ~956 MiB cumulative allocation is transient: `encoding/csv`
allocates one backing string per record, which is freed by the GC immediately
after the row is folded into the totals (end-of-run heap is only ~3 MiB).

## Raw `/usr/bin/time -l` output

See [run_time.txt](run_time.txt).
