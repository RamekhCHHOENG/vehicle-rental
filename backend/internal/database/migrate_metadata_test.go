package database

import "testing"

// The rows this has to survive were typed by real owners, who put the model in
// the make field and the make in the model field. Both cases below are verbatim
// from production.
func TestLongestMatch(t *testing.T) {
	makes := []string{"Toyota", "Lexus", "Honda", "Rover", "Range Rover", "Lamborghini", "Tesla"}

	tests := []struct {
		name          string
		text          string
		names         []string
		wantMatch     string
		wantRemainder string
	}{
		{"fields swapped, trailing space", "lamborghini veneno  lamborghini", makes, "Lamborghini", "veneno"},
		{"model spans two words", "tesla model y tesla", makes, "Tesla", "model y"},
		{"longest name wins over the one inside it", "range rover sport", makes, "Range Rover", "sport"},
		{"ordinary make and model", "toyota camry", makes, "Toyota", "camry"},
		{"case is ignored", "HONDA Civic", makes, "Honda", "civic"},
		{"nothing matches", "polestar 2", makes, "", "polestar 2"},
		{"empty text", "", makes, "", ""},
		{"make only leaves nothing behind", "lexus", makes, "Lexus", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, remainder := longestMatch(tt.text, tt.names)

			got := ""
			if i >= 0 {
				got = tt.names[i]
			}
			if got != tt.wantMatch {
				t.Errorf("match = %q, want %q", got, tt.wantMatch)
			}
			if remainder != tt.wantRemainder {
				t.Errorf("remainder = %q, want %q", remainder, tt.wantRemainder)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct{ in, want string }{
		{"model y", "Model Y"},
		{"veneno", "Veneno"},
		{"  cx-5  ", "Cx-5"},
		{"mercedes-benz", "Mercedes-Benz"},
		{"land  cruiser", "Land Cruiser"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := titleCase(tt.in); got != tt.want {
				t.Errorf("titleCase(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
