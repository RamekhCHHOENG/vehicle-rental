package handlers

import "time"

// RentalDays returns the number of charged days between pick-up and return.
// A same-day return counts as 0 (invalid); pick-up Monday, return Tuesday = 1 day.
func RentalDays(start, end time.Time) int {
	return int(end.Sub(start).Hours() / 24)
}

// RentalTotal computes the full, no-hidden-fees price for a rental.
func RentalTotal(pricePerDay float64, days int) float64 {
	return pricePerDay * float64(days)
}

// DatesOverlap reports whether [aStart, aEnd) overlaps [bStart, bEnd).
// The end date is the return day, so back-to-back rentals
// (one ends the day the next starts) do NOT overlap.
func DatesOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}
