package database

import (
	"log"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

// Vehicles used to carry make, model and location as free text. This moves
// those rows onto the reference tables and then drops the old columns.
//
// It has to drop them: they are NOT NULL and the Vehicle struct no longer
// writes them, so leaving them in place would fail every future insert. That
// makes the order load-bearing — read the text, match it, write the ids, and
// only then drop.
//
// The matching is deliberately forgiving. Real rows look like
// make="Lamborghini Veneno ", model="Lamborghini" — the fields were swapped and
// the whitespace kept — so both are searched as one string. Anything that
// still finds no match gets a make or model created for it rather than being
// dropped on the floor; an admin can merge the stragglers afterwards.

type legacyVehicle struct {
	ID       uuid.UUID
	Type     string
	Make     string
	Model    string
	Location string
}

func backfillVehicleReferences(db *gorm.DB) error {
	m := db.Migrator()
	hasMake := m.HasColumn(&models.Vehicle{}, "make")
	hasLocation := m.HasColumn(&models.Vehicle{}, "location")
	if !hasMake && !hasLocation {
		return nil
	}

	var rows []legacyVehicle
	if err := db.Table("vehicles").Select("id, type, make, model, location").Scan(&rows).Error; err != nil {
		return err
	}

	if len(rows) > 0 {
		log.Printf("metadata: migrating %d vehicle(s) off free-text make/model/location", len(rows))
	}

	for _, row := range rows {
		updates := map[string]any{}

		if id, ok := matchProvince(db, row.Location); ok {
			updates["province_id"] = id
		} else if strings.TrimSpace(row.Location) != "" {
			log.Printf("metadata: vehicle %s has location %q matching no province — left unset", row.ID, row.Location)
		}

		makeID, modelID, err := matchMakeAndModel(db, row)
		if err != nil {
			return err
		}
		if makeID != uuid.Nil {
			updates["make_id"] = makeID
			updates["model_id"] = modelID
		}

		if len(updates) == 0 {
			continue
		}
		if err := db.Table("vehicles").Where("id = ?", row.ID).Updates(updates).Error; err != nil {
			return err
		}
	}

	for _, col := range []string{"make", "model", "location"} {
		if !m.HasColumn(&models.Vehicle{}, col) {
			continue
		}
		if err := m.DropColumn(&models.Vehicle{}, col); err != nil {
			return err
		}
	}
	log.Printf("metadata: dropped the legacy make/model/location columns")

	return nil
}

// matchProvince resolves free-text location against the seeded province names,
// tolerating case and stray whitespace.
func matchProvince(db *gorm.DB, location string) (uuid.UUID, bool) {
	needle := strings.ToLower(strings.TrimSpace(location))
	if needle == "" {
		return uuid.Nil, false
	}

	var provinces []models.Province
	if err := db.Find(&provinces).Error; err != nil {
		return uuid.Nil, false
	}
	for _, p := range provinces {
		name := strings.ToLower(p.NameEn)
		if needle == name || strings.Contains(needle, name) || p.NameKm == strings.TrimSpace(location) {
			return p.ID, true
		}
	}
	return uuid.Nil, false
}

// matchMakeAndModel searches the make and model text as a single string,
// because the two were routinely filled in the wrong order. The longest make
// name wins so that "Range Rover" is not beaten by a shorter make it contains.
func matchMakeAndModel(db *gorm.DB, row legacyVehicle) (uuid.UUID, *uuid.UUID, error) {
	text := strings.ToLower(strings.Join(strings.Fields(row.Make+" "+row.Model), " "))
	if text == "" {
		return uuid.Nil, nil, nil
	}

	var makes []models.VehicleMake
	if err := db.Find(&makes).Error; err != nil {
		return uuid.Nil, nil, err
	}

	names := make([]string, len(makes))
	for i, mk := range makes {
		names[i] = mk.Name
	}

	var matched *models.VehicleMake
	i, remainder := longestMatch(text, names)
	if i >= 0 {
		matched = &makes[i]
	} else {
		// Unrecognised manufacturer: keep it rather than lose the listing.
		name := titleCase(strings.Fields(text)[0])
		created := models.VehicleMake{Name: name, SortOrder: 900, Active: true}
		if err := db.Where("name = ?", name).FirstOrCreate(&created, models.VehicleMake{Name: name}).Error; err != nil {
			return uuid.Nil, nil, err
		}
		log.Printf("metadata: created make %q for vehicle %s", name, row.ID)
		matched = &created
		remainder = strings.Join(strings.Fields(text)[1:], " ")
	}

	var candidates []models.VehicleModel
	if err := db.Where("make_id = ?", matched.ID).Find(&candidates).Error; err != nil {
		return uuid.Nil, nil, err
	}
	modelNames := make([]string, len(candidates))
	for i, c := range candidates {
		modelNames[i] = c.Name
	}
	if j, _ := longestMatch(remainder, modelNames); j >= 0 {
		return matched.ID, &candidates[j].ID, nil
	}

	if remainder == "" {
		return matched.ID, nil, nil
	}
	name := titleCase(remainder)
	created := models.VehicleModel{
		MakeID: matched.ID, Name: name, Type: models.VehicleType(row.Type), Active: true,
	}
	if err := db.Where("make_id = ? AND name = ?", matched.ID, name).
		FirstOrCreate(&created, models.VehicleModel{MakeID: matched.ID, Name: name}).Error; err != nil {
		return uuid.Nil, nil, err
	}
	log.Printf("metadata: created model %q under %s for vehicle %s", name, matched.Name, row.ID)
	return matched.ID, &created.ID, nil
}

// titleCase capitalises each word, so "model y" is stored as "Model Y" and sits
// alongside the seeded names rather than looking like a different entry.
// Hyphens count as word breaks, because the imported catalogue is full of
// "MERCEDES-BENZ" and "Mercedes-benz" would look like a typo.
func titleCase(s string) string {
	upperNext := true
	return strings.Map(func(r rune) rune {
		switch {
		case r == ' ' || r == '-':
			upperNext = true
			return r
		case upperNext:
			upperNext = false
			return []rune(strings.ToUpper(string(r)))[0]
		default:
			return r
		}
	}, strings.Join(strings.Fields(s), " "))
}

// longestMatch finds which of names appears in text, preferring the longest so
// that "Range Rover" is not beaten by a "Rover" it contains, and returns what
// is left of text once that name is taken out. The leftover is the model:
// "lamborghini veneno lamborghini" minus its make is "veneno", whichever field
// the owner originally typed each half into.
//
// Returns -1 when nothing matches.
func longestMatch(text string, names []string) (int, string) {
	text = strings.ToLower(strings.Join(strings.Fields(text), " "))
	if text == "" {
		return -1, ""
	}

	best := -1
	for i, name := range names {
		lower := strings.ToLower(strings.TrimSpace(name))
		if lower == "" || !strings.Contains(text, lower) {
			continue
		}
		if best == -1 || len(lower) > len(strings.ToLower(names[best])) {
			best = i
		}
	}
	if best == -1 {
		return -1, text
	}

	remainder := strings.ReplaceAll(text, strings.ToLower(names[best]), " ")
	return best, strings.Join(strings.Fields(remainder), " ")
}
