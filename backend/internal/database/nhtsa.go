package database

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

// Importing the make and model catalogue from NHTSA's vPIC API — the US
// Department of Transportation's vehicle database. It is free, needs no key, and
// is the only open source of this data with real coverage.
//
// It is imported, not proxied. Every listing holds a foreign key into our own
// tables, so the catalogue has to outlive any outage at vpic.nhtsa.dot.gov, and
// an admin has to be able to retire an entry we disagree with. Importing also
// means the price of their per-make model endpoint is paid once.
//
// The data is US-registration-oriented and imperfect: names arrive shouted
// ("MERCEDES-BENZ"), Toyota's model list includes Scions, and the 1,684
// motorcycle "makes" include a long tail of tiny importers. Nothing here tries
// to clean that up beyond casing — the admin screen can retire what does not
// belong, and search makes the volume harmless.

const nhtsaBase = "https://vpic.nhtsa.dot.gov/api/vehicles"

// NHTSA calls them "car" and "motorcycle"; we call the second one motorbike.
var nhtsaTypes = map[string]models.VehicleType{
	"car":        models.VehicleCar,
	"motorcycle": models.VehicleMotorbike,
}

// Casing NHTSA shouts at us. Title-casing turns "MERCEDES-BENZ" into something
// readable but also turns "BMW" into "Bmw", so acronyms are listed rather than
// guessed at. Existing rows are matched case-insensitively and keep their own
// spelling, so the seeded names always win over anything imported.
var knownCasing = map[string]string{
	"bmw": "BMW", "gmc": "GMC", "mg": "MG", "byd": "BYD", "sym": "SYM",
	"ktm": "KTM", "bsa": "BSA", "cfmoto": "CFMoto", "ducati": "Ducati",
	"kymco": "KYMCO", "tvs": "TVS", "ural": "Ural", "ssr": "SSR",
	"am general": "AM General", "mv agusta": "MV Agusta",
}

type ImportResult struct {
	MakesAdded  int `json:"makes_added"`
	ModelsAdded int `json:"models_added"`
	MakesSeen   int `json:"makes_seen"`
	Failures    int `json:"failures"`
}

type nhtsaMakesResponse struct {
	Count   int `json:"Count"`
	Results []struct {
		MakeID   int    `json:"MakeId"`
		MakeName string `json:"MakeName"`
	} `json:"Results"`
}

type nhtsaModelsResponse struct {
	Count   int `json:"Count"`
	Results []struct {
		ModelName string `json:"Model_Name"`
	} `json:"Results"`
}

// ImportFromNHTSA pulls the makes for each vehicle type we rent, and optionally
// each make's models. withModels costs one request per make — around 1,900 of
// them — so the caller decides whether to pay it.
func ImportFromNHTSA(db *gorm.DB, withModels bool) (ImportResult, error) {
	var result ImportResult
	client := &http.Client{Timeout: 30 * time.Second}

	for nhtsaType, vehicleType := range nhtsaTypes {
		var payload nhtsaMakesResponse
		url := fmt.Sprintf("%s/GetMakesForVehicleType/%s?format=json", nhtsaBase, nhtsaType)
		if err := fetchJSON(client, url, &payload); err != nil {
			return result, fmt.Errorf("fetching %s makes: %w", nhtsaType, err)
		}

		log.Printf("nhtsa: %d %s makes", payload.Count, nhtsaType)

		for _, row := range payload.Results {
			name := normaliseName(row.MakeName)
			if name == "" {
				continue
			}
			result.MakesSeen++

			mk, created, err := upsertMake(db, name)
			if err != nil {
				return result, err
			}
			if created {
				result.MakesAdded++
			}
			if !withModels {
				continue
			}

			added, err := importModels(client, db, mk, row.MakeName, vehicleType)
			if err != nil {
				// One make's models failing is not worth abandoning the import.
				log.Printf("nhtsa: models for %s failed: %v", name, err)
				result.Failures++
				continue
			}
			result.ModelsAdded += added
		}
	}

	log.Printf("nhtsa: %d makes seen, %d makes added, %d models added, %d failures",
		result.MakesSeen, result.MakesAdded, result.ModelsAdded, result.Failures)
	return result, nil
}

// upsertMake finds a make by name without regard to case, so an imported
// "TOYOTA" attaches to the seeded "Toyota" instead of sitting beside it.
func upsertMake(db *gorm.DB, name string) (models.VehicleMake, bool, error) {
	var existing models.VehicleMake
	err := db.Where("LOWER(name) = ?", strings.ToLower(name)).First(&existing).Error
	if err == nil {
		return existing, false, nil
	}
	if err != gorm.ErrRecordNotFound {
		return existing, false, err
	}

	// Imported makes sort after the curated ones: the twenty that actually rent
	// here should stay at the top of a list of two thousand.
	created := models.VehicleMake{Name: name, SortOrder: 500, Active: true}
	if err := db.Create(&created).Error; err != nil {
		// A concurrent import may have won the race; take whatever is there.
		if lookupErr := db.Where("LOWER(name) = ?", strings.ToLower(name)).First(&existing).Error; lookupErr == nil {
			return existing, false, nil
		}
		return created, false, err
	}
	return created, true, nil
}

func importModels(client *http.Client, db *gorm.DB, mk models.VehicleMake, nhtsaName string, vehicleType models.VehicleType) (int, error) {
	var payload nhtsaModelsResponse
	url := fmt.Sprintf("%s/GetModelsForMake/%s?format=json", nhtsaBase, urlPart(nhtsaName))
	if err := fetchJSON(client, url, &payload); err != nil {
		return 0, err
	}
	if len(payload.Results) == 0 {
		return 0, nil
	}

	rows := make([]models.VehicleModel, 0, len(payload.Results))
	for _, row := range payload.Results {
		name := normaliseName(row.ModelName)
		if name == "" {
			continue
		}
		rows = append(rows, models.VehicleModel{
			MakeID: mk.ID, Name: name, Type: vehicleType, Active: true,
		})
	}
	if len(rows) == 0 {
		return 0, nil
	}

	// DoNothing on the (make_id, name) key: re-running the import is free, and a
	// model an admin has retired stays retired.
	res := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "make_id"}, {Name: "name"}},
		DoNothing: true,
	}).Create(&rows)
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}

func fetchJSON(client *http.Client, url string, dest any) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

// normaliseName trims, collapses whitespace, and fixes the shouting.
func normaliseName(raw string) string {
	name := strings.Join(strings.Fields(raw), " ")
	if name == "" {
		return ""
	}
	if fixed, ok := knownCasing[strings.ToLower(name)]; ok {
		return fixed
	}
	// Already mixed case means someone wrote it deliberately; leave it.
	if name != strings.ToUpper(name) {
		return name
	}
	return titleCase(strings.ToLower(name))
}

// urlPart escapes only what NHTSA's path segments actually need.
func urlPart(name string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(name), "%20"), "#", "")
}
