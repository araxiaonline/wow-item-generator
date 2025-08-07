package mysql

import (
	"fmt"
)

type Dungeon struct {
	Id          int    `db:"Id"`
	Name        string `db:"Name"`
	Level       int
	ExpansionId int `db:"ExpansionId"`
}

// dungeon instance id : avg level
var dungeonLevels = map[int]int{

	// Classic WoW dungeons
	389: 18, // Ragefire Chasm
	43:  25, // Wailing Caverns
	36:  23, // The Deadmines
	33:  30, // Shadowfang Keep
	34:  30, // The Stockade
	48:  32, // Blackfathom Deeps
	90:  38, // Gnomeregan
	47:  40, // Razorfen Kraul
	189: 45, // Scarlet Monastery (Graveyard)
	289: 60, // Scholomance
	109: 60, // Sunken Temple
	129: 33, // Razorfen Downs
	70:  40, // Uldaman
	329: 60, // Stratholme
	229: 60, // Blackrock Spire (Lower)
	230: 60, // Blackrock Spire (Upper)
	429: 60, // Dire Maul
	209: 50, // Zul'Farrak
	349: 55, // Maraudon
	269: 57, // Temple of Atal'Hakkar

	// The Burning Crusade dungeons
	540: 70, // Shattered Halls
	542: 65, // Hellfire The Blood Furnace
	543: 62, // Hellfire Ramparts
	545: 64, // Coilfang Steamvaults
	546: 65, // Coilfang Reservoir: The Underbog
	547: 64, // Coilfang Reservoir: The Underbog
	557: 66, // Auchindoun: Mana-Tombs
	558: 67, // Auchindoun: Auchenai Crypts
	556: 70, // Auchindoun: Sethekk Halls
	555: 70, // Auchindoun: Shadow Labyrinth
	560: 68, // Caverns of Time: Old Hillsbrad Foothills
	553: 70, // Tempest Keep: The Botanica
	554: 70, // Tempest Keep: The Mechanar
	552: 70, // Tempest Keep: The Arcatraz
	585: 70, // Magisters' Terrace

	// Wrath of the Lich King dungeons
	574: 72, // Utgarde keep
	575: 76, // Utgarde Pinnacle
	619: 75, // Ahn'kahet: The Old Kingdom
	576: 73, // The Nexus
	595: 80, // The Culling of Stratholme
	600: 76, // Drak'Tharon Keep
	601: 75, // Azjol-Nerub
	608: 77, // The Violet Hold
	604: 78, // Gundrak
	599: 78, // Halls of Stone
	602: 80, // Halls of Lightning
	578: 78, // The Oculus
	650: 80, // Trial of the Champion
	632: 80, // The Forge of Souls
	658: 80, // Pit of Saron
	668: 80, // Halls of Reflection
}

var raidLevels = map[int]int{
	249: 60, // Onyxia's Lair
	409: 60, // Molten Core
	469: 60, // Blackwing Lair
	309: 60, // Zul'Gurub
	531: 60, // Ahn'Qiraj Temple
	509: 60, // Ruins of Ahn'Qiraj
	532: 70, // Karazhan
	548: 70, // Serpentshrine Cavern
	550: 70, // Tempest Keep
	534: 70, // The Battle for Mount Hyjal
	564: 70, // Black Temple
	565: 70, // Gruul's Lair
	568: 70, // Zul'Aman
	580: 70, // Sunwell Plateau
	533: 80, // Naxxramas
	616: 80, // The Eye of Eternity
	615: 80, // The Obsidian Sanctum
	624: 80, // Vault of Archavon
	603: 80, // Ulduar
	649: 80, // Trial of the Crusader
	631: 80, // Icecrown Citadel
	724: 80, // The Ruby Sanctum
}

func (db *MySqlDb) GetDungeons(expansionId int) ([]Dungeon, error) {
	dungeons := []Dungeon{}

	sql := `
		SELECT ID as Id, MapName_Lang_enUS as Name, ExpansionID as ExpansionId 
		FROM map_dbc 
		WHERE InstanceType = 1 AND MapName_Lang_enUS NOT LIKE '%unused%'
	`
	var err error
	if expansionId != -1 {
		sql = sql + "AND ExpansionID = ?"
		err = db.Select(&dungeons, sql, expansionId)
	} else {
		err = db.Select(&dungeons, sql)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get dungeons %v", err)
	}

	for i := range dungeons {
		if level, ok := dungeonLevels[dungeons[i].Id]; ok {
			dungeons[i].Level = level
		}
	}

	return dungeons, nil
}

// Gets a list of other rare+ items that drop in a specific instance.
func (db *MySqlDb) GetAddlDungeonDrops(instanceId int) ([]DbItem, error) {

	fields := GetItemFields("it")
	var items []DbItem
	sql := fmt.Sprintf(`
	SELECT `+fields+`
from
    acore_world.map_dbc m
    join acore_world.creature c on m.ID = c.map
    join acore_world.creature_template ct on c.id1 = ct.entry
    left join acore_world.creature_loot_template clt on ct.lootid = clt.Entry
    left join reference_loot_template rlt on clt.Reference = rlt.Entry
    left join item_template it on rlt.Item = it.entry
WHERE m.ID = %v and Quality >= 3 and it.bonding = 2 and class IN(2,4)

UNION
SELECT `+fields+`
from
    acore_world.map_dbc m
    join acore_world.creature c on m.ID = c.map
    join acore_world.creature_template ct on c.id1 = ct.entry
    left join acore_world.creature_loot_template clt on clt.Entry = ct.Entry
    left join item_template it on clt.Item = it.entry
WHERE m.ID = %v and Quality >= 3 and it.bonding = 2 and it.class IN(2,4)
UNION
SELECT `+fields+`
from
    acore_world.map_dbc m
    join acore_world.gameobject go on m.ID = go.map
    left join acore_world.gameobject_template got on go.id = got.entry
    left join acore_world.gameobject_loot_template glt on glt.Entry = got.Data1
    left join reference_loot_template rlt on glt.Reference = rlt.Entry
    left join item_template it on rlt.Item = it.entry
where m.ID = %v and Quality >=3 and it.bonding IN(1,2) and it.class IN(2,4);
	`, instanceId, instanceId, instanceId)

	// log.Printf("sql: %s", sql)

	err := db.Select(&items, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get additional dungeon items: %v ", err)
	}

	return items, nil
}
