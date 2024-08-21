// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"slices"
// 	"strconv"
// 	"strings"

// 	// "github.com/araxiaonline/endgame-item-generator/internal/models"
// 	"github.com/araxiaonline/endgame-item-generator/internal/utils"
// 	"github.com/joho/godotenv"
// 	_ "github.com/mattn/go-sqlite3"
// )

// type EndGameItem struct {
// 	Entry     int    `db:"entry"`
// 	Class     int    `db:"class"`
// 	Subclass  int    `db:"subclass"`
// 	StatsList string `db:"stats_list"`
// }

// func createItemsTable(db *sql.DB) {
// 	createItems := `CREATE TABLE IF NOT EXISTS items (
// 		entry int unsigned NOT NULL DEFAULT '0',
// 		class tinyint unsigned NOT NULL DEFAULT '0',
// 		name varchar(250) NOT NULL DEFAULT '',
// 		Quality int unsigned NOT NULL DEFAULT '0',
// 		itemLevel int unsigned NOT NULL DEFAULT '0',
// 		subclass tinyint unsigned NOT NULL DEFAULT '0',
// 		stats_list varchar(250) NOT NULL DEFAULT '',
// 		PRIMARY KEY (entry)
// 	  )`

// 	_, err := db.Exec(createItems)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func ConvertIntSliceToString(slice []int) string {
// 	sliceStr := make([]string, len(slice))
// 	for i, v := range slice {
// 		sliceStr[i] = strconv.Itoa(v)
// 	}

// 	return strings.Join(sliceStr, ",")
// }

// func main() {

// 	liteDb, err := sql.Open("sqlite3", "./items.db")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	godotenv.Load("../../.env")
// 	models.Connect()
// 	sqlDb := models.DB.Client()

// 	defer liteDb.Close()
// 	defer sqlDb.Close()

// 	// create the items table if it doesnt exist
// 	createItemsTable(liteDb)

// 	// create the endgames items lookup table locally for items 200 and above
// 	var items []models.Item
// 	sql := `
// 		SELECT ` + utils.GetItemFields("") + `
// 		from acore_world.item_template
// 		where ItemLevel >= 200 and Quality >= 3 and ItemLevel < 290
// 		AND name NOT LIKE 'NPC Equip%' and name NOT LIKE 'OLD%'
// 		AND name NOT LIKE '%(test)%' AND name NOT LIKE '%Deprecated%'
// 		AND name NOT LIKE  '%Monster - %'
// 		AND ((class = 2 and subclass IN(0,1,2,3,4,5,6,7,8,10,11,12,13,15,16,17,18,19)) or ((class = 4 AND subclass IN (1,2,3,4,6))))
// 	`
// 	err = sqlDb.Select(&items, sql)
// 	if err != nil {
// 		log.Printf("Failed to get items: %v", err)
// 	}

// 	for _, item := range items {
// 		var statsList []int
// 		for i := 1; i <= 10; i++ {
// 			val, _ := item.GetField(fmt.Sprintf("StatValue%v", i))
// 			statId, _ := item.GetField(fmt.Sprintf("StatType%v", i))
// 			if val != 0 {
// 				statsList = append(statsList, statId)
// 			}
// 		}
// 		slices.Sort(statsList)
// 		statsListStr := ConvertIntSliceToString(statsList)
// 		log.Printf("StatList %s for Item %v", statsListStr, item.Name)
// 		_, err = liteDb.Exec("INSERT OR IGNORE INTO items (entry, class, name, Quality, itemLevel, subclass, stats_list) VALUES (?, ?, ?, ?, ?,?,?)", item.Entry, *item.Class, item.Name, *item.Quality, *item.ItemLevel, *item.Subclass, statsListStr)
// 		if err != nil {
// 			log.Printf("Failed to insert item %v: %v", item.Entry, err)
// 		}

// 	}

// 	log.Printf("Items: %v", len(items))
// }
