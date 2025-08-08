package mysql

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/araxiaonline/endgame-item-generator/internal/config"
)

type DbItem struct {
	Entry          int
	Name           string
	DisplayId      int `db:"displayid"`
	Quality        *int
	ItemLevel      *int `db:"ItemLevel"`
	Class          *int
	Subclass       *int
	Armor          *int     `db:"armor"`
	Material       *int     `db:"material"`
	InventoryType  *int     `db:"inventoryType"`
	AllowableClass *int     `db:"allowableClass"`
	AllowableRace  *int     `db:"allowableRace"`
	RequiredSkill  *int     `db:"requiredSkill"`
	RequiredLevel  *int     `db:"requiredLevel"`
	Durability     *int     `db:"MaxDurability"`
	MinDmg1        *float64 `db:"dmg_min1"`
	MaxDmg1        *float64 `db:"dmg_max1"`
	MinDmg2        *float64 `db:"dmg_min2"`
	MaxDmg2        *float64 `db:"dmg_max2"`
	DmgType1       *int     `db:"dmg_type1"`
	DmgType2       *int     `db:"dmg_type2"`
	Delay          *float64
	Sheath         *int
	StatsCount     *int `db:"statsCount"`
	StatType1      *int `db:"stat_type1"`
	StatValue1     *int `db:"stat_value1"`
	StatType2      *int `db:"stat_type2"`
	StatValue2     *int `db:"stat_value2"`
	StatType3      *int `db:"stat_type3"`
	StatValue3     *int `db:"stat_value3"`
	StatType4      *int `db:"stat_type4"`
	StatValue4     *int `db:"stat_value4"`
	StatType5      *int `db:"stat_type5"`
	StatValue5     *int `db:"stat_value5"`
	StatType6      *int `db:"stat_type6"`
	StatValue6     *int `db:"stat_value6"`
	StatType7      *int `db:"stat_type7"`
	StatValue7     *int `db:"stat_value7"`
	StatType8      *int `db:"stat_type8"`
	StatValue8     *int `db:"stat_value8"`
	StatType9      *int `db:"stat_type9"`
	StatValue9     *int `db:"stat_value9"`
	StatType10     *int `db:"stat_type10"`
	StatValue10    *int `db:"stat_value10"`
	HolyRes        *int `db:"holy_res"`
	FireRes        *int `db:"fire_res"`
	NatureRes      *int `db:"nature_res"`
	FrostRes       *int `db:"frost_res"`
	ShadowRes      *int `db:"shadow_res"`
	ArcaneRes      *int `db:"arcane_res"`
	SpellId1       *int `db:"spellid_1"`
	SpellId2       *int `db:"spellid_2"`
	SpellId3       *int `db:"spellid_3"`
	SpellTrigger1  *int `db:"spelltrigger_1"`
	SpellTrigger2  *int `db:"spelltrigger_2"`
	SpellTrigger3  *int `db:"spelltrigger_3"`
	SocketColor1   *int `db:"socketColor_1"`
	SocketContent1 *int `db:"socketContent_1"`
	SocketColor2   *int `db:"socketColor_2"`
	SocketContent2 *int `db:"socketContent_2"`
	SocketColor3   *int `db:"socketColor_3"`
	SocketContent3 *int `db:"socketContent_3"`
	SocketBonus    *int `db:"socketBonus"`
	GemProperties  *int `db:"GemProperties"`
}

type DbItemCsv struct {
	Entry         int    `csv:"entry"`
	Name          string `csv:"name"`
	DisplayId     int    `csv:"displayid"`
	Quality       *int   `csv:"Quality"`
	ItemLevel     *int   `csv:"ItemLevel"`
	Class         *int   `csv:"class"`
	Subclass      *int   `csv:"subclass"`
	InventoryType *int   `csv:"inventoryType"`
	RequiredLevel *int   `csv:"requiredLevel"`
	StatsCount    *int   `csv:"statsCount"`
	StatType1     *int   `csv:"stat_type1"`
	StatValue1    *int   `csv:"stat_value1"`
	StatType2     *int   `csv:"stat_type2"`
	StatValue2    *int   `csv:"stat_value2"`
	StatType3     *int   `csv:"stat_type3"`
	StatValue3    *int   `csv:"stat_value3"`
	StatType4     *int   `csv:"stat_type4"`
	StatValue4    *int   `csv:"stat_value4"`
	StatType5     *int   `csv:"stat_type5"`
	StatValue5    *int   `csv:"stat_value5"`
	StatType6     *int   `csv:"stat_type6"`
	StatValue6    *int   `csv:"stat_value6"`
	StatType7     *int   `csv:"stat_type7"`
	StatValue7    *int   `csv:"stat_value7"`
	StatType8     *int   `csv:"stat_type8"`
	StatValue8    *int   `csv:"stat_value8"`
	StatType9     *int   `csv:"stat_type9"`
	StatValue9    *int   `csv:"stat_value9"`
	StatType10    *int   `csv:"stat_type10"`
	StatValue10   *int   `csv:"stat_value10"`
}

func (db *MySqlDb) GetItem(entry int) (DbItem, error) {
	if entry == 0 {
		return DbItem{}, fmt.Errorf("entry cannot be 0")
	}

	item := DbItem{}
	sql := "SELECT " + GetItemFields("") + " FROM item_template WHERE entry = ?"
	err := db.Get(&item, sql, entry)
	if err != nil {
		return DbItem{}, err
	}

	return item, nil
}

// Look up a mythic item by name
func (db *MySqlDb) GetByNameAndDifficulty(name string, difficulty int) (DbItem, error) {
	item := DbItem{}
	var min, max int = 0, 0
	if difficulty == 0 {
		return DbItem{}, errors.New("difficulty cannot be 0")
	}

	if difficulty == 3 {
		min = config.MythicItemLevelStart
		max = config.MythicItemLevelEnd
	}

	if difficulty == 4 {
		min = config.LegendaryItemLevelStart
		max = config.LegendaryItemLevelEnd
	}

	if difficulty == 5 {
		min = config.AscendantItemLevelStart
		max = config.AscendantItemLevelEnd
	}

	// though levels are flexible I am going to assume here that mythics are between 300 to 340 these can be overriden in config
	sql := "SELECT " + GetItemFields("") + " FROM item_template WHERE name like ? and ItemLevel >= ? and ItemLevel < ? LIMIT 1"
	name = "%" + name
	err := db.Get(&item, sql, name, min, max)
	if err != nil {
		log.Printf("failed to get item: %v sql: %v", err, fmt.Sprintf("SELECT * FROM item_template WHERE name like %v and ItemLevel >= %v and ItemLevel < %v LIMIT 1", name, min, max))
		return DbItem{}, err
	}

	return item, nil
}

// returns all items from item_template where the quality is between rare and legendary items
func (db *MySqlDb) GetRarePlusItems(limit, offset int) ([]DbItem, error) {
	items := []DbItem{}
	sql := "SELECT " + GetItemFields("") + " FROM item_template WHERE Quality >= 3 and Quality <= 5 and (class = 2 or class = 4) "
	sql += "and subclass != 20 AND entry < 20000000 ORDER BY entry ASC"

	if limit != 0 && offset != 0 {
		sql += fmt.Sprintf("LIMIT %v OFFSET %v", limit, offset)
	}

	err := db.Select(&items, sql)
	if err != nil {
		return []DbItem{}, err
	}

	return items, nil
}

func (db *MySqlDb) GetBossMapItems(mapId int, bossEntries []int, gameObjectEntries []int, limit, offset int) ([]DbItem, error) {
	items := []DbItem{}

	// Build the boss entries condition
	bossEntriesCondition := ""
	if len(bossEntries) > 0 {
		bossEntriesStr := make([]string, len(bossEntries))
		for i, entry := range bossEntries {
			bossEntriesStr[i] = fmt.Sprintf("%d", entry)
		}
		bossEntriesCondition = fmt.Sprintf("OR ct.entry IN (%s)", strings.Join(bossEntriesStr, ","))
	}

	// Build the GameObject entries condition
	gameObjectEntriesCondition := ""
	if len(gameObjectEntries) > 0 {
		gameObjectEntriesStr := make([]string, len(gameObjectEntries))
		for i, entry := range gameObjectEntries {
			gameObjectEntriesStr[i] = fmt.Sprintf("%d", entry)
		}
		gameObjectEntriesCondition = fmt.Sprintf("AND got.entry IN (%s)", strings.Join(gameObjectEntriesStr, ","))
	}

	sql := `SELECT DISTINCT ` + GetItemFields("it") + ` 
FROM acore_world.creature_template ct
LEFT JOIN acore_world.creature c ON c.id1 = ct.entry
LEFT JOIN acore_world.map_dbc m ON c.map = m.ID
LEFT JOIN acore_world.creature_loot_template clt ON ct.lootid = clt.Entry
LEFT JOIN acore_world.reference_loot_template rlt ON clt.Reference = rlt.Entry
LEFT JOIN acore_world.item_template it ON rlt.Item = it.entry

WHERE
    ( m.ID = ? ` + bossEntriesCondition + ` )
    AND ct.rank IN (3)
    AND it.class IN (2, 4)              -- Weapons and armor
    AND it.bonding IN (1, 2)            -- Binds when picked up/equipped
    AND it.Quality >= 4             -- Epic and above`

	// Only add the UNION clause if we have GameObject entries
	if len(gameObjectEntries) > 0 {
		sql += `

UNION

SELECT DISTINCT ` + GetItemFields("it") + ` 

FROM acore_world.gameobject go
JOIN acore_world.gameobject_template got ON go.id = got.entry
LEFT JOIN acore_world.gameobject_loot_template glt ON got.Data1 = glt.Entry
LEFT JOIN acore_world.reference_loot_template rlt ON glt.Reference = rlt.Entry
LEFT JOIN acore_world.item_template it ON rlt.Item = it.entry

WHERE go.map = ? 
    ` + gameObjectEntriesCondition + `
    AND it.class IN (2, 4)              -- Weapons and armor
    AND it.bonding IN (1, 2)            -- Binds when picked up/equipped
    AND it.Quality >= 4`
	}

	if limit != 0 && offset != 0 {
		sql += fmt.Sprintf(" LIMIT %v OFFSET %v", limit, offset)
	}

	// Prepare query parameters
	var args []interface{}
	args = append(args, mapId)
	if len(gameObjectEntries) > 0 {
		args = append(args, mapId) // Second mapId for the UNION query
	}

	err := db.Select(&items, sql, args...)
	if err != nil {
		return []DbItem{}, err
	}

	return items, nil
}

func (db *MySqlDb) GetRaidPhase1Items(class, subclass, limit, offset int) ([]DbItem, error) {
	items := []DbItem{}

	sql := `SELECT DISTINCT ` + GetItemFields("it") + ` 
	FROM acore_world.creature c
	JOIN acore_world.creature_template ct ON c.id1 = ct.entry
	JOIN acore_world.map_dbc m ON c.map = m.ID
	LEFT JOIN acore_world.creature_loot_template clt ON ct.lootid = clt.Entry
	LEFT JOIN acore_world.reference_loot_template rlt ON clt.Reference = rlt.Entry
	LEFT JOIN acore_world.item_template it ON rlt.Item = it.entry

WHERE
    m.ID IN (533,615,616)
    AND ct.rank = 3
    AND it.class = ?              -- Weapons and armor
	AND it.subclass = ?
    AND it.bonding IN (1, 2)            -- Binds when picked up/equipped
    AND it.Quality >= 3              -- Epic and above
`

	if limit != 0 && offset != 0 {
		sql += fmt.Sprintf("LIMIT %v OFFSET %v", limit, offset)
	}

	err := db.Select(&items, sql, class, subclass)
	if err != nil {
		return []DbItem{}, err
	}

	return items, nil
}

func GetItemFields(prefix string) string {
	pre := ""
	if prefix != "" {
		pre = prefix + "."
	}

	return `	
	` + pre + `entry, ` + pre + `name, ` + pre + `displayid,
	quality, ItemLevel, class, subclass, inventoryType,
	allowableClass, allowableRace,
	armor,material,
	holy_res, fire_res, nature_res, frost_res, shadow_res, arcane_res,
	requiredSkill, requiredLevel,
	dmg_min1, dmg_max1,
	dmg_min2,dmg_max2,
	dmg_type1, dmg_type2,
	delay, sheath, MaxDurability,
	statsCount,
	stat_type1, stat_value1,
	stat_type2, stat_value2,
	stat_type3, stat_value3,
	stat_type4, stat_value4,
	stat_type5, stat_value5,
	stat_type6, stat_value6,
	stat_type7, stat_value7,
	stat_type8, stat_value8,
	stat_type9, stat_value9,
	stat_type10, stat_value10,
	spellid_1, spellid_2, spellid_3, 
	spelltrigger_1, spelltrigger_2, spelltrigger_3,
	socketColor_1, socketContent_1, socketColor_2, socketContent_2, socketColor_3, socketContent_3,
	socketBonus, GemProperties`
}

// This will write an DBItem to the database of the specified table..
// It must match the item_template schema
// CopyItem copies an item from one table to another with an optional new ID
// This uses a temporary table approach to handle copying ALL fields without having to list them
// If newId is 0, it keeps the original ID
func (db *MySqlDb) CopyItem(sourceTable string, destTable string, itemEntry int, newEntry int) error {
	// Generate a unique temporary table name using timestamp
	tempTableName := fmt.Sprintf("temp_item_copy_%d", time.Now().UnixNano())

	// Create a temporary table with the same structure as the source table
	sql := fmt.Sprintf("CREATE TEMPORARY TABLE %s LIKE %s", tempTableName, sourceTable)
	_, err := db.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to create temporary table: %w", err)
	}

	// Copy the item to the temporary table
	sql = fmt.Sprintf("INSERT INTO %s SELECT * FROM %s WHERE entry = %d",
		tempTableName, sourceTable, itemEntry)
	_, err = db.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to copy item to temporary table: %w", err)
	}

	// If we need to change the entry, update it in the temporary table
	if newEntry > 0 && newEntry != itemEntry {
		sql = fmt.Sprintf("UPDATE %s SET entry = %d", tempTableName, newEntry)
		_, err = db.Exec(sql)
		if err != nil {
			return fmt.Errorf("failed to update item entry in temporary table: %w", err)
		}
	}

	// Copy from temporary table to destination table
	sql = fmt.Sprintf("REPLACE INTO %s SELECT * FROM %s", destTable, tempTableName)
	_, err = db.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to copy item from temporary table to destination: %w", err)
	}

	// Drop the temporary table
	sql = fmt.Sprintf("DROP TEMPORARY TABLE IF EXISTS %s", tempTableName)
	_, err = db.Exec(sql)
	if err != nil {
		log.Printf("Warning: failed to drop temporary table %s: %v", tempTableName, err)
	}

	return nil
}

func (db *MySqlDb) WriteItem(table string, item DbItem) error {
	// We'll use INSERT ... ON DUPLICATE KEY UPDATE to preserve fields not explicitly set
	// First, build the field list for the INSERT part
	fields := GetItemFields("")

	// Construct the SQL insert statement using the item fields
	sql := "INSERT INTO " + table + " (" + fields + ") VALUES (" +
		"?, ?, ?, " + // entry, name, displayid
		"?, ?, ?, ?, ?, " + // quality, ItemLevel, class, subclass, inventoryType
		"?, ?, " + // allowableClass, allowableRace
		"?, ?, " + // armor, material
		"?, ?, ?, ?, ?, ?, " + // holy_res, fire_res, nature_res, frost_res, shadow_res, arcane_res
		"?, ?, " + // requiredSkill, requiredLevel
		"?, ?, " + // dmg_min1, dmg_max1
		"?, ?, " + // dmg_min2, dmg_max2
		"?, ?, " + // dmg_type1, dmg_type2
		"?, ?, ?, " + // delay, sheath, MaxDurability
		"?, " + // statsCount
		"?, ?, ?, ?, ?, ?, ?, ?, ?, ?, " + // stat_type1-5, stat_value1-5
		"?, ?, ?, ?, ?, ?, ?, ?, ?, ?, " + // stat_type6-10, stat_value6-10
		"?, ?, ?, " + // spellid_1-3
		"?, ?, ?, " + // spelltrigger_1-3
		"?, ?, ?, ?, ?, ?, " + // socketColor_1-3, socketContent_1-3
		"?, ?" + // socketBonus, GemProperties
		") ON DUPLICATE KEY UPDATE " +
		"name = VALUES(name), " +
		"quality = VALUES(quality), " +
		"ItemLevel = VALUES(ItemLevel), " +
		"requiredLevel = VALUES(requiredLevel), " +
		"dmg_min1 = VALUES(dmg_min1), dmg_max1 = VALUES(dmg_max1), " +
		"dmg_min2 = VALUES(dmg_min2), dmg_max2 = VALUES(dmg_max2), " +
		"dmg_type1 = VALUES(dmg_type1), dmg_type2 = VALUES(dmg_type2), " +
		"statsCount = VALUES(statsCount), " +
		"stat_type1 = VALUES(stat_type1), stat_value1 = VALUES(stat_value1), " +
		"stat_type2 = VALUES(stat_type2), stat_value2 = VALUES(stat_value2), " +
		"stat_type3 = VALUES(stat_type3), stat_value3 = VALUES(stat_value3), " +
		"stat_type4 = VALUES(stat_type4), stat_value4 = VALUES(stat_value4), " +
		"stat_type5 = VALUES(stat_type5), stat_value5 = VALUES(stat_value5), " +
		"stat_type6 = VALUES(stat_type6), stat_value6 = VALUES(stat_value6), " +
		"stat_type7 = VALUES(stat_type7), stat_value7 = VALUES(stat_value7), " +
		"stat_type8 = VALUES(stat_type8), stat_value8 = VALUES(stat_value8), " +
		"stat_type9 = VALUES(stat_type9), stat_value9 = VALUES(stat_value9), " +
		"stat_type10 = VALUES(stat_type10), stat_value10 = VALUES(stat_value10), " +
		"spellid_1 = VALUES(spellid_1), spellid_2 = VALUES(spellid_2), spellid_3 = VALUES(spellid_3)"

	// Execute the query with all the item fields as parameters
	_, err := db.Exec(sql,
		item.Entry, item.Name, item.DisplayId,
		item.Quality, item.ItemLevel, item.Class, item.Subclass, item.InventoryType,
		item.AllowableClass, item.AllowableRace,
		item.Armor, item.Material,
		item.HolyRes, item.FireRes, item.NatureRes, item.FrostRes, item.ShadowRes, item.ArcaneRes,
		item.RequiredSkill, item.RequiredLevel,
		item.MinDmg1, item.MaxDmg1,
		item.MinDmg2, item.MaxDmg2,
		item.DmgType1, item.DmgType2,
		item.Delay, item.Sheath, item.Durability,
		item.StatsCount,
		item.StatType1, item.StatValue1,
		item.StatType2, item.StatValue2,
		item.StatType3, item.StatValue3,
		item.StatType4, item.StatValue4,
		item.StatType5, item.StatValue5,
		item.StatType6, item.StatValue6,
		item.StatType7, item.StatValue7,
		item.StatType8, item.StatValue8,
		item.StatType9, item.StatValue9,
		item.StatType10, item.StatValue10,
		item.SpellId1, item.SpellId2, item.SpellId3,
		item.SpellTrigger1, item.SpellTrigger2, item.SpellTrigger3,
		item.SocketColor1, item.SocketContent1,
		item.SocketColor2, item.SocketContent2,
		item.SocketColor3, item.SocketContent3,
		item.SocketBonus, item.GemProperties,
	)

	if err != nil {
		log.Printf("Failed to run sql query: %v", sql)
		return fmt.Errorf("failed to insert item into %s: %w", table, err)
	}

	return nil
}

// This will convert a DbItemCsv to a DbItem and return
// It will first look up the original item in the database though first and populate
// the DbItem with the original item's values then override with the values from the csv
func (db *MySqlDb) ConvertCsvToDbItem(csv DbItemCsv) (DbItem, error) {

	// Try to find the original item in the database
	lookupEntry := csv.Entry - 2000000
	item, err := db.GetItem(lookupEntry)
	if err != nil {
		// Log the error with more details
		log.Printf("Error finding original item with entry %d: %v", lookupEntry, err)

		item, err = db.GetItem(csv.Entry)

		if err != nil {
			log.Printf("Error finding item with direct entry %d: %v", csv.Entry, err)

			// Create a new item with the CSV data
			item = DbItem{
				Entry:     csv.Entry,
				Name:      csv.Name,
				DisplayId: csv.DisplayId,
			}
		}
	}

	// Override values from CSV
	item.Name = csv.Name

	// Override pointer fields if they exist in CSV
	if csv.Quality != nil {
		item.Quality = csv.Quality
	}
	if csv.ItemLevel != nil {
		item.ItemLevel = csv.ItemLevel
	}
	if csv.RequiredLevel != nil {
		item.RequiredLevel = csv.RequiredLevel
	}

	// Handle stats count and stats
	if csv.StatsCount != nil {
		item.StatsCount = csv.StatsCount
	} else if csv.StatType1 != nil {
		// Calculate stats count if not provided but stats exist
		statsCount := 0
		if csv.StatType1 != nil && *csv.StatType1 > 0 {
			statsCount++
		}
		if csv.StatType2 != nil && *csv.StatType2 > 0 {
			statsCount++
		}
		if csv.StatType3 != nil && *csv.StatType3 > 0 {
			statsCount++
		}
		if csv.StatType4 != nil && *csv.StatType4 > 0 {
			statsCount++
		}
		if csv.StatType5 != nil && *csv.StatType5 > 0 {
			statsCount++
		}
		if csv.StatType6 != nil && *csv.StatType6 > 0 {
			statsCount++
		}
		if csv.StatType7 != nil && *csv.StatType7 > 0 {
			statsCount++
		}
		if csv.StatType8 != nil && *csv.StatType8 > 0 {
			statsCount++
		}
		if csv.StatType9 != nil && *csv.StatType9 > 0 {
			statsCount++
		}
		if csv.StatType10 != nil && *csv.StatType10 > 0 {
			statsCount++
		}
		item.StatsCount = &statsCount
	}

	// Override stat types and values
	if csv.StatType1 != nil {
		item.StatType1 = csv.StatType1
	}
	if csv.StatValue1 != nil {
		item.StatValue1 = csv.StatValue1
	}
	if csv.StatType2 != nil {
		item.StatType2 = csv.StatType2
	}
	if csv.StatValue2 != nil {
		item.StatValue2 = csv.StatValue2
	}
	if csv.StatType3 != nil {
		item.StatType3 = csv.StatType3
	}
	if csv.StatValue3 != nil {
		item.StatValue3 = csv.StatValue3
	}
	if csv.StatType4 != nil {
		item.StatType4 = csv.StatType4
	}
	if csv.StatValue4 != nil {
		item.StatValue4 = csv.StatValue4
	}
	if csv.StatType5 != nil {
		item.StatType5 = csv.StatType5
	}
	if csv.StatValue5 != nil {
		item.StatValue5 = csv.StatValue5
	}
	if csv.StatType6 != nil {
		item.StatType6 = csv.StatType6
	}
	if csv.StatValue6 != nil {
		item.StatValue6 = csv.StatValue6
	}
	if csv.StatType7 != nil {
		item.StatType7 = csv.StatType7
	}
	if csv.StatValue7 != nil {
		item.StatValue7 = csv.StatValue7
	}
	if csv.StatType8 != nil {
		item.StatType8 = csv.StatType8
	}
	if csv.StatValue8 != nil {
		item.StatValue8 = csv.StatValue8
	}
	if csv.StatType9 != nil {
		item.StatType9 = csv.StatType9
	}
	if csv.StatValue9 != nil {
		item.StatValue9 = csv.StatValue9
	}
	if csv.StatType10 != nil {
		item.StatType10 = csv.StatType10
	}
	if csv.StatValue10 != nil {
		item.StatValue10 = csv.StatValue10
	}

	return item, nil
}
