// Command aggregator streams an advertising CSV file and writes two reports:
// the top campaigns by click-through rate and the top campaigns by lowest cost
// per acquisition.
//
// Usage:
//
//	aggregator --input ad_data.csv --output results/ [--top 10] [--stats]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/haunt/ad-aggregator/internal/aggregate"
	"github.com/haunt/ad-aggregator/internal/model"
	"github.com/haunt/ad-aggregator/internal/report"
)

const (
	ctrFileName = "top10_ctr.csv"
	cpaFileName = "top10_cpa.csv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	input := flag.String("input", "", "path to the input CSV file (required)")
	output := flag.String("output", "results", "directory to write result CSV files")
	top := flag.Int("top", 10, "number of campaigns to include in each report")
	stats := flag.Bool("stats", false, "print processing time and memory usage to stderr")
	flag.Parse()

	if *input == "" {
		return fmt.Errorf("--input is required")
	}
	if *top <= 0 {
		return fmt.Errorf("--top must be a positive integer, got %d", *top)
	}

	start := time.Now()

	f, err := os.Open(*input)
	if err != nil {
		return fmt.Errorf("opening input: %w", err)
	}
	defer f.Close()

	res, err := aggregate.Stream(f)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(*output, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	ctrRows := report.TopByCTR(res.Campaigns, *top)
	cpaRows := report.TopByCPA(res.Campaigns, *top)

	if err := writeReport(filepath.Join(*output, ctrFileName), ctrRows); err != nil {
		return err
	}
	if err := writeReport(filepath.Join(*output, cpaFileName), cpaRows); err != nil {
		return err
	}

	fmt.Printf("Processed %d rows (%d skipped) across %d campaigns.\n",
		res.RowsTotal, res.RowsSkipped, len(res.Campaigns))
	fmt.Printf("Wrote %s and %s to %s\n", ctrFileName, cpaFileName, *output)

	if *stats {
		printStats(start, res)
	}
	return nil
}

// writeReport writes the ranked rows to a CSV file at path.
func writeReport(path string, rows []*model.Campaign) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	if err := report.Write(f, rows); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return f.Close()
}

// printStats reports elapsed time and Go runtime memory usage to stderr.
func printStats(start time.Time, res *aggregate.Result) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	elapsed := time.Since(start)

	fmt.Fprintf(os.Stderr, "\n--- stats ---\n")
	fmt.Fprintf(os.Stderr, "elapsed:        %s\n", elapsed.Round(time.Millisecond))
	if elapsed.Seconds() > 0 {
		fmt.Fprintf(os.Stderr, "throughput:     %.0f rows/sec\n", float64(res.RowsTotal)/elapsed.Seconds())
	}
	fmt.Fprintf(os.Stderr, "campaigns:      %d\n", len(res.Campaigns))
	fmt.Fprintf(os.Stderr, "heap in use:    %.2f MiB\n", float64(m.HeapAlloc)/(1<<20))
	fmt.Fprintf(os.Stderr, "total alloc:    %.2f MiB\n", float64(m.TotalAlloc)/(1<<20))
	fmt.Fprintf(os.Stderr, "sys reserved:   %.2f MiB\n", float64(m.Sys)/(1<<20))
	fmt.Fprintf(os.Stderr, "Note: for true peak RSS run: /usr/bin/time -l ./aggregator ...\n")
}
