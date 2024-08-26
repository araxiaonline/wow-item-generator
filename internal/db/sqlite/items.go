package sqlite

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
)

type HighLevelItem struct {
	Entry     int    `db:"entry"`
	Class     *int   `db:"class"`
	Name      string `db:"name"`
	Quality   int    `db:"Quality"`
	ItemLevel int    `db:"itemLevel"`
	Subclass  *int   `db:"subclass"`
	StatsList string `db:"stats_list"`
}

type DungeonItem struct {
	Entry        int `db:"entry"`
	MapId        int `db:"mapId"`
	CreatureId   int `db:"creatureId"`
	Quality      int `db:"Quality"`
	Expansion    int `db:"expansion"`
	DungeonLevel int `db:"dungeonLevel"`
}

func (db *SqlLite) GetItem(entry int) (HighLevelItem, error) {
	if entry == 0 {
		return HighLevelItem{}, fmt.Errorf("entry cannot be 0")
	}

	item := HighLevelItem{}
	sql := "SELECT " + mysql.GetItemFields("") + " FROM item_template WHERE entry = ?"
	err := db.Get(&item, sql, entry)
	if err != nil {
		return HighLevelItem{}, err
	}

	return item, nil
}

// This gets a random item that is close in stats type to the lower level items with some randomness
func (db *SqlLite) GetRandItem(class, subclass int, statsList []int, end bool) (HighLevelItem, error) {
	rndItem := HighLevelItem{}
	var statsTxt string
	var err error
	var sql string

	// if we have a stats_list try to match by that first, if not then just select a random item from the class and subclass
	if len(statsList) == 0 {
		sql = "SELECT * FROM items WHERE class = ? and subclass = ? ORDER BY RANDOM() LIMIT 1"
		err = db.Get(&rndItem, sql, class, subclass)
	} else {
		// convert the array of ints to a commas string for lookup
		statsTxt = intSliceToString(statsList)
		sql = "SELECT * FROM items WHERE class = ? and subclass = ? and stats_list like ? ORDER BY RANDOM() LIMIT 1"
		err = db.Get(&rndItem, sql, class, subclass, statsTxt+"%")
	}

	if err != nil {

		// if we hit the last check and still no item match then return an error
		if end {
			msg := fmt.Sprintf("Failed to find a matching item class %v subclass %v statsTxt %v", class, subclass, statsTxt)
			return HighLevelItem{}, errors.New(msg)
		}

		// if there was not a remove the last stat and try again
		if err.Error() == "sql: no rows in result set" {

			if len(statsList) == 0 {
				return db.GetRandItem(class, subclass, statsList, true)
			} else {
				statsList = statsList[:len(statsList)-1]
				return db.GetRandItem(class, subclass, statsList, false)
			}

		}

		log.Fatalf("Error getting random sql: %v error: %v", sql, err)
		return HighLevelItem{}, err
	}

	return rndItem, nil
}

func (db *SqlLite) GetItemFromDungeon(itemEntry int) (DungeonItem, error) {
	item := DungeonItem{}
	sql := "SELECT * FROM dungeon_items WHERE entry = ?"

	err := db.Get(&item, sql, itemEntry)
	if err != nil {
		return item, err
	}

	return item, nil
}

func intSliceToString(slice []int) string {
	str := fmt.Sprint(slice)
	str = strings.Trim(str, "[]")
	return strings.ReplaceAll(str, " ", ",")
}
