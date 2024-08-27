package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/araxiaonline/endgame-item-generator/internal/config"
	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"github.com/araxiaonline/endgame-item-generator/internal/db/sqlite"
	"github.com/araxiaonline/endgame-item-generator/internal/items"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	godotenv.Load()
	// database.models.Connect()

	debug := flag.Bool("debug", false, "Enable verbose logging inside generator")
	difficulty := flag.Int("difficulty", 3, "set the difficulty of the dungeon, defaults to 3 (mythic) 4 (legendary) 5 (ascendant)")
	// levelUp := flag.Bool("levelUp", false, "Boss items require higher +1 level to equip, defaults to false")
	baselevel := flag.Int("baselevel", 80, "set the base level for items to be used, defaults to 80 this is required for levelUp flag")
	flag.Parse()

	if difficulty == nil || *difficulty < 3 || *difficulty > 5 {
		log.Fatal("difficulty must be between 3-5")
		os.Exit(1)
	}

	if baselevel == nil || *baselevel < 0 {
		log.Fatal("base level must be greater than 80")
		os.Exit(1)
	}

	var itemLevel *int = new(int)
	switch *difficulty {
	case 3:
		*itemLevel = config.MythicItemLevelStart
	case 4:
		*itemLevel = config.LegendaryItemLevelStart
	case 5:
		*itemLevel = config.AscendantItemLevelStart
	}

	if *debug {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard)
	}

	// Connect to Mysql
	mysqlDb, err := mysql.Connect(&mysql.MySqlConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})

	if err != nil {
		log.Fatal(err)
	}

	// Connect to SqlList for EndGame Mapping
	sqliteDb, err := sqlite.Connect("./data/items.db")
	if err != nil {
		log.Fatal(err)
	}

	// Get all rare items int the acore_world.item_template that are rare or higher quality
	rareItems, err := mysqlDb.GetRarePlusItems(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	// do scaling and write sql for all items that are processed from the rareItems list
	for itr, dbItem := range rareItems {

		// convert from a dbModel item to Item entity
		item := items.ItemFromDbItem(dbItem)

		// the lookup Item is a check to see if the item comes from a dungeon on higher difficulties (4,5) we only process dungeon items
		lookupItem, err := sqliteDb.GetItemFromDungeon(item.Entry)
		if err != nil {
			if !strings.Contains(err.Error(), "no rows in result set") {
				log.Printf("failed to lookup item %v from dungeon: %v", item.Entry, err)
			}
		}
		log.Printf("Lookup %v", lookupItem)
		// skip items not from a dungeon on higher difficulties
		if *difficulty > 3 {
			if lookupItem.Entry == 0 {
				log.Printf("Item %v Entry: %v is not from a dungeon\n", item.Name, item.Entry)
				continue
			} else {
				log.Printf("Item %v Entry: %v is from a dungeon\n", item.Name, item.Entry)
			}
		}

		// if it is a rare item then we need to scale it up to epic
		if *item.Quality < 5 {
			*item.Quality = 4
		}

		statsList, err := item.GetStatList()
		if err != nil {
			log.Print(err)
			continue
		}

		log.Printf("Item: %v Entry: %v StatsList: %v\n", item.Name, item.Entry, statsList)

		var highLevelItem mysql.DbItem
		if *difficulty == 3 {
			rndItem, err := sqliteDb.GetRandItem(*item.Class, *item.Subclass, statsList, false)
			if err != nil {
				log.Print(err)
				continue
			}

			if rndItem == (sqlite.HighLevelItem{}) {
				log.Fatalf("Failed to get random item for %v Entry: %v\n", item.Name, item.Entry)
			}

			log.Printf("Random Item: %v Entry: %v\n", rndItem.Name, rndItem.Entry)

			// Take the high level item that has been selected for stats and remap to current item
			highLevelItem, err = mysqlDb.GetItem(rndItem.Entry)
			if err != nil {
				log.Fatal(err)
				continue
			}
		} else {

			highLevelItem, err = mysqlDb.GetByNameAndDifficulty(item.Name, *difficulty-1)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		// difficulty is used to tweak things in the scaling proces specifically modifiers so stats are not inflated twice by quality multiples
		item.SetDifficulty(*difficulty)

		// if the item is not from a dungeon and we made it here, then just scale to mythic which can be used for weekly loot chests or new recipes.
		if lookupItem.Entry == 0 {
			Scale(highLevelItem, &item, *itemLevel, *item.Quality)
			fmt.Print(items.ItemToSql(item, *baselevel, *difficulty))
			continue
		}

		// if the item is from a dungeon and not a boss item
		if lookupItem.CreatureId == 0 {

			if lookupItem.DungeonLevel < 60 && lookupItem.Expansion == 0 {
				Scale(highLevelItem, &item, *itemLevel+5, *item.Quality)
				fmt.Print(items.ItemToSql(item, *baselevel, *difficulty))
			}

			if lookupItem.DungeonLevel == 60 && lookupItem.Expansion == 0 {
				Scale(highLevelItem, &item, *itemLevel+10, *item.Quality)
				fmt.Print(items.ItemToSql(item, *baselevel, *difficulty))
			}

			if lookupItem.DungeonLevel < 70 && lookupItem.Expansion == 1 {
				Scale(highLevelItem, &item, *itemLevel+7, *item.Quality)
				fmt.Print(items.ItemToSql(item, *baselevel, *difficulty))
			}

			if lookupItem.DungeonLevel == 70 && lookupItem.Expansion == 1 {
				Scale(highLevelItem, &item, *itemLevel+10, *item.Quality)
				fmt.Print(items.ItemToSql(item, *baselevel, *difficulty))
			}

			if lookupItem.DungeonLevel < 80 && lookupItem.Expansion == 2 {
				Scale(highLevelItem, &item, *itemLevel+7, *item.Quality)
				fmt.Print(items.ItemToSql(item, *baselevel, *difficulty))
			}

			if lookupItem.DungeonLevel == 80 && lookupItem.Expansion == 2 {
				Scale(highLevelItem, &item, *itemLevel+10, *item.Quality)
				fmt.Print(items.ItemToSql(item, *baselevel+2, *difficulty))
			}
		} else {

			var finalBonus int = 0
			var quality int = 4

			// adjust qualities and levels required based on power and difficulty
			if mysql.IsFinalBoss(lookupItem.CreatureId) {
				fmt.Printf("-- Final Boss Item: %v Entry: %v difficulty %v\n", item.Name, item.Entry, *difficulty)
				finalBonus = 5

				if *difficulty >= 4 {
					quality = 5
				}
			}

			var reqLevel int
			if *difficulty == 4 || *difficulty == 5 {
				reqLevel = *baselevel + 5
			} else {
				reqLevel = *baselevel + 2
			}

			// if the item is from a boss fight
			if lookupItem.DungeonLevel < 60 && lookupItem.Expansion == 0 {
				Scale(highLevelItem, &item, *itemLevel+9+finalBonus, quality)
				fmt.Print(items.ItemToSql(item, reqLevel-1, *difficulty))
			}

			if lookupItem.DungeonLevel == 60 && lookupItem.Expansion == 0 {
				Scale(highLevelItem, &item, *itemLevel+23+finalBonus, quality)
				fmt.Print(items.ItemToSql(item, reqLevel, *difficulty))
			}

			if lookupItem.DungeonLevel < 70 && lookupItem.Expansion == 1 {
				Scale(highLevelItem, &item, *itemLevel+10+finalBonus, quality)
				fmt.Print(items.ItemToSql(item, reqLevel-1, *difficulty))
			}

			if lookupItem.DungeonLevel == 70 && lookupItem.Expansion == 1 {
				Scale(highLevelItem, &item, *itemLevel+23+finalBonus, quality)
				fmt.Print(items.ItemToSql(item, reqLevel, *difficulty))
			}

			if lookupItem.DungeonLevel < 80 && lookupItem.Expansion == 2 {
				Scale(highLevelItem, &item, *itemLevel+12+finalBonus, quality)
				fmt.Print(items.ItemToSql(item, reqLevel-1, *difficulty))
			}

			if lookupItem.DungeonLevel == 80 && lookupItem.Expansion == 2 {
				Scale(highLevelItem, &item, *itemLevel+25+finalBonus, quality)
				fmt.Print(items.ItemToSql(item, reqLevel, *difficulty))
			}
		}

		fmt.Printf("\n -- Item Updated: %v Entry: %v\n", item.Name, item.Entry)
		if itr >= 300 {
			// os.Exit(0)
		}
	}
}

func Scale(highLevelItem mysql.DbItem, item *items.Item, itemLevel, quality int) {
	item.ApplyStats(items.ItemFromDbItem(highLevelItem))
	item.ScaleItem(itemLevel, quality)
	log.Printf("Item Name: %v Stat1: %v Stat2: %v Stat3: %v Stat4: %v Stat5: %v Stat6: %v Stat7: %v Stat8: %v \n",
		item.Name, *item.StatValue1, *item.StatValue2, *item.StatValue3, *item.StatValue4, *item.StatValue5, *item.StatValue6, *item.StatValue7, *item.StatValue8)
}
