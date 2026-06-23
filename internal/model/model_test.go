package model

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestSpendDollars(t *testing.T) {
	c := &Campaign{SpendCents: 9370}
	if got := c.SpendDollars(); !almostEqual(got, 93.70) {
		t.Fatalf("SpendDollars() = %v, want 93.70", got)
	}
}

func TestCTR(t *testing.T) {
	c := &Campaign{Impressions: 26000, Clicks: 640}
	got, ok := c.CTR()
	if !ok {
		t.Fatal("CTR() ok = false, want true")
	}
	if !almostEqual(got, 640.0/26000.0) {
		t.Fatalf("CTR() = %v, want %v", got, 640.0/26000.0)
	}
}

func TestCTRZeroImpressions(t *testing.T) {
	c := &Campaign{Impressions: 0, Clicks: 5}
	if _, ok := c.CTR(); ok {
		t.Fatal("CTR() ok = true for zero impressions, want false")
	}
}

func TestCPA(t *testing.T) {
	c := &Campaign{SpendCents: 1500, Conversions: 3}
	got, ok := c.CPA()
	if !ok {
		t.Fatal("CPA() ok = false, want true")
	}
	if !almostEqual(got, 5.0) {
		t.Fatalf("CPA() = %v, want 5.0", got)
	}
}

func TestCPAZeroConversions(t *testing.T) {
	c := &Campaign{SpendCents: 1500, Conversions: 0}
	if _, ok := c.CPA(); ok {
		t.Fatal("CPA() ok = true for zero conversions, want false")
	}
}
