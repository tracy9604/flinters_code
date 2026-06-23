// Package report ranks aggregated campaigns and writes the CSV result files.
package report

import (
	"encoding/csv"
	"io"
	"sort"
	"strconv"

	"github.com/haunt/ad-aggregator/internal/model"
)

// header is the column order shared by both output files.
var header = []string{
	"campaign_id",
	"total_impressions",
	"total_clicks",
	"total_spend",
	"total_conversions",
	"CTR",
	"CPA",
}

// TopByCTR returns up to n campaigns with the highest CTR. Campaigns with zero
// impressions (undefined CTR) are excluded. Ties break on campaign_id ascending
// for deterministic output.
func TopByCTR(campaigns map[string]*model.Campaign, n int) []*model.Campaign {
	filtered := make([]*model.Campaign, 0, len(campaigns))
	for _, c := range campaigns {
		if _, ok := c.CTR(); ok {
			filtered = append(filtered, c)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		ci, _ := filtered[i].CTR()
		cj, _ := filtered[j].CTR()
		if ci != cj {
			return ci > cj
		}
		return filtered[i].ID < filtered[j].ID
	})
	return head(filtered, n)
}

// TopByCPA returns up to n campaigns with the lowest CPA. Campaigns with zero
// conversions (undefined CPA) are excluded. Ties break on campaign_id ascending.
func TopByCPA(campaigns map[string]*model.Campaign, n int) []*model.Campaign {
	filtered := make([]*model.Campaign, 0, len(campaigns))
	for _, c := range campaigns {
		if _, ok := c.CPA(); ok {
			filtered = append(filtered, c)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		ci, _ := filtered[i].CPA()
		cj, _ := filtered[j].CPA()
		if ci != cj {
			return ci < cj
		}
		return filtered[i].ID < filtered[j].ID
	})
	return head(filtered, n)
}

// head returns the first n elements of s, or all of s if it is shorter.
func head(s []*model.Campaign, n int) []*model.Campaign {
	if n < 0 {
		n = 0
	}
	if len(s) > n {
		return s[:n]
	}
	return s
}

// Write emits the ranked campaigns as a comma-separated CSV with a header row.
// CTR is formatted to 4 decimals and CPA/total_spend to 2 decimals. An undefined
// metric is written as an empty field.
func Write(w io.Writer, rows []*model.Campaign) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}

	record := make([]string, len(header))
	for _, c := range rows {
		record[0] = c.ID
		record[1] = strconv.FormatInt(c.Impressions, 10)
		record[2] = strconv.FormatInt(c.Clicks, 10)
		record[3] = strconv.FormatFloat(c.SpendDollars(), 'f', 2, 64)
		record[4] = strconv.FormatInt(c.Conversions, 10)

		if ctr, ok := c.CTR(); ok {
			record[5] = strconv.FormatFloat(ctr, 'f', 4, 64)
		} else {
			record[5] = ""
		}
		if cpa, ok := c.CPA(); ok {
			record[6] = strconv.FormatFloat(cpa, 'f', 2, 64)
		} else {
			record[6] = ""
		}

		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
