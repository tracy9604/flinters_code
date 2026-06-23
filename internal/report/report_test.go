package report

import (
	"strings"
	"testing"

	"github.com/haunt/ad-aggregator/internal/model"
)

func sampleCampaigns() map[string]*model.Campaign {
	return map[string]*model.Campaign{
		"CMP001": {ID: "CMP001", Impressions: 26000, Clicks: 640, SpendCents: 9370, Conversions: 27},
		"CMP002": {ID: "CMP002", Impressions: 16500, Clicks: 270, SpendCents: 5900, Conversions: 9},
		"CMP003": {ID: "CMP003", Impressions: 5000, Clicks: 60, SpendCents: 1500, Conversions: 3},
	}
}

func ids(rows []*model.Campaign) []string {
	out := make([]string, len(rows))
	for i, c := range rows {
		out[i] = c.ID
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTopByCTR(t *testing.T) {
	got := ids(TopByCTR(sampleCampaigns(), 10))
	want := []string{"CMP001", "CMP002", "CMP003"}
	if !equal(got, want) {
		t.Errorf("TopByCTR order = %v, want %v", got, want)
	}
}

func TestTopByCPA(t *testing.T) {
	got := ids(TopByCPA(sampleCampaigns(), 10))
	want := []string{"CMP001", "CMP003", "CMP002"}
	if !equal(got, want) {
		t.Errorf("TopByCPA order = %v, want %v", got, want)
	}
}

func TestTopLimit(t *testing.T) {
	got := TopByCTR(sampleCampaigns(), 2)
	if len(got) != 2 {
		t.Errorf("TopByCTR(n=2) returned %d rows, want 2", len(got))
	}
}

func TestTopByCTRExcludesZeroImpressions(t *testing.T) {
	campaigns := sampleCampaigns()
	campaigns["CMP004"] = &model.Campaign{ID: "CMP004", Impressions: 0, Clicks: 10, SpendCents: 100, Conversions: 1}
	for _, c := range TopByCTR(campaigns, 10) {
		if c.ID == "CMP004" {
			t.Fatal("TopByCTR included a zero-impression campaign")
		}
	}
}

func TestTopByCPAExcludesZeroConversions(t *testing.T) {
	campaigns := sampleCampaigns()
	campaigns["CMP004"] = &model.Campaign{ID: "CMP004", Impressions: 100, Clicks: 10, SpendCents: 100, Conversions: 0}
	for _, c := range TopByCPA(campaigns, 10) {
		if c.ID == "CMP004" {
			t.Fatal("TopByCPA included a zero-conversion campaign")
		}
	}
}

func TestTieBreakByID(t *testing.T) {
	// Two campaigns with identical CTR should be ordered by ID ascending.
	campaigns := map[string]*model.Campaign{
		"CMP_B": {ID: "CMP_B", Impressions: 1000, Clicks: 100, SpendCents: 1000, Conversions: 10},
		"CMP_A": {ID: "CMP_A", Impressions: 2000, Clicks: 200, SpendCents: 2000, Conversions: 20},
	}
	got := ids(TopByCTR(campaigns, 10))
	want := []string{"CMP_A", "CMP_B"}
	if !equal(got, want) {
		t.Errorf("tie-break order = %v, want %v", got, want)
	}
}

func TestWriteFormat(t *testing.T) {
	rows := []*model.Campaign{
		{ID: "CMP003", Impressions: 5000, Clicks: 60, SpendCents: 1500, Conversions: 3},
	}
	var sb strings.Builder
	if err := Write(&sb, rows); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got := sb.String()

	wantHeader := "campaign_id,total_impressions,total_clicks,total_spend,total_conversions,CTR,CPA"
	if !strings.Contains(got, wantHeader) {
		t.Errorf("output missing header.\ngot:\n%s", got)
	}
	// CTR = 60/5000 = 0.0120, CPA = 15.00/3 = 5.00, spend = 15.00
	wantRow := "CMP003,5000,60,15.00,3,0.0120,5.00"
	if !strings.Contains(got, wantRow) {
		t.Errorf("output missing expected row %q.\ngot:\n%s", wantRow, got)
	}
}

func TestWriteEmptyMetrics(t *testing.T) {
	rows := []*model.Campaign{
		{ID: "CMP000", Impressions: 0, Clicks: 0, SpendCents: 500, Conversions: 0},
	}
	var sb strings.Builder
	if err := Write(&sb, rows); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	// CTR and CPA undefined -> trailing empty fields.
	if !strings.Contains(sb.String(), "CMP000,0,0,5.00,0,,") {
		t.Errorf("undefined metrics not written as empty fields.\ngot:\n%s", sb.String())
	}
}
