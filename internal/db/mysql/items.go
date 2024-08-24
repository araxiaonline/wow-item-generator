package mysql

import (
	"fmt"
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

func (db *MySqlDb) GetRarePlusItems(limit, offset int) ([]DbItem, error) {
	items := []DbItem{}
	sql := "SELECT " + GetItemFields("") + " FROM item_template WHERE Quality >= 3 and Quality <= 5 and (class = 2 or class = 4)"

	if limit != 0 && offset != 0 {
		sql += fmt.Sprintf("LIMIT %v OFFSET %v", limit, offset)
	}

	err := db.Select(&items, sql)
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
