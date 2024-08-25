package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

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
	itemLevel := flag.Int("ilvl", 305, "Specify the item level to start scaling from, expansion and difficulty modifiers scale up.")
	difficulty := flag.Int("difficulty", 3, "set the difficulty of the dungeon, defaults to 3 (mythic) 4 (legendary) 5 (ascendant)")
	// levelUp := flag.Bool("levelUp", false, "Boss items require higher +1 level to equip, defaults to false")
	baselevel := flag.Int("baselevel", 80, "set the base level for items to be used, defaults to 80 this is required for levelUp flag")
	flag.Parse()

	if difficulty == nil || *difficulty < 3 || *difficulty > 5 {
		log.Fatal("difficulty must be between 3-5")
		os.Exit(1)
	}

	if itemLevel == nil || *itemLevel < 280 {
		log.Fatal("item level must be greater than 280")
		os.Exit(1)
	}

	if baselevel == nil || *baselevel < 0 {
		log.Fatal("base level must be greater than 80")
		os.Exit(1)
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

	rareItems, err := mysqlDb.GetRarePlusItems(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	for itr, dbItem := range rareItems {
		item := items.ItemFromDbItem(dbItem)

		statsList, err := item.GetStatList()
		if err != nil {
			log.Fatal(err)
			continue
		}

		log.Printf("Item: %v Entry: %v StatsList: %v\n", item.Name, item.Entry, statsList)
		rndItem, err := sqliteDb.GetRandItem(*item.Class, *item.Subclass, statsList, false)
		if err != nil {
			log.Fatal(err)
			continue
		}

		if rndItem == (sqlite.HighLevelItem{}) {
			log.Fatalf("Failed to get random item for %v Entry: %v\n", item.Name, item.Entry)
		}

		log.Printf("Random Item: %v Entry: %v\n", rndItem.Name, rndItem.Entry)

		highLevelItem, err := mysqlDb.GetItem(rndItem.Entry)
		if err != nil {
			log.Fatal(err)
			continue
		}

		// Print all the status for the item that was copied
		// log.Printf("Item Name: %v Stat1: %v Stat2: %v Stat3: %v Stat4: %v Stat5: %v Stat6: %v Stat7: %v Stat8: %v \n",
		// 	item.Entry, *item.StatValue1, *item.StatValue2, *item.StatValue3, *item.StatValue4, *item.StatValue5, *item.StatValue6, *item.StatValue7, *item.StatValue8)

		item.ApplyStats(items.ItemFromDbItem(highLevelItem))

		item.ScaleItem(*itemLevel, 3)
		// log.Printf("Item Name: %v Stat1: %v Stat2: %v Stat3: %v Stat4: %v Stat5: %v Stat6: %v Stat7: %v Stat8: %v \n",
		// 	item.Name, *item.StatValue1, *item.StatValue2, *item.StatValue3, *item.StatValue4, *item.StatValue5, *item.StatValue6, *item.StatValue7, *item.StatValue8)

		fmt.Print(items.ItemToSql(item, 80, 3))

		if itr > 600 {
			break
		}
	}

	// 	// main loop
	// 	dungeons, err := models.DB.GetDungeons(-1)
	// 	if err != nil {
	// 		log.Panicf("failed to get dungeons for expansion %v error: %v", 0, err)
	// 	}

	// 	for _, dungeon := range dungeons {

	// 		log.Printf("+++++Dungeon:  %s ID: %v level %v \n", dungeon.Name, dungeon.Id, dungeon.Level)

	// 		bosses, err := models.DB.GetBosses(dungeon.Id)
	// 		if err != nil {
	// 			log.Fatal("failed to get bosses")
	// 		}

	// 		// Determine the scale value of the item based on expansion and dungeon level
	// 		scaleValue := *itemLevel
	// 		endGameDung := false
	// 		if dungeon.Level == 60 {
	// 			scaleValue += 10
	// 			endGameDung = true
	// 		}

	// 		if dungeon.ExpansionId == 1 && dungeon.Level <= 70 {
	// 			scaleValue += 3
	// 		}

	// 		if dungeon.ExpansionId == 1 && dungeon.Level == 70 {
	// 			scaleValue += 12
	// 			endGameDung = true
	// 		}

	// 		if dungeon.ExpansionId == 2 && dungeon.Level <= 80 {
	// 			scaleValue += 4
	// 		}

	// 		if dungeon.ExpansionId == 2 && dungeon.Level == 80 {
	// 			scaleValue += 15
	// 			endGameDung = true
	// 		}

	// 		// Apply difficuly modifiers for gear scale
	// 		// mythic:      Bosses-Epic Gear (Purple) drops and Rare (Blue) for random drops (BOE)
	// 		// legendary:   Bosses-Epic Gear (Purple) drops and Epic (Purple) for random drops (BOE)
	// 		// ascendant:   Bosses-Legendary Gear (Yellow) drops and Epic (Purple) for random drops (BOE)
	// 		var bossQuality int
	// 		var boeQuality int

	// 		if *difficulty == 4 {
	// 			bossQuality = 4
	// 			boeQuality = 4
	// 		} else if *difficulty == 5 {
	// 			bossQuality = 5
	// 			boeQuality = 4
	// 		} else {
	// 			bossQuality = 4
	// 			boeQuality = 3
	// 		}

	// 		for _, boss := range bosses {

	// 			items, err := models.DB.GetBossLoot(boss.Entry)
	// 			log.Printf("++++++++++ Boss: %s Entry: %v has %v items\n", boss.Name, boss.Entry, len(items))
	// 			if err != nil {
	// 				log.Fatal(err)
	// 				continue
	// 			}

	// 			for _, item := range items {

	// 				_, error := item.ScaleItem(scaleValue, bossQuality)
	// 				if error != nil {
	// 					log.Printf("Failed to scale item: %v", error)
	// 					continue
	// 				}

	// 				fmt.Printf("\n-- Item %v Entry: %v ItemLevel %v \n", item.Name, item.Entry, *item.ItemLevel)
	// 				if *levelUp && endGameDung {
	// 					fmt.Print(ItemToSql(item, *baselevel+1, *difficulty))
	// 				} else {
	// 					fmt.Print(ItemToSql(item, *baselevel, *difficulty))
	// 				}
	// 			}

	// 		}

	// 		items2, err := models.DB.GetAddlDungeonDrops(dungeon.Id)
	// 		if err != nil {
	// 			log.Printf("failed to get additional loot for dungeon %v - err: %v", dungeon.Id, err)
	// 		}

	// 		for _, item := range items2 {

	// 			// reduce item level of dungeon random drops since they are not boss fights
	// 			adjScaleValue := scaleValue - 4

	// 			_, error := item.ScaleItem(adjScaleValue, boeQuality)
	// 			if error != nil {
	// 				log.Printf("Failed to scale item: %v", error)
	// 				continue
	// 			}

	// 			fmt.Printf("\n-- Item %v Entry: %v ItemLevel %v \n", item.Name, item.Entry, *item.ItemLevel)
	// 			fmt.Print(ItemToSql(item, *baselevel, *difficulty))

	// 		}
	// 		log.Printf("++++++++++ Additional Count: %v\n", len(items2))
	// 	}

	// defer models.DB.Close()
}
