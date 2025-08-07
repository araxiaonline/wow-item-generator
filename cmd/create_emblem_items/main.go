package main

import (
	"flag"
	"log"
	"os"

	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"github.com/araxiaonline/endgame-item-generator/internal/items"
	"github.com/gocarina/gocsv"
	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

// This will accept a list of existing items pre-scaled by ChatGPT and scale stats
// based on our server modifiers and tier modifiers.  A sample format is in the same directory.
func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	godotenv.Load("../../.env")

	filename := flag.String("filename", "", "csv of the items to read in")
	tier := flag.Int("tier", 1, "tier of the items to read in")
	flag.Parse()

	if *filename == "" {
		log.Fatal("item file is required")
	}

	itemsFile, err := os.OpenFile(*filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer itemsFile.Close()

	// Connection to mysql database
	mysqlDb, err := mysql.Connect(&mysql.MySqlConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})

	if err != nil {
		log.Fatal(err)
	}

	csvItems := []*mysql.DbItemCsv{}

	if err := gocsv.UnmarshalFile(itemsFile, &csvItems); err != nil { // Load items from file
		panic(err)
	}

	// dbItems := []*mysql.DbItem{}

	for _, item := range csvItems {
		// ConvertCsvToDbItem already tries to find the original item and preserve its fields
		dbItem, err := mysqlDb.ConvertCsvToDbItem(*item)
		if err != nil {
			log.Printf("Failed to convert item %d - %s to DbItem: %v", item.Entry, item.Name, err)
			continue
		}

		// Get the original item for reference (e.g., for scaling calculations)
		originalEntry := item.Entry - 2000000
		originalItem, err := mysqlDb.GetItem(originalEntry)
		if err != nil {
			log.Printf("Failed to get original item %d - %s: %v", originalEntry, item.Name, err)
			continue
		}

		// Create a new item from the dbItem (which already has preserved fields)
		newItem := items.ItemFromDbItem(dbItem)

		// Scale armor based on the new item level
		newItem.ScaleArmor(*item.ItemLevel)

		if *newItem.Class == 2 && *newItem.MinDmg1 > 0 {
			_, err := newItem.ScaleDPS(*originalItem.ItemLevel, *item.ItemLevel)
			if err != nil {
				log.Printf("Failed to scale DPS: %v", err)
			} else {
				log.Printf("Successfully scaled DPS for item %d - %s to oldMinD	dmg1: %f, oldMaxDmg1: %f, minDmg1: %f, maxDmg1: %f", item.Entry, item.Name, *originalItem.MinDmg1, *originalItem.MaxDmg1, *newItem.MinDmg1, *newItem.MaxDmg1)
			}
		}

		// Now need to apply tier modifiers and stat modifiers to the items
		newItem.ApplyTierModifiers(*tier)

		// Make a copy of the spells for the new item
		spells, err := newItem.GetSpells()
		if err != nil {
			log.Printf("Failed to get spells for item %d - %s: %v", item.Entry, item.Name, err)
			continue
		}

		for _, spell := range spells {

			newSpellId := spell.ID + 3000000

			// Copy the spell to the new vendor table (why vendor... not sure just random I guess I made up)
			mysqlDb.CopySpell("spell_dbc", "spells_new_vendor", spell.ID, newSpellId)

			// Scale the spell now and replace the key scaling aspects.
			spell.ForceScaleSpell(*originalItem.ItemLevel, *newItem.ItemLevel, *newItem.Quality, *tier)

			// Create a copy of the spell with the new ID for the update
			scaledSpell := spell.DbSpell
			scaledSpell.ID = newSpellId

			// Write the scaled spell values to update the copied spell
			// log.Printf("Writing scaled spell ID %d with base points: %d, %d, %d",
			// 	newSpellId, scaledSpell.EffectBasePoints1, scaledSpell.EffectBasePoints2, scaledSpell.EffectBasePoints3)
			mysqlDb.WriteSpell("spells_new_vendor", scaledSpell)

			// Update the original newItem spellID with the new scaled spell ID
			newItem.UpdateSpellID(spell.ID, newSpellId)
		}

		// First, copy the original item to preserve all fields
		newEntry := originalEntry + 2000000
		err = mysqlDb.CopyItem("item_template", "item_template_new_vendor", originalEntry, newEntry)
		if err != nil {
			log.Printf("Failed to copy item %d: %v", originalEntry, err)
			continue
		}

		// Then write the updated item to override specific fields
		newItem.DbItem.Entry = newEntry
		mysqlDb.WriteItem("item_template_new_vendor", newItem.DbItem)
		log.Printf("Successfully wrote item %d - %s to database", newEntry, item.Name)

		// oldGenEntry := originalEntry + 20000000
		// oldGenItem, err := mysqlDb.GetItem(oldGenEntry)

		// if err != nil {
		// 	//log.Printf("Failed to get old generated item %d - %s: %v", oldGenEntry, item.Name, err)
		// 	// Even if we can't find the old item, still show the new item's stats
		// 	emptyOldItem := mysql.DbItem{}
		// 	// Set the name for the comparison output
		// 	// Pass -1 for itemClass to disable filtering
		// 	ComparePowerStats(newItem.DbItem, emptyOldItem, item.Name, 4, 7)
		// } else {
		// 	// Compare Attack Power and Spell Power stats between old and new items
		// 	ComparePowerStats(newItem.DbItem, oldGenItem, item.Name, 4, 7)
		// }

		// err = mysqlDb.WriteItem("item_template_new_vendor", dbItem)
		// if err != nil {
		// 	log.Printf("Failed to write item %d - %s to database: %v", item.Entry, item.Name, err)
		// 	continue
		// } else {
		// 	log.Printf("Successfully wrote item %d - %s to database", item.Entry, item.Name)
		// }

		// Print the item in a more readable format
		// jsonData, err := json.MarshalIndent(dbItem, "", "  ")
		// if err != nil {
		// 	log.Printf("Error marshaling item to JSON: %v", err)
		// } else {
		// 	log.Printf("Item %d - %s:\n%s", dbItem.Entry, dbItem.Name, string(jsonData))
		// }
	}
}
