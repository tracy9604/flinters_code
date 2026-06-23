// Package aggregate streams a campaign CSV and accumulates per-campaign totals.
//
// The file is read row-by-row so memory stays proportional to the number of
// distinct campaign_id values, not the size of the input file.
package aggregate

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/haunt/ad-aggregator/internal/model"
)

// readBufferSize is the size of the buffered reader wrapping the input. A large
// buffer reduces the number of syscalls when scanning a multi-gigabyte file.
const readBufferSize = 1 << 20 // 1 MiB

// bufioReader wraps r in a large buffered reader to minimize syscalls.
func bufioReader(r io.Reader) io.Reader {
	return bufio.NewReaderSize(r, readBufferSize)
}

// requiredColumns are the headers the input CSV must contain. Lookups are done
// by name so the column order in the file does not matter.
var requiredColumns = []string{
	"campaign_id",
	"date",
	"impressions",
	"clicks",
	"spend",
	"conversions",
}

// Result is the outcome of streaming and aggregating an input file.
type Result struct {
	// Campaigns maps campaign_id to its accumulated totals.
	Campaigns map[string]*model.Campaign
	// RowsTotal is the number of data rows read (excluding the header).
	RowsTotal int64
	// RowsSkipped is the number of malformed data rows that were ignored.
	RowsSkipped int64
}

// Stream reads CSV data from r, aggregating totals per campaign_id. Malformed
// rows are skipped and counted rather than aborting the whole run. A fatal I/O
// or header error is returned.
func Stream(r io.Reader) (*Result, error) {
	reader := csv.NewReader(bufioReader(r))
	reader.ReuseRecord = true // reuse the backing slice to avoid per-row allocation

	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("input is empty: no header row found")
		}
		return nil, fmt.Errorf("reading header: %w", err)
	}

	cols, err := columnIndex(header)
	if err != nil {
		return nil, err
	}

	res := &Result{Campaigns: make(map[string]*model.Campaign)}

	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			var parseErr *csv.ParseError
			if errors.As(err, &parseErr) {
				// Structurally malformed row (e.g. wrong field count); skip it.
				res.RowsTotal++
				res.RowsSkipped++
				continue
			}
			return nil, fmt.Errorf("reading row %d: %w", res.RowsTotal+1, err)
		}

		res.RowsTotal++
		if err := accumulate(res.Campaigns, record, cols); err != nil {
			res.RowsSkipped++
		}
	}

	return res, nil
}

// columns holds the resolved position of each required field in a row.
type columns struct {
	campaignID  int
	impressions int
	clicks      int
	spend       int
	conversions int
}

// columnIndex resolves required column positions from the header row.
func columnIndex(header []string) (columns, error) {
	pos := make(map[string]int, len(header))
	for i, name := range header {
		pos[strings.ToLower(strings.TrimSpace(name))] = i
	}

	for _, name := range requiredColumns {
		if _, ok := pos[name]; !ok {
			return columns{}, fmt.Errorf("missing required column %q in header", name)
		}
	}

	return columns{
		campaignID:  pos["campaign_id"],
		impressions: pos["impressions"],
		clicks:      pos["clicks"],
		spend:       pos["spend"],
		conversions: pos["conversions"],
	}, nil
}

// accumulate parses a single record and folds it into the campaign totals.
// It returns an error if the row cannot be parsed, leaving totals untouched.
func accumulate(campaigns map[string]*model.Campaign, record []string, cols columns) error {
	id := strings.TrimSpace(record[cols.campaignID])
	if id == "" {
		return errors.New("empty campaign_id")
	}

	impressions, err := parseInt(record[cols.impressions])
	if err != nil {
		return err
	}
	clicks, err := parseInt(record[cols.clicks])
	if err != nil {
		return err
	}
	conversions, err := parseInt(record[cols.conversions])
	if err != nil {
		return err
	}
	spendCents, err := parseCents(record[cols.spend])
	if err != nil {
		return err
	}

	c := campaigns[id]
	if c == nil {
		c = &model.Campaign{ID: id}
		campaigns[id] = c
	}
	c.Impressions += impressions
	c.Clicks += clicks
	c.Conversions += conversions
	c.SpendCents += spendCents
	return nil
}

// parseInt parses a trimmed non-negative integer field.
func parseInt(s string) (int64, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", s, err)
	}
	if v < 0 {
		return 0, fmt.Errorf("negative value not allowed: %q", s)
	}
	return v, nil
}

// parseCents parses a dollar amount string into integer cents, rounding to the
// nearest cent. This keeps summed spend exact regardless of float representation.
func parseCents(s string) (int64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid spend %q: %w", s, err)
	}
	if f < 0 {
		return 0, fmt.Errorf("negative spend not allowed: %q", s)
	}
	return int64(math.Round(f * 100)), nil
}
