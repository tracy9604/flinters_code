//go:build ignore

// generate_data.go produces a synthetic advertising CSV for benchmarking.
//
// Usage:
//
//	go run scripts/generate_data.go --output ad_data.csv --size-gb 1
//	go run scripts/generate_data.go --output ad_data.csv --rows 20000000
//
// --rows takes precedence over --size-gb when both are provided.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
)

func main() {
	output := flag.String("output", "ad_data.csv", "path to write the generated CSV")
	sizeGB := flag.Float64("size-gb", 1.0, "approximate target file size in gigabytes")
	rows := flag.Int64("rows", 0, "exact number of data rows (overrides --size-gb)")
	campaigns := flag.Int("campaigns", 1000, "number of distinct campaign IDs")
	seed := flag.Int64("seed", 42, "random seed for reproducible data")
	flag.Parse()

	f, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, 1<<20)
	defer w.Flush()

	if _, err := w.WriteString("campaign_id,date,impressions,clicks,spend,conversions\n"); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	rng := rand.New(rand.NewSource(*seed))

	// Roughly 45 bytes per data row; used to translate a size target into rows.
	const approxBytesPerRow = 45
	targetRows := *rows
	if targetRows <= 0 {
		targetRows = int64(*sizeGB * (1 << 30) / approxBytesPerRow)
	}

	dates := []string{
		"2025-01-01", "2025-01-02", "2025-01-03", "2025-01-04", "2025-01-05",
		"2025-01-06", "2025-01-07", "2025-01-08", "2025-01-09", "2025-01-10",
	}

	for i := int64(0); i < targetRows; i++ {
		id := fmt.Sprintf("CMP%05d", rng.Intn(*campaigns))
		date := dates[rng.Intn(len(dates))]
		impressions := rng.Intn(50000) + 1
		clicks := rng.Intn(impressions + 1)
		spendCents := rng.Intn(100000)
		conversions := rng.Intn(clicks + 1)

		fmt.Fprintf(w, "%s,%s,%d,%d,%d.%02d,%d\n",
			id, date, impressions, clicks, spendCents/100, spendCents%100, conversions)
	}

	fmt.Printf("Wrote %d rows to %s\n", targetRows, *output)
}
