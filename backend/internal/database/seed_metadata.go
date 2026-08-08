package database

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ramekhchhoeng/car-rental/backend/internal/models"
)

// The reference data every listing is written against. It is seeded on first
// startup and then owned by the admin screens — this file is the starting
// vocabulary, not the permanent one, so entries are matched on their natural
// key and left alone if they already exist. Renaming "Toyota" in the admin UI
// must not be undone by the next deploy.

type provinceSeed struct {
	code, en, km string
}

// Cambodia's 25 provinces, in the order they are conventionally listed.
// Phnom Penh is pulled to the front: it is where most listings will be.
var provinceSeeds = []provinceSeed{
	{"phnom-penh", "Phnom Penh", "ភ្នំពេញ"},
	{"banteay-meanchey", "Banteay Meanchey", "បន្ទាយមានជ័យ"},
	{"battambang", "Battambang", "បាត់ដំបង"},
	{"kampong-cham", "Kampong Cham", "កំពង់ចាម"},
	{"kampong-chhnang", "Kampong Chhnang", "កំពង់ឆ្នាំង"},
	{"kampong-speu", "Kampong Speu", "កំពង់ស្ពឺ"},
	{"kampong-thom", "Kampong Thom", "កំពង់ធំ"},
	{"kampot", "Kampot", "កំពត"},
	{"kandal", "Kandal", "កណ្ដាល"},
	{"kep", "Kep", "កែប"},
	{"koh-kong", "Koh Kong", "កោះកុង"},
	{"kratie", "Kratie", "ក្រចេះ"},
	{"mondulkiri", "Mondulkiri", "មណ្ឌលគិរី"},
	{"oddar-meanchey", "Oddar Meanchey", "ឧត្ដរមានជ័យ"},
	{"pailin", "Pailin", "ប៉ៃលិន"},
	{"preah-sihanouk", "Preah Sihanouk", "ព្រះសីហនុ"},
	{"preah-vihear", "Preah Vihear", "ព្រះវិហារ"},
	{"prey-veng", "Prey Veng", "ព្រៃវែង"},
	{"pursat", "Pursat", "ពោធិ៍សាត់"},
	{"ratanakiri", "Ratanakiri", "រតនគិរី"},
	{"siem-reap", "Siem Reap", "សៀមរាប"},
	{"stung-treng", "Stung Treng", "ស្ទឹងត្រែង"},
	{"svay-rieng", "Svay Rieng", "ស្វាយរៀង"},
	{"takeo", "Takeo", "តាកែវ"},
	{"tbong-khmum", "Tbong Khmum", "ត្បូងឃ្មុំ"},
}

type makeSeed struct {
	name       string
	cars       []string
	motorbikes []string
}

// Weighted towards what is actually on the road here: Toyota and Lexus for
// cars, Honda for motorbikes. The luxury makes are thin on purpose — they are
// listed occasionally and an admin can extend any of this without a deploy.
var makeSeeds = []makeSeed{
	{name: "Toyota", cars: []string{"Alphard", "Camry", "Corolla", "Fortuner", "Highlander", "Hiace", "Hilux", "Land Cruiser", "Prius", "RAV4", "Vios"}},
	{name: "Lexus", cars: []string{"ES", "GX", "LX", "NX", "RX"}},
	{name: "Honda", cars: []string{"Accord", "City", "Civic", "CR-V", "Fit", "Jazz"}, motorbikes: []string{"Air Blade", "Click", "Dream", "PCX", "Scoopy", "Wave", "Winner"}},
	{name: "Hyundai", cars: []string{"Accent", "Elantra", "Santa Fe", "Starex", "Tucson"}},
	{name: "Kia", cars: []string{"Carnival", "Cerato", "Morning", "Seltos", "Sorento", "Sportage"}},
	{name: "Ford", cars: []string{"Escape", "Everest", "Ranger", "Territory"}},
	{name: "Mitsubishi", cars: []string{"Outlander", "Pajero", "Triton", "Xpander"}},
	{name: "Nissan", cars: []string{"Almera", "Navara", "Terra", "X-Trail"}},
	{name: "Mazda", cars: []string{"BT-50", "CX-5", "CX-8", "Mazda3"}},
	{name: "Suzuki", cars: []string{"Ertiga", "Swift"}, motorbikes: []string{"Raider", "Smash"}},
	{name: "Yamaha", motorbikes: []string{"Exciter", "Fino", "Grande", "Mio", "NMAX", "Nouvo"}},
	{name: "SYM", motorbikes: []string{"Angel", "Attila", "Elegant"}},
	{name: "Vespa", motorbikes: []string{"LX", "Primavera", "Sprint"}},
	{name: "BYD", cars: []string{"Atto 3", "Dolphin", "Han", "Seal", "Song"}},
	{name: "Tesla", cars: []string{"Model 3", "Model S", "Model X", "Model Y"}},
	{name: "BMW", cars: []string{"3 Series", "5 Series", "X3", "X5", "X7"}},
	{name: "Mercedes-Benz", cars: []string{"C-Class", "E-Class", "GLC", "GLE", "S-Class"}},
	{name: "Range Rover", cars: []string{"Evoque", "Sport", "Velar", "Vogue"}},
	{name: "Porsche", cars: []string{"Cayenne", "Macan", "Panamera"}},
	{name: "Lamborghini", cars: []string{"Aventador", "Huracan", "Urus", "Veneno"}},
}

type featureSeed struct {
	code, name, icon string
	appliesTo        models.VehicleType
}

// appliesTo empty means the feature suits both cars and motorbikes.
var featureSeeds = []featureSeed{
	{code: "air-conditioning", name: "Air conditioning", icon: "i-lucide-snowflake", appliesTo: models.VehicleCar},
	{code: "gps", name: "GPS navigation", icon: "i-lucide-map-pin"},
	{code: "bluetooth", name: "Bluetooth audio", icon: "i-lucide-bluetooth", appliesTo: models.VehicleCar},
	{code: "usb-charging", name: "USB charging", icon: "i-lucide-battery-charging"},
	{code: "child-seat", name: "Child seat", icon: "i-lucide-baby", appliesTo: models.VehicleCar},
	{code: "dashcam", name: "Dashcam", icon: "i-lucide-video", appliesTo: models.VehicleCar},
	{code: "spare-tyre", name: "Spare tyre", icon: "i-lucide-circle-dot", appliesTo: models.VehicleCar},
	{code: "helmet", name: "Helmets included", icon: "i-lucide-hard-hat", appliesTo: models.VehicleMotorbike},
	{code: "rain-gear", name: "Rain gear", icon: "i-lucide-cloud-rain", appliesTo: models.VehicleMotorbike},
	{code: "phone-holder", name: "Phone holder", icon: "i-lucide-smartphone", appliesTo: models.VehicleMotorbike},
	{code: "top-box", name: "Top box", icon: "i-lucide-package", appliesTo: models.VehicleMotorbike},
	{code: "delivery", name: "Delivery available", icon: "i-lucide-truck"},
	{code: "insurance", name: "Full insurance", icon: "i-lucide-shield-check"},
}

// seedMetadata inserts anything missing and leaves everything present alone.
// DoNothing on conflict is what makes it safe to run on every startup.
func seedMetadata(db *gorm.DB) error {
	provinces := make([]models.Province, len(provinceSeeds))
	for i, p := range provinceSeeds {
		provinces[i] = models.Province{
			Code: p.code, NameEn: p.en, NameKm: p.km, SortOrder: i, Active: true,
		}
	}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&provinces).Error; err != nil {
		return err
	}

	for i, ms := range makeSeeds {
		mk := models.VehicleMake{Name: ms.name, SortOrder: i, Active: true}
		if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
			Create(&mk).Error; err != nil {
			return err
		}
		// OnConflict-DoNothing leaves ID zero when the row already existed, so
		// read back the one that is actually in the table.
		if err := db.Where("name = ?", ms.name).First(&mk).Error; err != nil {
			return err
		}

		var vehicleModels []models.VehicleModel
		for _, name := range ms.cars {
			vehicleModels = append(vehicleModels, models.VehicleModel{
				MakeID: mk.ID, Name: name, Type: models.VehicleCar, Active: true,
			})
		}
		for _, name := range ms.motorbikes {
			vehicleModels = append(vehicleModels, models.VehicleModel{
				MakeID: mk.ID, Name: name, Type: models.VehicleMotorbike, Active: true,
			})
		}
		if len(vehicleModels) == 0 {
			continue
		}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "make_id"}, {Name: "name"}},
			DoNothing: true,
		}).Create(&vehicleModels).Error; err != nil {
			return err
		}
	}

	features := make([]models.Feature, len(featureSeeds))
	for i, f := range featureSeeds {
		features[i] = models.Feature{
			Code: f.code, Name: f.name, Icon: f.icon, AppliesTo: f.appliesTo, SortOrder: i, Active: true,
		}
	}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&features).Error; err != nil {
		return err
	}

	var counts struct{ Provinces, Makes, Models, Features int64 }
	db.Model(&models.Province{}).Count(&counts.Provinces)
	db.Model(&models.VehicleMake{}).Count(&counts.Makes)
	db.Model(&models.VehicleModel{}).Count(&counts.Models)
	db.Model(&models.Feature{}).Count(&counts.Features)
	log.Printf("metadata: %d provinces, %d makes, %d models, %d features",
		counts.Provinces, counts.Makes, counts.Models, counts.Features)

	return nil
}
