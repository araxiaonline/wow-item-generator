package mysql

import (
	"fmt"
	"strconv"
)

type DbSpell struct {
	ID                        int    `db:"ID"`
	Name                      string `db:"Name_Lang_enUS"`
	Description               string `db:"Description_Lang_enUS"`
	AuraDescription           string `db:"AuraDescription_Lang_enUS"`
	ProcChance                int    `db:"ProcChance"`
	SpellLevel                int    `db:"SpellLevel"`
	Effect1                   int    `db:"Effect_1"`
	Effect2                   int    `db:"Effect_2"`
	Effect3                   int    `db:"Effect_3"`
	EffectDieSides1           int    `db:"EffectDieSides_1"`
	EffectDieSides2           int    `db:"EffectDieSides_2"`
	EffectDieSides3           int    `db:"EffectDieSides_3"`
	EffectRealPointsPerLevel1 int    `db:"EffectRealPointsPerLevel_1"`
	EffectRealPointsPerLevel2 int    `db:"EffectRealPointsPerLevel_2"`
	EffectRealPointsPerLevel3 int    `db:"EffectRealPointsPerLevel_3"`
	EffectBasePoints1         int    `db:"EffectBasePoints_1"`
	EffectBasePoints2         int    `db:"EffectBasePoints_2"`
	EffectBasePoints3         int    `db:"EffectBasePoints_3"`
	EffectAura1               int    `db:"EffectAura_1"`
	EffectAura2               int    `db:"EffectAura_2"`
	EffectAura3               int    `db:"EffectAura_3"`
	EffectBonusMultiplier1    int    `db:"EffectBonusMultiplier_1"`
	EffectBonusMultiplier2    int    `db:"EffectBonusMultiplier_2"`
	EffectBonusMultiplier3    int    `db:"EffectBonusMultiplier_3"`
}

func (db *MySqlDb) GetSpell(id int) (DbSpell, error) {

	if id == 0 {
		return DbSpell{}, fmt.Errorf("id cannot be 0")
	}

	spell := DbSpell{}
	sql := "SELECT " + GetSpellFields() + " FROM `spell_dbc` WHERE ID = ? -- " + strconv.Itoa(id)

	err := db.Get(&spell, sql, id)
	if err != nil {
		return DbSpell{}, fmt.Errorf("failed to get spell: %v", err)
	}

	return spell, nil
}

func GetSpellFields() string {
	return `
	ID,
    Name_Lang_enUS,
    Description_Lang_enUS,
    AuraDescription_Lang_enUS,
    ProcChance,
    SpellLevel,
    Effect_1,
    Effect_2,
    Effect_3,
    EffectDieSides_1,
    EffectDieSides_2,
    EffectDieSides_3,
    EffectRealPointsPerLevel_1,
    EffectRealPointsPerLevel_2,
    EffectRealPointsPerLevel_3,
    EffectBasePoints_1,
    EffectBasePoints_2,
    EffectBasePoints_3,
    EffectAura_1,
    EffectAura_2,
    EffectAura_3,
    EffectBonusMultiplier_1,
    EffectBonusMultiplier_2,
    EffectBonusMultiplier_3	
	`
}
