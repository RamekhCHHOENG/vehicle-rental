package handlers

import "testing"

// validate is the half of the check that needs no database. The reference ids
// are proven in resolve, against live rows, so nothing here asserts on them.
func TestVehicleRequestValidate(t *testing.T) {
	valid := func() vehicleRequest {
		return vehicleRequest{
			Type: "car", Year: 2020, Transmission: "auto",
			Seats: 5, PricePerDay: 30,
		}
	}

	tests := []struct {
		name    string
		mutate  func(*vehicleRequest)
		wantErr string
	}{
		{"a complete request passes", func(r *vehicleRequest) {}, ""},
		{"motorbike is a type too", func(r *vehicleRequest) { r.Type = "motorbike"; r.Seats = 2 }, ""},
		{"unknown type", func(r *vehicleRequest) { r.Type = "helicopter" }, "type must be car or motorbike"},
		{"missing type", func(r *vehicleRequest) { r.Type = "" }, "type must be car or motorbike"},
		{"year before cars were worth renting", func(r *vehicleRequest) { r.Year = 1890 }, "year looks invalid"},
		{"year in the far future", func(r *vehicleRequest) { r.Year = 3000 }, "year looks invalid"},
		{"unknown transmission", func(r *vehicleRequest) { r.Transmission = "cvt" }, "transmission must be auto or manual"},
		{"free is not a price", func(r *vehicleRequest) { r.PricePerDay = 0 }, "price per day must be positive"},
		{"negative price", func(r *vehicleRequest) { r.PricePerDay = -5 }, "price per day must be positive"},
		{"zero seats", func(r *vehicleRequest) { r.Seats = 0 }, "seats must be between 1 and 64"},
		{"a bus is not a rental car", func(r *vehicleRequest) { r.Seats = 900 }, "seats must be between 1 and 64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid()
			tt.mutate(&req)

			got := req.validate()
			if got != tt.wantErr {
				t.Errorf("validate() = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

// Codes end up in URLs and in the features query parameter, so they have to be
// predictable whatever an admin types into the name field.
func TestSlugify(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Phnom Penh", "phnom-penh"},
		{"  Kampong   Cham  ", "kampong-cham"},
		{"Preah Sihanouk!", "preah-sihanouk"},
		{"Air conditioning", "air-conditioning"},
		{"4WD / Off-road", "4wd-off-road"},
		{"!!!", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := slugify(tt.in); got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// An omitted "active" must not retire a row that was active, which is the whole
// reason the field is a pointer.
func TestBoolOr(t *testing.T) {
	yes, no := true, false
	tests := []struct {
		name     string
		v        *bool
		fallback bool
		want     bool
	}{
		{"omitted keeps the current value", nil, true, true},
		{"omitted keeps a retired row retired", nil, false, false},
		{"explicit false retires", &no, true, false},
		{"explicit true restores", &yes, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := boolOr(tt.v, tt.fallback); got != tt.want {
				t.Errorf("boolOr() = %v, want %v", got, tt.want)
			}
		})
	}
}
