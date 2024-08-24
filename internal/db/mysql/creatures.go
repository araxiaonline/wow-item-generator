package mysql

import (
	"errors"
	"fmt"
)

type Boss struct {
	Entry              int
	Name               string
	ScriptName         string `db:"ScriptName"`
	ExperienceModifier int    `db:"ExperienceModifier"`
}

func (db *MySqlDb) GetBosses(mapId int) ([]Boss, error) {

	if mapId == 0 {
		return nil, errors.New("mapId cannot be 0")
	}

	bosses := []Boss{}
	var sql string

	// 540 is pre-classic dungeons so XP Multiplier is best way to determine bosses / rare mobs
	if mapId < 540 {
		sql = `
			SELECT ct.entry, ct.name, ct.ScriptName, ct.ExperienceModifier from acore_world.creature c
			JOIN acore_world.creature_template ct ON(c.id1 = ct.entry)  WHERE map = ? and ExperienceModifier >= 2;
		`
	} else {
		sql = `
			SELECT ct.entry, ct.name, ct.ScriptName, ct.ExperienceModifier from acore_world.creature c
    		JOIN acore_world.creature_template ct ON(c.id1 = ct.entry)  WHERE map = ? and ct.ScriptName Like 'boss_%'
		`
	}

	err := db.Select(&bosses, sql, mapId)
	if err != nil {
		return nil, err
	}

	return bosses, nil
}

func (db *MySqlDb) GetBossLoot(bossId int) ([]DbItem, error) {
	if bossId == 0 {
		return nil, errors.New("bossId cannot be 0")
	}

	// This will first find items that are not in the reference boss loot table
	items := []DbItem{}
	fields := GetItemFields("")
	sql := `
	SELECT ` + fields + `
	from acore_world.item_template 
	where 
	entry in
		(SELECT item from acore_world.creature_loot_template where entry = ? and GroupId != 0 and Reference = 0)
	and Quality > 2
	`

	udb := db.Unsafe()
	err := udb.Select(&items, sql, bossId)
	if err != nil {
		return nil, err
	}

	// Get all the boss reference items now
	var references []int
	sql = `
		SELECT reference 
		FROM acore_world.creature_loot_template
		WHERE entry = ? AND Reference != 0
	`
	err = db.Select(&references, sql, bossId)
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %v sql %s", err, sql)
	}

	if len(references) == 0 {
		return items, nil
	}

	refItems := []DbItem{}

	// For each reference we now need to get the items and add them to the items slice
	for _, ref := range references {
		sql = `
		SELECT ` + GetItemFields("it") + ` 
		FROM acore_world.reference_loot_template rlt 
		  JOIN acore_world.item_template it ON rlt.Item = it.entry 
		WHERE rlt.Entry = ? and it.Quality > 2 and it.StatsCount > 0
		`
		err = db.Select(&refItems, sql, ref)
		if err != nil {
			return nil, fmt.Errorf("failed to get ref items: %v sql %s", err, sql)
		}

		items = append(items, refItems...)
	}

	return items, nil
}
