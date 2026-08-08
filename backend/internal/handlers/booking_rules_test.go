package handlers

import (
	"testing"
	"time"
)

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestRentalDays(t *testing.T) {
	tests := []struct {
		name  string
		start string
		end   string
		want  int
	}{
		{"one day", "2026-08-10", "2026-08-11", 1},
		{"three days", "2026-08-10", "2026-08-13", 3},
		{"same day is zero", "2026-08-10", "2026-08-10", 0},
		{"end before start is negative", "2026-08-13", "2026-08-10", -3},
		{"across month boundary", "2026-08-30", "2026-09-02", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RentalDays(date(tt.start), date(tt.end))
			if got != tt.want {
				t.Errorf("RentalDays(%s, %s) = %d, want %d", tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestRentalTotal(t *testing.T) {
	tests := []struct {
		name        string
		pricePerDay float64
		days        int
		want        float64
	}{
		{"three days at 45", 45, 3, 135},
		{"one day at 12.50", 12.50, 1, 12.50},
		{"week at 100", 100, 7, 700},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RentalTotal(tt.pricePerDay, tt.days)
			if got != tt.want {
				t.Errorf("RentalTotal(%v, %d) = %v, want %v", tt.pricePerDay, tt.days, got, tt.want)
			}
		})
	}
}

func TestDatesOverlap(t *testing.T) {
	tests := []struct {
		name                       string
		aStart, aEnd, bStart, bEnd string
		want                       bool
	}{
		{"identical ranges", "2026-08-10", "2026-08-13", "2026-08-10", "2026-08-13", true},
		{"a inside b", "2026-08-11", "2026-08-12", "2026-08-10", "2026-08-13", true},
		{"partial overlap at end", "2026-08-12", "2026-08-15", "2026-08-10", "2026-08-13", true},
		{"partial overlap at start", "2026-08-08", "2026-08-11", "2026-08-10", "2026-08-13", true},
		{"back-to-back does not overlap", "2026-08-13", "2026-08-15", "2026-08-10", "2026-08-13", false},
		{"completely before", "2026-08-01", "2026-08-05", "2026-08-10", "2026-08-13", false},
		{"completely after", "2026-08-20", "2026-08-25", "2026-08-10", "2026-08-13", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DatesOverlap(date(tt.aStart), date(tt.aEnd), date(tt.bStart), date(tt.bEnd))
			if got != tt.want {
				t.Errorf("DatesOverlap(%s-%s, %s-%s) = %v, want %v",
					tt.aStart, tt.aEnd, tt.bStart, tt.bEnd, got, tt.want)
			}
		})
	}
}
