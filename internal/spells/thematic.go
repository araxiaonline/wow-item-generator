package spells

import (
	"fmt"
	"log"
	"strings"

	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
)

// SpellTheme represents different magical themes for spell assignment
type SpellTheme int

const (
	ThemeNone SpellTheme = iota
	ThemeFrost
	ThemeFire
	ThemeNature
	ThemeShadow
	ThemeArcane
	ThemeHoly
)

// String returns the string representation of a SpellTheme
func (t SpellTheme) String() string {
	switch t {
	case ThemeFrost:
		return "Frost"
	case ThemeFire:
		return "Fire"
	case ThemeNature:
		return "Nature"
	case ThemeShadow:
		return "Shadow"
	case ThemeArcane:
		return "Arcane"
	case ThemeHoly:
		return "Holy"
	default:
		return "None"
	}
}

// ThematicSpellAssigner handles assignment of thematic spells to items
type ThematicSpellAssigner struct {
	db *mysql.MySqlDb
}

// NewThematicSpellAssigner creates a new thematic spell assigner
func NewThematicSpellAssigner(db *mysql.MySqlDb) *ThematicSpellAssigner {
	return &ThematicSpellAssigner{db: db}
}

// ShouldHaveProc determines if an item should have a proc spell based on its properties
func (tsa *ThematicSpellAssigner) ShouldHaveProc(itemClass, itemSubclass, itemLevel, quality int, itemName string) bool {
	log.Printf("DEBUG: ShouldHaveProc - Class: %d, Subclass: %d, Level: %d, Quality: %d, Name: %s", 
		itemClass, itemSubclass, itemLevel, quality, itemName)
	
	// Weapons should have procs
	if itemClass == 2 {
		log.Printf("DEBUG: Item is a weapon (class 2)")
		// Higher quality items more likely to have procs
		if quality >= 3 && itemLevel >= 40 {
			log.Printf("DEBUG: Weapon qualifies - quality %d >= 3 and level %d >= 40", quality, itemLevel)
			return true
		} else {
			log.Printf("DEBUG: Weapon does not qualify - quality %d < 3 or level %d < 40", quality, itemLevel)
		}
	}

	// Trinkets should have procs
	if itemClass == 4 && itemSubclass == 0 {
		log.Printf("DEBUG: Item is a trinket - qualifies for proc")
		return true
	}

	// Jewelry with thematic names
	if itemClass == 4 && (itemSubclass == 1 || itemSubclass == 2) { // Neck, Ring
		theme := tsa.DetectThemeFromName(itemName)
		if theme != ThemeNone {
			log.Printf("DEBUG: Jewelry with theme %s - qualifies for proc", theme.String())
			return true
		}
	}

	log.Printf("DEBUG: Item does not qualify for proc")
	return false
}

// DetectThemeFromName analyzes item name and description to determine magical theme
func (tsa *ThematicSpellAssigner) DetectThemeFromName(itemName string) SpellTheme {
	name := strings.ToLower(itemName)

	// Frost theme keywords - spells like Frostbolt, Frost Nova, Ice Lance
	frostKeywords := []string{
		"frost", "cold", "ice", "frozen", "chill", "winter", "arctic", "glacial",
		"coldrage", "frostbite", "icicle", "blizzard", "freeze", "shiver", "crystal",
		"permafrost", "hoarfrost", "rimefang", "icecrown", "northrend", "wintergrasp",
		"frostmourne", "sindragosa", "arthas", "lich", "undead", "scourge",
	}

	// Fire theme keywords - spells like Fireball, Flame Strike, Pyroblast
	fireKeywords := []string{
		"fire", "flame", "burn", "ember", "inferno", "molten", "volcanic",
		"blazing", "scorching", "flaming", "ignite", "pyre", "magma", "lava",
		"sulfuron", "ragnaros", "flamewaker", "firelands", "phoenix", "immolate",
		"conflagrate", "incinerate", "combustion", "pyroblast", "scorch",
	}

	// Nature theme keywords - spells like Lightning Bolt, Earth Shock, Entangling Roots
	natureKeywords := []string{
		"nature", "earth", "storm", "lightning", "thunder", "wind", "leaf",
		"thorn", "root", "growth", "poison", "venom", "toxic", "acid", "druid",
		"shaman", "elemental", "totem", "spirit", "wolf", "bear", "cat", "tree",
		"hurricane", "typhoon", "earthquake", "rockbiter", "windfury", "stormstrike",
		"maelstrom", "chain", "healing", "rejuvenation", "regrowth", "lifebloom",
	}

	// Shadow theme keywords - spells like Shadow Bolt, Drain Life, Fear
	shadowKeywords := []string{
		"shadow", "dark", "void", "curse", "doom", "death", "soul", "spirit",
		"nightmare", "corruption", "plague", "blight", "necro", "warlock", "priest",
		"undead", "bone", "skull", "grave", "tomb", "crypt", "wraith", "specter",
		"banshee", "lich", "drain", "siphon", "vampiric", "life", "mana", "burn",
		"fear", "horror", "terror", "agony", "torment", "suffering", "pain",
	}

	// Arcane theme keywords - spells like Arcane Missiles, Polymorph, Counterspell
	arcaneKeywords := []string{
		"arcane", "magic", "mystic", "enchant", "spell", "mana", "ethereal",
		"astral", "cosmic", "temporal", "dimensional", "mage", "wizard", "sorcerer",
		"intellect", "wisdom", "knowledge", "study", "tome", "grimoire", "staff",
		"orb", "crystal", "gem", "power", "energy", "force", "teleport", "portal",
		"polymorph", "counterspell", "dispel", "silence", "interrupt",
	}

	// Holy theme keywords - spells like Heal, Flash of Light, Consecration
	holyKeywords := []string{
		"holy", "divine", "blessed", "sacred", "light", "radiant", "celestial",
		"righteous", "pure", "sanctified", "hallowed", "paladin", "priest", "cleric",
		"angel", "seraph", "cherub", "heaven", "paradise", "salvation", "redemption",
		"consecration", "blessing", "grace", "mercy", "faith", "devotion", "prayer",
		"heal", "cure", "restore", "renew", "resurrect", "revive", "sanctuary",
	}

	// Check each theme
	for _, keyword := range frostKeywords {
		if strings.Contains(name, keyword) {
			return ThemeFrost
		}
	}

	for _, keyword := range fireKeywords {
		if strings.Contains(name, keyword) {
			return ThemeFire
		}
	}

	for _, keyword := range natureKeywords {
		if strings.Contains(name, keyword) {
			return ThemeNature
		}
	}

	for _, keyword := range shadowKeywords {
		if strings.Contains(name, keyword) {
			return ThemeShadow
		}
	}

	for _, keyword := range arcaneKeywords {
		if strings.Contains(name, keyword) {
			return ThemeArcane
		}
	}

	for _, keyword := range holyKeywords {
		if strings.Contains(name, keyword) {
			return ThemeHoly
		}
	}

	return ThemeNone
}

// FindThematicSpell finds an appropriate spell for the given theme
func (tsa *ThematicSpellAssigner) FindThematicSpell(theme SpellTheme, itemClass, itemSubclass int) (int, error) {
	log.Printf("Finding thematic spell for theme: %s, class: %d, subclass: %d", theme.String(), itemClass, itemSubclass)

	// First, try to find existing scaled spells (31000000+)
	scaledSpellId, err := tsa.findScaledThematicSpell(theme, itemClass, itemSubclass)
	if err == nil && scaledSpellId > 0 {
		log.Printf("Found existing scaled spell: %d for theme %s", scaledSpellId, theme.String())
		return scaledSpellId, nil
	}

	// If no scaled spell found, look in spell_dbc for base spells
	baseSpellId, err := tsa.findBaseThematicSpell(theme, itemClass, itemSubclass)
	if err == nil && baseSpellId > 0 {
		log.Printf("Found base spell: %d for theme %s", baseSpellId, theme.String())
		return baseSpellId, nil
	}

	return 0, fmt.Errorf("no suitable spell found for theme %s", theme.String())
}

// findScaledThematicSpell looks for existing scaled spells (31000000+) that match the theme
func (tsa *ThematicSpellAssigner) findScaledThematicSpell(theme SpellTheme, itemClass, itemSubclass int) (int, error) {
	// Define spell pools for each theme (existing scaled spells starting from 31000000)
	// These would be spells that have already been scaled by the system
	scaledSpellPools := map[SpellTheme][]int{
		ThemeFrost: {
			// Start with high-priority frost spells for weapons/trinkets
			31013439, // Scaled Frostbolt (damage + slow)
			31012737, // Scaled Frost Armor (defensive)
			31021401, // Scaled Frost Nova (AoE slow)
		},
		ThemeFire: {
			// Fire damage spells
			31013140, // Scaled Fireball (direct damage)
			31008413, // Scaled Fire Nova (AoE damage)
			31011564, // Scaled Flame Strike (ground effect)
		},
		ThemeNature: {
			// Nature/Lightning spells
			31000421, // Scaled Lightning Bolt (instant damage)
			31008050, // Scaled Earth Shock (damage + interrupt)
			31010414, // Scaled Entangling Roots (root effect)
		},
		ThemeShadow: {
			// Shadow/Dark magic spells
			31000980, // Scaled Shadow Bolt (shadow damage)
			31007648, // Scaled Drain Life (damage + heal)
			31011671, // Scaled Shadow Word: Pain (DoT)
		},
		ThemeArcane: {
			// Arcane magic spells
			31005143, // Scaled Arcane Missiles (channeled damage)
			31001953, // Scaled Blink (teleport)
			31000118, // Scaled Polymorph (crowd control)
		},
		ThemeHoly: {
			// Holy/Light spells
			31000139, // Scaled Heal (direct healing)
			31010060, // Scaled Flash of Light (fast heal)
			31001244, // Scaled Divine Favor (buff)
		},
	}

	spells, exists := scaledSpellPools[theme]
	if !exists || len(spells) == 0 {
		return 0, fmt.Errorf("no scaled spells available for theme %s", theme.String())
	}

	// For weapons, prefer damage spells; for trinkets/jewelry, prefer utility
	if itemClass == 2 { // Weapons
		// Return first damage spell (they're ordered by preference)
		return spells[0], nil
	}

	// For trinkets/jewelry, prefer utility spells (later in list)
	if len(spells) > 1 {
		return spells[len(spells)-1], nil
	}

	return spells[0], nil
}

// findBaseThematicSpell looks in spell_dbc for base spells that match the theme
func (tsa *ThematicSpellAssigner) findBaseThematicSpell(theme SpellTheme, itemClass, itemSubclass int) (int, error) {
	// Define base spell pools for each theme
	baseSpellPools := map[SpellTheme][]int{
		ThemeFrost: {
			13439, // Frostbolt
			21401, // Frost Nova
			12737, // Frost Armor
			11113, // Blast Wave (Frost)
		},
		ThemeFire: {
			13140, // Fireball
			8413,  // Fire Nova
			11564, // Flame Strike
			11129, // Combustion
		},
		ThemeNature: {
			421,   // Lightning Bolt
			8050,  // Earth Shock
			10414, // Entangling Roots
			16689, // Nature's Grasp
		},
		ThemeShadow: {
			980,   // Shadow Bolt
			7648,  // Drain Life
			11671, // Shadow Word: Pain
			18265, // Siphon Soul
		},
		ThemeArcane: {
			5143, // Arcane Missiles
			1953, // Blink
			118,  // Polymorph
			12051, // Evocation
		},
		ThemeHoly: {
			139,   // Heal
			1244,  // Divine Favor
			10060, // Flash of Light
			10308, // Hammer of Justice
		},
	}

	spells, exists := baseSpellPools[theme]
	if !exists || len(spells) == 0 {
		return 0, fmt.Errorf("no base spells available for theme %s", theme.String())
	}

	// Verify spell exists in database
	for _, spellId := range spells {
		_, err := tsa.db.GetSpell(spellId)
		if err == nil {
			return spellId, nil
		}
	}

	return 0, fmt.Errorf("no valid base spells found for theme %s", theme.String())
}

// AssignThematicSpell assigns a thematic spell to an item if appropriate
func (tsa *ThematicSpellAssigner) AssignThematicSpell(itemEntry, itemClass, itemSubclass, itemLevel, quality int, itemName string) (int, error) {
	// Check if item should have a proc
	if !tsa.ShouldHaveProc(itemClass, itemSubclass, itemLevel, quality, itemName) {
		return 0, fmt.Errorf("item does not qualify for proc assignment")
	}

	// Detect theme from name
	theme := tsa.DetectThemeFromName(itemName)
	if theme == ThemeNone {
		// For weapons without clear theme, don't assign a spell
		return 0, fmt.Errorf("no clear theme detected for item: %s", itemName)
	}

	// Find appropriate spell
	spellId, err := tsa.FindThematicSpell(theme, itemClass, itemSubclass)
	if err != nil {
		return 0, fmt.Errorf("failed to find thematic spell: %v", err)
	}

	log.Printf("Assigned spell %d (theme: %s) to item %s (%d)", spellId, theme.String(), itemName, itemEntry)
	return spellId, nil
}
