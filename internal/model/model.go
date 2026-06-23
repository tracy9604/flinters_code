// Package model defines the core data structures and metric calculations
// used when aggregating advertising campaign data.
package model

// Campaign holds the running totals for a single campaign_id.
//
// SpendCents stores spend as an integer number of cents rather than a float.
// Summing many float64 values accumulates rounding error; integer cents keep
// the totals exact and only convert back to dollars at output time.
type Campaign struct {
	ID          string
	Impressions int64
	Clicks      int64
	SpendCents  int64
	Conversions int64
}

// SpendDollars returns the total spend converted from cents to dollars.
func (c *Campaign) SpendDollars() float64 {
	return float64(c.SpendCents) / 100.0
}

// CTR returns the click-through rate (clicks / impressions) and whether it is
// defined. CTR is undefined when there are zero impressions.
func (c *Campaign) CTR() (value float64, ok bool) {
	if c.Impressions == 0 {
		return 0, false
	}
	return float64(c.Clicks) / float64(c.Impressions), true
}

// CPA returns the cost per acquisition (spend / conversions) and whether it is
// defined. CPA is undefined when there are zero conversions.
func (c *Campaign) CPA() (value float64, ok bool) {
	if c.Conversions == 0 {
		return 0, false
	}
	return c.SpendDollars() / float64(c.Conversions), true
}
