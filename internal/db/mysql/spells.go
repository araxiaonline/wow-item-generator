package mysql

import (
	"fmt"
	"log"
	"strconv"
	"time"
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
    COALESCE(Description_Lang_enUS, '') as Description_Lang_enUS,
    COALESCE(AuraDescription_Lang_enUS, '') as AuraDescription_Lang_enUS,
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
func GetSpellWriteFields() string {
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

// Writes a new upgraded spell to the database for the specified table
func (db *MySqlDb) WriteSpell(table string, spell DbSpell) error {
	// Get the field names from GetSpellFields
	fields := GetSpellWriteFields()

	// Construct the SQL insert statement using the spell fields
	sql := "INSERT INTO " + table + " (" + fields + ") VALUES (" +
		"?, ?, ?, ?, ?, ?, " + // ID, Name_Lang_enUS, Description_Lang_enUS, AuraDescription_Lang_enUS, ProcChance, SpellLevel
		"?, ?, ?, " + // Effect_1, Effect_2, Effect_3
		"?, ?, ?, " + // EffectDieSides_1, EffectDieSides_2, EffectDieSides_3
		"?, ?, ?, " + // EffectRealPointsPerLevel_1, EffectRealPointsPerLevel_2, EffectRealPointsPerLevel_3
		"?, ?, ?, " + // EffectBasePoints_1, EffectBasePoints_2, EffectBasePoints_3
		"?, ?, ?, " + // EffectAura_1, EffectAura_2, EffectAura_3
		"?, ?, ?" + // EffectBonusMultiplier_1, EffectBonusMultiplier_2, EffectBonusMultiplier_3
		") ON DUPLICATE KEY UPDATE " +
		"Name_Lang_enUS = VALUES(Name_Lang_enUS), " +
		"Description_Lang_enUS = VALUES(Description_Lang_enUS), " +
		"AuraDescription_Lang_enUS = VALUES(AuraDescription_Lang_enUS), " +
		"ProcChance = VALUES(ProcChance), " +
		"SpellLevel = VALUES(SpellLevel), " +
		"Effect_1 = VALUES(Effect_1), " +
		"Effect_2 = VALUES(Effect_2), " +
		"Effect_3 = VALUES(Effect_3), " +
		"EffectDieSides_1 = VALUES(EffectDieSides_1), " +
		"EffectDieSides_2 = VALUES(EffectDieSides_2), " +
		"EffectDieSides_3 = VALUES(EffectDieSides_3), " +
		"EffectRealPointsPerLevel_1 = VALUES(EffectRealPointsPerLevel_1), " +
		"EffectRealPointsPerLevel_2 = VALUES(EffectRealPointsPerLevel_2), " +
		"EffectRealPointsPerLevel_3 = VALUES(EffectRealPointsPerLevel_3), " +
		"EffectBasePoints_1 = VALUES(EffectBasePoints_1), " +
		"EffectBasePoints_2 = VALUES(EffectBasePoints_2), " +
		"EffectBasePoints_3 = VALUES(EffectBasePoints_3), " +
		"EffectAura_1 = VALUES(EffectAura_1), " +
		"EffectAura_2 = VALUES(EffectAura_2), " +
		"EffectAura_3 = VALUES(EffectAura_3), " +
		"EffectBonusMultiplier_1 = VALUES(EffectBonusMultiplier_1), " +
		"EffectBonusMultiplier_2 = VALUES(EffectBonusMultiplier_2), " +
		"EffectBonusMultiplier_3 = VALUES(EffectBonusMultiplier_3)"

	// Execute the query with all the spell fields as parameters
	_, err := db.Exec(sql,
		spell.ID,
		spell.Name,
		spell.Description,
		spell.AuraDescription,
		spell.ProcChance,
		spell.SpellLevel,
		spell.Effect1,
		spell.Effect2,
		spell.Effect3,
		spell.EffectDieSides1,
		spell.EffectDieSides2,
		spell.EffectDieSides3,
		spell.EffectRealPointsPerLevel1,
		spell.EffectRealPointsPerLevel2,
		spell.EffectRealPointsPerLevel3,
		spell.EffectBasePoints1,
		spell.EffectBasePoints2,
		spell.EffectBasePoints3,
		spell.EffectAura1,
		spell.EffectAura2,
		spell.EffectAura3,
		spell.EffectBonusMultiplier1,
		spell.EffectBonusMultiplier2,
		spell.EffectBonusMultiplier3,
	)

	if err != nil {
		log.Printf("Failed to run sql query: %v", sql)
		return fmt.Errorf("failed to insert spell into %s: %w", table, err)
	}

	return nil
}

// CopySpell copies a spell from one table to another with an optional new ID
// This uses a temporary table approach to handle copying ALL fields without having to list them
// If newId is 0, it keeps the original ID
func (db *MySqlDb) CopySpell(sourceTable string, destTable string, spellId int, newId int) error {
	// Generate a unique temporary table name using timestamp
	tempTableName := fmt.Sprintf("temp_spell_copy_%d", time.Now().UnixNano())

	// Create a temporary table with the same structure as the source table
	sql := fmt.Sprintf("CREATE TEMPORARY TABLE %s LIKE %s", tempTableName, sourceTable)
	_, err := db.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to create temporary table: %w", err)
	}

	// Copy the spell to the temporary table
	sql = fmt.Sprintf("INSERT INTO %s SELECT * FROM %s WHERE ID = %d",
		tempTableName, sourceTable, spellId)
	_, err = db.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to copy spell to temporary table: %w", err)
	}

	// If we need to change the ID, update it in the temporary table
	if newId > 0 && newId != spellId {
		sql = fmt.Sprintf("UPDATE %s SET ID = %d", tempTableName, newId)
		_, err = db.Exec(sql)
		if err != nil {
			return fmt.Errorf("failed to update spell ID in temporary table: %w", err)
		}
	}

	// Copy from temporary table to destination table
	sql = fmt.Sprintf("INSERT INTO %s SELECT * FROM %s", destTable, tempTableName)
	_, err = db.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to copy spell from temporary table to destination: %w", err)
	}

	// Drop the temporary table (MySQL automatically drops temporary tables at the end of the session,
	// but it's good practice to clean up explicitly)
	sql = fmt.Sprintf("DROP TEMPORARY TABLE IF EXISTS %s", tempTableName)
	_, err = db.Exec(sql)
	if err != nil {
		log.Printf("Warning: failed to drop temporary table %s: %v", tempTableName, err)
	}

	log.Printf("Successfully copied spell %d to %s", spellId, destTable)
	if newId > 0 && newId != spellId {
		log.Printf("  with new ID: %d", newId)
	}

	return nil
}
