package items

import (
	"testing"

	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
)

func TestGetClassUserType(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected int
	}{
		// Tank items (return 6) - Priority stats: Parry, Defense, Block
		{
			name: "Tank item with Parry Rating",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     1,
					Name:      "Tank Weapon",
					Class:     ptrInt(2), // Weapon
					Subclass:  ptrInt(1), // One-handed sword
					StatType1: ptrInt(STAT.ParryRating),
					StatValue1: ptrInt(20),
				},
			},
			expected: 6,
		},
		{
			name: "Tank item with Defense Rating",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     2,
					Name:      "Tank Shield",
					Class:     ptrInt(4), // Armor
					Subclass:  ptrInt(4), // Plate
					StatType1: ptrInt(STAT.DefenseSkillRating),
					StatValue1: ptrInt(15),
				},
			},
			expected: 6,
		},
		{
			name: "Tank item with Block Rating",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     3,
					Name:      "Tank Shield",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.BlockRating),
					StatValue1: ptrInt(25),
				},
			},
			expected: 6,
		},
		{
			name: "Tank item with Block Value",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     4,
					Name:      "Tank Shield",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.BlockValue),
					StatValue1: ptrInt(30),
				},
			},
			expected: 6,
		},

		// Healer items (return 5) - Priority stats: MP5, Spell Healing Done
		{
			name: "Healer item with Mana Regeneration",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     5,
					Name:      "Healer Staff",
					Class:     ptrInt(2), // Weapon
					StatType1: ptrInt(STAT.ManaRegeneration),
					StatValue1: ptrInt(10),
				},
			},
			expected: 5,
		},
		{
			name: "Healer item with Spell Healing Done",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     6,
					Name:      "Healer Robe",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.SpellHealingDone),
					StatValue1: ptrInt(50),
				},
			},
			expected: 5,
		},

		// Mage items (return 4) - Priority stat: Spell Penetration
		{
			name: "Mage item with Spell Penetration",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     7,
					Name:      "Mage Wand",
					Class:     ptrInt(2), // Weapon
					StatType1: ptrInt(STAT.SpellPenetration),
					StatValue1: ptrInt(15),
				},
			},
			expected: 4,
		},

		// Ranged Attacker items (return 3) - Priority stats: Ranged Attack Power, Ranged Crit, Ranged Hit
		{
			name: "Ranged item with Ranged Attack Power",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     8,
					Name:      "Hunter Bow",
					Class:     ptrInt(2), // Weapon
					StatType1: ptrInt(STAT.RangedAttackPower),
					StatValue1: ptrInt(40),
				},
			},
			expected: 3,
		},
		{
			name: "Ranged item with Ranged Crit Rating",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     9,
					Name:      "Hunter Crossbow",
					Class:     ptrInt(2), // Weapon
					StatType1: ptrInt(STAT.CritRangedRating),
					StatValue1: ptrInt(20),
				},
			},
			expected: 3,
		},
		{
			name: "Ranged item with Ranged Hit Rating",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     10,
					Name:      "Hunter Gun",
					Class:     ptrInt(2), // Weapon
					StatType1: ptrInt(STAT.HitRangedRating),
					StatValue1: ptrInt(15),
				},
			},
			expected: 3,
		},

		// Cloth armor classification (return 4 for mage)
		{
			name: "Cloth armor without healer stats",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:        11,
					Name:         "Cloth Robe",
					Class:        ptrInt(4), // Armor
					Subclass:     ptrInt(1), // Cloth
					InventoryType: ptrInt(5), // Not off-hand (16)
					StatType1:    ptrInt(STAT.Intellect),
					StatValue1:   ptrInt(20),
				},
			},
			expected: 4,
		},

		// Plate armor classification (return 1 for strength melee)
		{
			name: "Plate armor",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     12,
					Name:      "Plate Chest",
					Class:     ptrInt(4), // Armor
					Subclass:  ptrInt(4), // Plate
					StatType1: ptrInt(STAT.Strength),
					StatValue1: ptrInt(25),
				},
			},
			expected: 1,
		},

		// Mail/Leather armor with spell stats (return 4 for mage)
		{
			name: "Mail armor with spell power",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     13,
					Name:      "Mail Vest",
					Class:     ptrInt(4), // Armor
					Subclass:  ptrInt(3), // Mail
					StatType1: ptrInt(STAT.SpellPower),
					StatValue1: ptrInt(30),
				},
			},
			expected: 4,
		},
		{
			name: "Leather armor with spell crit",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     14,
					Name:      "Leather Tunic",
					Class:     ptrInt(4), // Armor
					Subclass:  ptrInt(2), // Leather
					StatType1: ptrInt(STAT.CritSpellRating),
					StatValue1: ptrInt(15),
				},
			},
			expected: 4,
		},

		// Mail/Leather armor without spell stats (return 2 for agility)
		{
			name: "Mail armor without spell stats",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     15,
					Name:      "Mail Armor",
					Class:     ptrInt(4), // Armor
					Subclass:  ptrInt(3), // Mail
					StatType1: ptrInt(STAT.Agility),
					StatValue1: ptrInt(20),
				},
			},
			expected: 2,
		},

		// Weapon classifications
		{
			name: "Fist weapon (agility)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     16,
					Name:      "Fist Weapon",
					Class:     ptrInt(2),  // Weapon
					Subclass:  ptrInt(13), // Fist weapon
					StatType1: ptrInt(STAT.Agility),
					StatValue1: ptrInt(15),
				},
			},
			expected: 2,
		},
		{
			name: "Throwing weapon (agility)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     17,
					Name:      "Throwing Knife",
					Class:     ptrInt(2),  // Weapon
					Subclass:  ptrInt(16), // Thrown
					StatType1: ptrInt(STAT.Agility),
					StatValue1: ptrInt(12),
				},
			},
			expected: 2,
		},
		{
			name: "Wand (mage)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     18,
					Name:      "Magic Wand",
					Class:     ptrInt(2),  // Weapon
					Subclass:  ptrInt(19), // Wand
					StatType1: ptrInt(STAT.SpellPower),
					StatValue1: ptrInt(25),
				},
			},
			expected: 4,
		},

		// Polearm/Spear with strength (return 1)
		{
			name: "Polearm with strength",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     19,
					Name:      "Strength Polearm",
					Class:     ptrInt(2),  // Weapon
					Subclass:  ptrInt(17), // Polearm
					StatType1: ptrInt(STAT.Strength),
					StatValue1: ptrInt(30),
				},
			},
			expected: 1,
		},
		{
			name: "Spear with attack power",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     20,
					Name:      "Attack Power Spear",
					Class:     ptrInt(2), // Weapon
					Subclass:  ptrInt(6), // Spear
					StatType1: ptrInt(STAT.AttackPower),
					StatValue1: ptrInt(50),
				},
			},
			expected: 1,
		},

		// Polearm/Spear with agility (return 2)
		{
			name: "Polearm with agility",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     21,
					Name:      "Agility Polearm",
					Class:     ptrInt(2),  // Weapon
					Subclass:  ptrInt(17), // Polearm
					StatType1: ptrInt(STAT.Agility),
					StatValue1: ptrInt(25),
				},
			},
			expected: 2,
		},

		// Polearm/Spear fallback (return 5 for healer)
		{
			name: "Polearm fallback to healer",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     22,
					Name:      "Generic Polearm",
					Class:     ptrInt(2),  // Weapon
					Subclass:  ptrInt(17), // Polearm
					StatType1: ptrInt(STAT.Stamina),
					StatValue1: ptrInt(20),
				},
			},
			expected: 5,
		},

		// Bow/Gun/Crossbow with strength (return 1)
		{
			name: "Bow with strength",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     23,
					Name:      "Strength Bow",
					Class:     ptrInt(2), // Weapon
					Subclass:  ptrInt(2), // Bow
					StatType1: ptrInt(STAT.Strength),
					StatValue1: ptrInt(20),
				},
			},
			expected: 1,
		},

		// Bow/Gun/Crossbow fallback (return 3 for ranged)
		{
			name: "Bow fallback to ranged",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     24,
					Name:      "Generic Bow",
					Class:     ptrInt(2), // Weapon
					Subclass:  ptrInt(2), // Bow
					StatType1: ptrInt(STAT.Stamina),
					StatValue1: ptrInt(15),
				},
			},
			expected: 3,
		},

		// Base stat priority tests
		{
			name: "Item with Spirit (healer priority)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     25,
					Name:      "Spirit Item",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.Spirit),
					StatValue1: ptrInt(15),
				},
			},
			expected: 5,
		},
		{
			name: "Item with Intellect (mage priority)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     26,
					Name:      "Intellect Item",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.Intellect),
					StatValue1: ptrInt(20),
				},
			},
			expected: 4,
		},
		{
			name: "Item with Strength (strength melee priority)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     27,
					Name:      "Strength Item",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.Strength),
					StatValue1: ptrInt(25),
				},
			},
			expected: 1,
		},
		{
			name: "Item with Agility (agility melee priority)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     28,
					Name:      "Agility Item",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.Agility),
					StatValue1: ptrInt(18),
				},
			},
			expected: 2,
		},

		// Secondary stat tests
		{
			name: "Item with Attack Power (agility melee)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     29,
					Name:      "Attack Power Item",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.AttackPower),
					StatValue1: ptrInt(40),
				},
			},
			expected: 2,
		},
		{
			name: "Item with Melee Haste (agility melee)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     30,
					Name:      "Melee Haste Item",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.HasteMeleeRating),
					StatValue1: ptrInt(15),
				},
			},
			expected: 2,
		},
		{
			name: "Item with Spell Power (mage)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     31,
					Name:      "Spell Power Item",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.SpellPower),
					StatValue1: ptrInt(35),
				},
			},
			expected: 4,
		},
		{
			name: "Item with Ranged Attack Power (ranged)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     32,
					Name:      "Ranged AP Item",
					Class:     ptrInt(4), // Armor
					StatType1: ptrInt(STAT.RangedAttackPower),
					StatValue1: ptrInt(30),
				},
			},
			expected: 3,
		},

		// Armor type fallback tests
		{
			name: "Cloth armor fallback",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     33,
					Name:      "Generic Cloth",
					Class:     ptrInt(4), // Armor
					Subclass:  ptrInt(1), // Cloth
					StatType1: ptrInt(STAT.Stamina),
					StatValue1: ptrInt(20),
				},
			},
			expected: 4,
		},
		{
			name: "Plate armor fallback",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     34,
					Name:      "Generic Plate",
					Class:     ptrInt(4), // Armor
					Subclass:  ptrInt(4), // Plate
					StatType1: ptrInt(STAT.Stamina),
					StatValue1: ptrInt(25),
				},
			},
			expected: 1,
		},
		{
			name: "Leather armor fallback",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     35,
					Name:      "Generic Leather",
					Class:     ptrInt(4), // Armor
					Subclass:  ptrInt(2), // Leather
					StatType1: ptrInt(STAT.Stamina),
					StatValue1: ptrInt(18),
				},
			},
			expected: 2,
		},

		// Generic fallback (return 7)
		{
			name: "Generic item with no identifiable stats",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     36,
					Name:      "Generic Item",
					Class:     ptrInt(15), // Miscellaneous
					StatType1: ptrInt(STAT.Stamina),
					StatValue1: ptrInt(10),
				},
			},
			expected: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.GetClassUserType()
			if result != tt.expected {
				t.Errorf("GetClassUserType() = %v, expected %v for item: %s", result, tt.expected, tt.item.Name)
			}
		})
	}
}

// Test edge cases and complex scenarios
func TestGetClassUserTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected int
	}{
		{
			name: "Item with multiple stats - tank priority wins",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:      100,
					Name:       "Multi-stat Tank Item",
					Class:      ptrInt(4), // Armor
					StatType1:  ptrInt(STAT.Strength),
					StatValue1: ptrInt(20),
					StatType2:  ptrInt(STAT.ParryRating), // Tank stat should win
					StatValue2: ptrInt(15),
					StatType3:  ptrInt(STAT.Agility),
					StatValue3: ptrInt(18),
				},
			},
			expected: 6, // Tank
		},
		{
			name: "Item with multiple stats - healer priority wins",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:      101,
					Name:       "Multi-stat Healer Item",
					Class:      ptrInt(4), // Armor
					StatType1:  ptrInt(STAT.Intellect),
					StatValue1: ptrInt(25),
					StatType2:  ptrInt(STAT.ManaRegeneration), // Healer stat should win
					StatValue2: ptrInt(8),
					StatType3:  ptrInt(STAT.SpellPower),
					StatValue3: ptrInt(30),
				},
			},
			expected: 5, // Healer
		},
		{
			name: "Cloth off-hand item (should not be classified as mage)",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:         102,
					Name:          "Cloth Off-hand",
					Class:         ptrInt(4),  // Armor
					Subclass:      ptrInt(1),  // Cloth
					InventoryType: ptrInt(16), // Off-hand
					StatType1:     ptrInt(STAT.Intellect),
					StatValue1:    ptrInt(20),
				},
			},
			expected: 4, // Still mage due to intellect in base stats
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.GetClassUserType()
			if result != tt.expected {
				t.Errorf("GetClassUserType() = %v, expected %v for item: %s", result, tt.expected, tt.item.Name)
			}
		})
	}
}
