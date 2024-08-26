package main

import (
	"database/sql"

	"log"
	"os"
	"strconv"
	"strings"

	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"github.com/araxiaonline/endgame-item-generator/internal/db/sqlite"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func createTable(db *sql.DB) {

	droptable := `DROP TABLE IF EXISTS dungeon_items`
	_, err := db.Exec(droptable)
	if err != nil {
		log.Fatal(err)
	}

	createTable := `CREATE TABLE IF NOT EXISTS dungeon_items (
		entry int unsigned NOT NULL DEFAULT '0',
		mapId tinyint unsigned NOT NULL DEFAULT '0',
		expansion tinyint unsigned NOT NULL DEFAULT '0',
		dungeonLevel tinyint unsigned NOT NULL DEFAULT '0',
		creatureId unsigned NULL DEFAULT NULL,		
		Quality int unsigned NOT NULL DEFAULT '0',		
		PRIMARY KEY (entry)
	  )`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

func ConvertIntSliceToString(slice []int) string {
	sliceStr := make([]string, len(slice))
	for i, v := range slice {
		sliceStr[i] = strconv.Itoa(v)
	}

	return strings.Join(sliceStr, ",")
}

func main() {
	godotenv.Load("../../.env")
	liteDb, err := sql.Open("sqlite3", "../../data/items.db")
	if err != nil {
		log.Fatal(err)
	}
	mysqlDb, err := mysql.Connect(&mysql.MySqlConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer liteDb.Close()
	defer mysqlDb.Close()

	// create the items table if it doesnt exist
	createTable(liteDb)

	// Get all the dungeons and crawl to get all loot and add to sqlite
	dungeons, err := mysql.MySql.GetDungeons(-1)
	if err != nil {
		log.Panicf("failed to get dungeons for expansion %v error: %v", 0, err)
	}

	for _, dungeon := range dungeons {

		log.Printf("+++++Dungeon:  %s ID: %v level %v \n", dungeon.Name, dungeon.Id, dungeon.Level)

		bosses, err := mysql.MySql.GetBosses(dungeon.Id)
		if err != nil {
			log.Fatal("failed to get bosses")
		}

		for _, boss := range bosses {
			dbItems, err := mysql.MySql.GetBossLoot(boss.Entry)

			if err != nil {
				log.Fatalf("failed to get boss loot: %v error:", boss.Name, err)
			}

			for _, dungItem := range dbItems {

				insertItem := sqlite.DungeonItem{
					Entry:        dungItem.Entry,
					MapId:        dungeon.Id,
					Quality:      *dungItem.Quality,
					CreatureId:   boss.Entry,
					Expansion:    dungeon.ExpansionId,
					DungeonLevel: dungeon.Level,
				}
				liteDb.Exec("INSERT OR IGNORE INTO dungeon_items (entry, mapId, expansion, dungeonLevel, creatureId, Quality) VALUES (?, ?, ?, ?, ?, ?)", insertItem.Entry, insertItem.MapId, insertItem.Expansion, insertItem.DungeonLevel, insertItem.CreatureId, insertItem.Quality)
				if err != nil {
					log.Printf("Failed to insert item %v: %v", insertItem.Entry, err)
				}
			}
		}

		dbItems, err := mysql.MySql.GetAddlDungeonDrops(dungeon.Id)
		if err != nil {
			log.Fatalf("failed to get additional dungeon drops: %v error: %v", dungeon.Name, err)
		}

		for _, dungItem := range dbItems {
			insertItem := sqlite.DungeonItem{
				Entry:        dungItem.Entry,
				MapId:        dungeon.Id,
				Quality:      *dungItem.Quality,
				CreatureId:   0,
				Expansion:    dungeon.ExpansionId,
				DungeonLevel: dungeon.Level,
			}
			liteDb.Exec("INSERT OR IGNORE INTO dungeon_items (entry, mapId, expansion, dungeonLevel, creatureId, Quality) VALUES (?, ?, ?, ?, ?, ?)", insertItem.Entry, insertItem.MapId, insertItem.Expansion, insertItem.DungeonLevel, insertItem.CreatureId, insertItem.Quality)
			if err != nil {
				log.Printf("Failed to insert item %v: %v", insertItem.Entry, err)
			}

			log.Printf("+++++Dungeon Item:  %s ID: %v level %v \n", dungItem.Name, dungItem.Entry, dungeon.Level)
		}
	}

}
