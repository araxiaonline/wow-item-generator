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

var BossIDs = map[int]bool{
	11520: true, // Taragaman the Hungerer (Ragefire Chasm)
	3654:  true, // Mutanus the Devourer (Wailing Caverns)
	639:   true, // Edwin VanCleef (The Deadmines)
	4275:  true, // Archmage Arugal (Shadowfang Keep)
	4829:  true, // Aku'mai (Blackfathom Deeps)
	1716:  true, // Bazil Thredd (Stormwind Stockade)
	7800:  true, // Mekgineer Thermaplugg (Gnomeregan)
	4421:  true, // Charlga Razorflank (Razorfen Kraul)
	4543:  true, // Bloodmage Thalnos (Scarlet Monastery Graveyard)
	6487:  true, // Arcanist Doan (Scarlet Monastery Library)
	3975:  true, // Herod (Scarlet Monastery Armory)
	3977:  true, // High Inquisitor Whitemane (Scarlet Monastery Cathedral)
	7358:  true, // Amnennar the Coldbringer (Razorfen Downs)
	2748:  true, // Archaedas (Uldaman)
	7267:  true, // Chief Ukorz Sandscalp (Zul'Farrak)
	12201: true, // Princess Theradras (Maraudon)
	8443:  true, // Avatar of Hakkar (Sunken Temple)
	9019:  true, // Emperor Dagran Thaurissan (Blackrock Depths)
	9568:  true, // Overlord Wyrmthalak (Lower Blackrock Spire)
	10363: true, // General Drakkisath (Upper Blackrock Spire)
	11492: true, // Alzzin the Wildshaper (Dire Maul East)
	11489: true, // Tendris Warpwood (Dire Maul West)
	11501: true, // King Gordok (Dire Maul North)
	10440: true, // Baron Rivendare (Stratholme Undead Side)
	10813: true, // Balnazzar (Stratholme Live Side)
	1853:  true, // Darkmaster Gandling (Scholomance)

	17307: true, // Vazruden (Hellfire Ramparts)
	17536: true, // Nazan (Hellfire Ramparts)
	17377: true, // Keli'dan the Breaker (The Blood Furnace)
	16808: true, // Warchief Kargath Bladefist (The Shattered Halls)
	17942: true, // Quagmirran (The Slave Pens)
	17826: true, // Swamplord Musel'ek (The Underbog)
	17798: true, // Warlord Kalithresh (The Steamvault)
	18344: true, // Nexus-Prince Shaffar (Mana-Tombs)
	18373: true, // Exarch Maladaar (Auchenai Crypts)
	18473: true, // Talon King Ikiss (Sethekk Halls)
	18708: true, // Murmur (Shadow Labyrinth)
	19220: true, // Pathaleon the Calculator (The Mechanar)
	17977: true, // Warp Splinter (The Botanica)
	20912: true, // Harbinger Skyriss (The Arcatraz)
	17881: true, // Aeonus (The Black Morass)
	18096: true, // Epoch Hunter (Old Hillsbrad Foothills)
	24664: true, // Kael'thas Sunstrider (Magisters' Terrace)

	23954: true, // Ingvar the Plunderer (Utgarde Keep)
	26723: true, // Keristrasza (The Nexus)
	29120: true, // Anub'arak (Azjol-Nerub)
	29311: true, // Herald Volazj (Ahn'kahet: The Old Kingdom)
	26632: true, // The Prophet Tharon'ja (Drak'Tharon Keep)
	31134: true, // Cyanigosa (Violet Hold)
	29306: true, // Gal'darah (Gundrak)
	27978: true, // Sjonnir the Ironshaper (Halls of Stone)
	28923: true, // Loken (Halls of Lightning)
	27656: true, // Ley-Guardian Eregos (The Oculus)
	26533: true, // Mal'Ganis (Culling of Stratholme)
	26861: true, // King Ymiron (Utgarde Pinnacle)
	35451: true, // The Black Knight (Trial of the Champion)
	36502: true, // Devourer of Souls (Forge of Souls)
	36658: true, // Scourgelord Tyrannus (Pit of Saron)
	37226: true, // The Lich King (Halls of Reflection)

	// Molten Core
	12118: true, // Lucifron
	11982: true, // Magmadar
	12259: true, // Gehennas
	12057: true, // Garr
	12264: true, // Shazzrah
	12056: true, // Baron Geddon
	12098: true, // Sulfuron Harbinger
	11988: true, // Golemagg the Incinerator
	12018: true, // Majordomo Executus
	11502: true, // Ragnaros

	// Blackwing Lair
	12435: true, // Razorgore the Untamed
	13020: true, // Vaelastrasz the Corrupt
	12017: true, // Broodlord Lashlayer
	11983: true, // Firemaw
	14601: true, // Ebonroc
	11981: true, // Flamegor
	14020: true, // Chromaggus
	11583: true, // Nefarian

	// Ruins of Ahn'Qiraj
	15348: true, // Kurinnaxx
	15341: true, // General Rajaxx
	15340: true, // Moam
	15370: true, // Buru the Gorger
	15369: true, // Ayamiss the Hunter
	15339: true, // Ossirian the Unscarred

	// Temple of Ahn'Qiraj
	15263: true, // The Prophet Skeram
	15516: true, // Battleguard Sartura
	15510: true, // Fankriss the Unyielding
	15509: true, // Princess Huhuran
	15275: true, // Emperor Vek'lor
	15276: true, // Emperor Vek'nilash
	15727: true, // C'Thun

	// Zul'Gurub
	14517: true, // High Priestess Jeklik
	14507: true, // High Priest Venoxis
	14510: true, // High Priestess Mar'li
	14509: true, // High Priest Thekal
	14515: true, // High Priestess Arlokk
	14834: true, // Hakkar the Soulflayer
	11382: true, // Bloodlord Mandokir
	11380: true, // Jin'do the Hexxer
	15114: true, // Gahz'ranka
	15082: true, // Renataki
	15083: true, // Grilek
	15084: true, // Hazza'rah
	15085: true, // Wushoolay

	// Karazhan (Full Boss List)
	16152: true, // Attumen the Huntsman (Karazhan)
	15687: true, // Moroes (Karazhan)
	16457: true, // Maiden of Virtue (Karazhan)
	17521: true, // The Big Bad Wolf (Karazhan)
	18168: true, // The Crone (Karazhan)
	17533: true, // Romulo (Karazhan)
	17534: true, // Julianne (Karazhan)
	15691: true, // The Curator (Karazhan)
	15688: true, // Terestian Illhoof (Karazhan)
	16524: true, // Shade of Aran (Karazhan)
	15689: true, // Netherspite (Karazhan)
	16816: true, // Chess Event / Echo of Medivh (Karazhan)
	15690: true, // Prince Malchezaar (Karazhan)
	17225: true, // Nightbane (Karazhan)
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
		WHERE rlt.Entry = ? and it.Quality > 2 
		`
		err = db.Select(&refItems, sql, ref)
		if err != nil {
			return nil, fmt.Errorf("failed to get ref items: %v sql %s", err, sql)
		}

		items = append(items, refItems...)
	}

	return items, nil
}

func IsFinalBoss(bossId int) bool {
	return BossIDs[bossId]
}
