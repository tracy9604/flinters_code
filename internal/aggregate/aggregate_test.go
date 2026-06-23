package aggregate

import (
	"strings"
	"testing"
)

const sampleCSV = `campaign_id,date,impressions,clicks,spend,conversions
CMP001,2025-01-01,12000,300,45.50,12
CMP002,2025-01-01,8000,120,28.00,4
CMP001,2025-01-02,14000,340,48.20,15
CMP003,2025-01-01,5000,60,15.00,3
CMP002,2025-01-02,8500,150,31.00,5
`

func TestStreamAggregates(t *testing.T) {
	res, err := Stream(strings.NewReader(sampleCSV))
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	if res.RowsTotal != 5 {
		t.Errorf("RowsTotal = %d, want 5", res.RowsTotal)
	}
	if res.RowsSkipped != 0 {
		t.Errorf("RowsSkipped = %d, want 0", res.RowsSkipped)
	}
	if len(res.Campaigns) != 3 {
		t.Fatalf("len(Campaigns) = %d, want 3", len(res.Campaigns))
	}

	c1 := res.Campaigns["CMP001"]
	if c1.Impressions != 26000 || c1.Clicks != 640 || c1.Conversions != 27 || c1.SpendCents != 9370 {
		t.Errorf("CMP001 = %+v, want impressions 26000, clicks 640, conversions 27, spendCents 9370", c1)
	}

	c2 := res.Campaigns["CMP002"]
	if c2.Impressions != 16500 || c2.Clicks != 270 || c2.Conversions != 9 || c2.SpendCents != 5900 {
		t.Errorf("CMP002 = %+v", c2)
	}
}

func TestStreamSkipsMalformedRows(t *testing.T) {
	const data = `campaign_id,date,impressions,clicks,spend,conversions
CMP001,2025-01-01,12000,300,45.50,12
CMP002,2025-01-01,notanumber,120,28.00,4
CMP003,2025-01-01,5000,60,15.00
,2025-01-01,1000,10,1.00,1
CMP004,2025-01-01,5000,60,15.00,3
`
	res, err := Stream(strings.NewReader(data))
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	// 4 data rows after header: bad-int, short-row, empty-id, plus 2 valid.
	if res.RowsSkipped != 3 {
		t.Errorf("RowsSkipped = %d, want 3", res.RowsSkipped)
	}
	if len(res.Campaigns) != 2 {
		t.Errorf("len(Campaigns) = %d, want 2 (CMP001, CMP004)", len(res.Campaigns))
	}
}

func TestStreamMissingColumn(t *testing.T) {
	const data = `campaign_id,date,impressions,clicks,spend
CMP001,2025-01-01,12000,300,45.50
`
	if _, err := Stream(strings.NewReader(data)); err == nil {
		t.Fatal("Stream() error = nil, want missing column error")
	}
}

func TestStreamEmptyInput(t *testing.T) {
	if _, err := Stream(strings.NewReader("")); err == nil {
		t.Fatal("Stream() error = nil, want empty input error")
	}
}

func TestColumnOrderIndependence(t *testing.T) {
	const data = `conversions,spend,clicks,impressions,date,campaign_id
12,45.50,300,12000,2025-01-01,CMP001
`
	res, err := Stream(strings.NewReader(data))
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	c := res.Campaigns["CMP001"]
	if c == nil || c.Impressions != 12000 || c.Clicks != 300 || c.Conversions != 12 || c.SpendCents != 4550 {
		t.Errorf("CMP001 = %+v, columns not resolved by name", c)
	}
}

func TestParseCents(t *testing.T) {
	cases := map[string]int64{
		"45.50": 4550,
		"28.00": 2800,
		"48.20": 4820,
		"15":    1500,
		"0.01":  1,
		"100.5": 10050,
	}
	for in, want := range cases {
		got, err := parseCents(in)
		if err != nil {
			t.Errorf("parseCents(%q) error = %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parseCents(%q) = %d, want %d", in, got, want)
		}
	}
}
