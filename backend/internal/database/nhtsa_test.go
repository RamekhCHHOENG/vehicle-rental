package database

import "testing"

// NHTSA shouts every name. Casing them is the only cleanup we do, so it has to
// not mangle the acronyms or overwrite names somebody wrote deliberately.
func TestNormaliseName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"shouted single word", "TOYOTA", "Toyota"},
		{"shouted two words", "ASTON MARTIN", "Aston Martin"},
		{"hyphenated", "MERCEDES-BENZ", "Mercedes-Benz"},
		{"acronym stays an acronym", "BMW", "BMW"},
		{"acronym in the middle of a name", "MV AGUSTA", "MV Agusta"},
		{"mixed case is left alone", "CFMoto", "CFMoto"},
		{"already correct", "Honda", "Honda"},
		{"whitespace collapsed", "  LAND   ROVER  ", "Land Rover"},
		{"empty", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normaliseName(tt.in); got != tt.want {
				t.Errorf("normaliseName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
