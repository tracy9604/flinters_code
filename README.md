# Ad Campaign CSV Aggregator

A small, dependency-free Go CLI that streams a large advertising CSV (~1GB) and
produces two ranked reports:

- `top10_ctr.csv` — the campaigns with the highest click-through rate (CTR)
- `top10_cpa.csv` — the campaigns with the lowest cost per acquisition (CPA)

The file is processed as a stream, so memory usage stays flat (~10 MiB peak)
regardless of input size.

## Input schema

The input CSV must contain a header row with these columns (order does not
matter; columns are resolved by name):

| Column        | Type    | Description                  |
| ------------- | ------- | ---------------------------- |
| `campaign_id` | string  | Campaign ID                  |
| `date`        | string  | Date `YYYY-MM-DD`            |
| `impressions` | integer | Number of impressions        |
| `clicks`      | integer | Number of clicks             |
| `spend`       | float   | Advertising cost (USD)       |
| `conversions` | integer | Number of conversions        |

## Aggregation

For each `campaign_id` the tool sums `impressions`, `clicks`, `spend`, and
`conversions`, then derives:

- `CTR = total_clicks / total_impressions`
- `CPA = total_spend / total_conversions`

Rules:

- CTR is undefined when `total_impressions = 0`; such campaigns are excluded
  from the CTR ranking.
- CPA is undefined when `total_conversions = 0`; such campaigns are excluded
  from the CPA ranking (per the task spec).
- Output: `CTR` is formatted to 4 decimal places, `CPA` and `total_spend` to 2.

## Setup

Requires [Go 1.24+](https://go.dev/dl/).

```bash
git clone <your-repo-url>
cd <repo>
go build -o aggregator .
```

No third-party dependencies are needed.

## Usage

```bash
./aggregator --input ad_data.csv --output results/
```

Or without building first:

```bash
go run . --input ad_data.csv --output results/
```

### Flags

| Flag       | Default     | Description                                          |
| ---------- | ----------- | ---------------------------------------------------- |
| `--input`  | (required)  | Path to the input CSV file                           |
| `--output` | `results`   | Directory for `top10_ctr.csv` and `top10_cpa.csv`    |
| `--top`    | `10`        | Number of campaigns per report                       |
| `--stats`  | `false`     | Print processing time and memory usage to stderr     |

The output directory is created if it does not exist.

## Output

Two comma-separated CSV files, each with the header:

```
campaign_id,total_impressions,total_clicks,total_spend,total_conversions,CTR,CPA
```

## Docker

```bash
docker build -t ad-aggregator .
docker run --rm -v "$PWD:/data" ad-aggregator --input /data/ad_data.csv --output /data/results
```

## Tests

```bash
go test ./...
```

Tests cover aggregation totals, CTR/CPA math, zero-division exclusions,
malformed-row skipping, column-order independence, ranking + tie-breaking, and
output formatting.

## Performance

Measured on Apple Silicon (darwin/arm64), Go 1.24, against the real input file
`testdata/ad_data.csv` (995 MiB, 26,843,544 rows, 50 distinct campaigns):

| Metric             | Value             |
| ------------------ | ----------------- |
| Processing time    | ~6.1 s            |
| Throughput         | ~4.4M rows/sec (~163 MiB/s) |
| Peak RSS           | ~10.2 MiB         |

The committed files in `results/` were produced from this real input. See
[benchmark/benchmark.md](benchmark/benchmark.md) for the full log.

To reproduce:

```bash
/usr/bin/time -l ./aggregator --input testdata/ad_data.csv --output results --stats   # macOS
/usr/bin/time -v ./aggregator --input testdata/ad_data.csv --output results --stats   # Linux
```

A synthetic file of any size can also be generated for testing:

```bash
go run scripts/generate_data.go --output ad_data.csv --size-gb 1
```

## Design decisions

- **Streaming, not loading.** The CSV is read row-by-row via `encoding/csv` over
  a 1 MiB buffered reader (`ReuseRecord = true` to avoid per-row allocation).
  Memory is `O(number of campaigns)`, not `O(file size)`, so a 1GB file uses
  ~10 MiB. The aggregated result set is tiny (hundreds/thousands of campaigns),
  so "top 10" is a cheap sort over that small set.
- **Spend as integer cents.** Summing millions of `float64` values accumulates
  rounding error. `spend` is parsed into `int64` cents (rounded to the nearest
  cent) and summed as integers, then converted to dollars only at output time.
  This keeps totals exact.
- **Deterministic ordering.** Ranking ties break on `campaign_id` ascending so
  output is reproducible across runs.
- **Resilient parsing.** Structurally malformed rows (wrong field count) and
  rows with unparseable/negative/empty values are skipped and counted rather
  than aborting the run. The skipped count is reported. A missing/unreadable
  input file or a missing required column is a hard error with a non-zero exit.
- **Columns resolved by name.** Header names (case-insensitive, trimmed) are
  mapped to indices, so the input column order is flexible.
- **Stdlib only.** No external dependencies keeps the build trivial, the binary
  small (~1.8 MB), and the supply chain clean.

## Libraries used

Standard library only: `encoding/csv`, `bufio`, `flag`, `sort`, `strconv`,
`math`, `os`, `path/filepath`, `runtime`, `time`.

## Project layout

```
.
├── main.go                       # CLI entry point
├── internal/
│   ├── model/                    # Campaign struct + CTR/CPA helpers
│   ├── aggregate/                # streaming CSV reader + aggregation
│   └── report/                   # ranking + CSV writing
├── scripts/generate_data.go      # synthetic data generator (benchmarking)
├── testdata/sample.csv           # small fixture
├── benchmark/                    # benchmark log + raw time output
├── results/                      # output CSVs
├── Dockerfile
├── PROMPTS.md
└── README.md
```
