package spells

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/araxiaonline/endgame-item-generator/internal/config"
	"github.com/thoas/go-funk"
)

// list of spell effects that require scaling
var SpellEffects = [...]int{
	2,  // School Damage
	6,  // AppyAura
	9,  // HealthLeech
	10, // Heal
	30, // Restores Mana
	35, // Apply Area Aura
}

// list of spell aura effects that require scaling
var SpellAuraEffects = [...]int{
	3,   // DOT
	8,   // HOT
	13,  // Modifies Spell Damage Done
	15,  // Modifies Damage Shield
	22,  // Modifies Resistance
	34,  // Modifies HEalth
	85,  // Modifies Mana Regen
	99,  // Modifies Attack Power
	124, // Modifies Range Attack Power
	135, // Modifies Healing Done
	189, // Modifies Critical Strike
}

// Mapping of spell aura effects to stat types they modify
var AuraEffectsStatMap = map[int]int{
	8:   46,
	13:  45,
	85:  43,
	99:  38,
	124: 38,
	135: 45,
}

// Usually in EffectMiscValueA to describe what the Aura modifies
var SpellModifiers = [...]int{
	0,  // damage
	1,  // duration
	2,  // threat
	3,  // effect1
	4,  // charges
	5,  // range
	6,  // radius
	7,  // crit chance
	8,  // all effects
	9,  // No pushback
	10, // Cast Time
	11, // CD
	12, // effect2
	13, // ignore armor
	14, // cost
	15, // crit damage bonus
	16, // resist miss chance
	17, // jump targets
	18, // Chance of success
	19, // Amplitude
	20, // Dmg multiplier
	21, // GCD
	22, // DoT
	23, // effect3
	24, // bonus multiplier
	26, // PPM
	27, // value multiplier
	28, // resist dispel chance
	29, // crit damage bonus 2
}

// result of a stat conversion from spell to raw stats on item
type ConvItemStat struct {
	StatType  int
	StatValue int
	Budget    int
}

// Spell Effect with max value for effect storage
type SpellEffect struct {
	Effect        int
	BasePoints    int
	DieSides      int
	CalculatedMax int
}

// DB Mapping from spell_dbc
type Spell struct {
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
	ItemSpellSlot             int
	Scaled                    bool
}

func calcMaxValue(base int, sides int) int {
	if base < 0 {
		return base - sides
	}

	return base + sides
}

// get a List of the spell effects (not auras) that need to be scaled
func (s Spell) GetSpellEffects() []SpellEffect {
	effects := make([]SpellEffect, 0)

	effects = append(effects, SpellEffect{
		Effect:        s.Effect1,
		BasePoints:    s.EffectBasePoints1,
		DieSides:      s.EffectDieSides1,
		CalculatedMax: calcMaxValue(s.EffectBasePoints1, s.EffectDieSides1),
	})

	effects = append(effects, SpellEffect{
		Effect:        s.Effect2,
		BasePoints:    s.EffectBasePoints2,
		DieSides:      s.EffectDieSides2,
		CalculatedMax: calcMaxValue(s.EffectBasePoints2, s.EffectDieSides2),
	})

	effects = append(effects, SpellEffect{
		Effect:        s.Effect3,
		BasePoints:    s.EffectBasePoints3,
		DieSides:      s.EffectDieSides3,
		CalculatedMax: calcMaxValue(s.EffectBasePoints3, s.EffectDieSides3),
	})

	return effects
}

// Get he effects and calculate the max value for the a
func (s Spell) GetAuraEffects() []SpellEffect {
	effects := make([]SpellEffect, 0)

	effects = append(effects, SpellEffect{
		Effect:        s.EffectAura1,
		BasePoints:    s.EffectBasePoints1,
		DieSides:      s.EffectDieSides1,
		CalculatedMax: calcMaxValue(s.EffectBasePoints1, s.EffectDieSides1),
	})

	effects = append(effects, SpellEffect{
		Effect:        s.EffectAura2,
		BasePoints:    s.EffectBasePoints2,
		DieSides:      s.EffectDieSides2,
		CalculatedMax: calcMaxValue(s.EffectBasePoints2, s.EffectDieSides2),
	})

	effects = append(effects, SpellEffect{
		Effect:        s.EffectAura3,
		BasePoints:    s.EffectBasePoints3,
		DieSides:      s.EffectDieSides3,
		CalculatedMax: calcMaxValue(s.EffectBasePoints3, s.EffectDieSides3),
	})

	return effects
}

// this spell effect has stats or effects that need to be scaled
func (s Spell) SpellEffectsNeedsScaled() bool {
	if s.Effect1 == 0 {
		return false
	}

	needsScaled := false
	effects := s.GetSpellEffects()
	for _, e := range effects {

		if !funk.Contains(SpellEffects, e.Effect) || e.Effect == 6 {
			continue
		}
		needsScaled = true
	}

	return needsScaled
}

// this aura effect has stats or effects that need to be scaled
func (s Spell) AuraEffectNeedsScaled() bool {
	if s.EffectAura1 == 0 {
		return false
	}

	for _, effect := range SpellAuraEffects {
		if s.EffectAura1 == effect || s.EffectAura2 == effect || s.EffectAura3 == effect {
			return true
		}
	}
	return false
}

func (s Spell) HasAuraEffect() bool {
	return s.EffectAura1 != 0 || s.EffectAura2 != 0 || s.EffectAura3 != 0
}

func AuraEffectCanBeConv(effect int) bool {
	statMods := [...]int{8, 13, 22, 34, 85, 99, 124, 135, 189}
	return funk.Contains(statMods, effect)
}

// Lookup details about the effect and return the stat type -1 indicates not found
func convertAuraEffect(effect int) int {
	if !funk.Contains(AuraEffectsStatMap, effect) {
		log.Printf("effect %v not found in SpellEffectStatMap skipping", effect)
		return -1
	}

	return AuraEffectsStatMap[effect]
}

// Converts spell buffs to item stats making it easier to convert and normalize
func (s Spell) ConvertToStats() ([]ConvItemStat, error) {
	stats := []ConvItemStat{}

	if s.Effect1 == 0 && s.EffectAura1 == 0 {
		return stats, fmt.Errorf("spell does not have an effect1 or auraEffect1")
	}

	effects := s.GetAuraEffects()

	if s.ID == 9397 {
		log.Printf("Spell: %v AuraEffect1: %v AuraEffect2: %v AuraEffect3: %v", s.Name, s.EffectAura1, s.EffectAura2, s.EffectAura3)

	}
	var seen []int
	for _, e := range effects {
		if !AuraEffectCanBeConv(e.Effect) {
			continue
		}

		statId := convertAuraEffect(e.Effect)
		if statId == -1 {
			continue
		}

		if funk.Contains(seen, statId) {
			continue
		}

		// keep track if we have already seen this stat so we do not duplicate
		// Wotlk changed everything to spell power so might as well do the same in
		// scaling process.
		seen = append(seen, statId)
		statMod := float64(config.StatModifiers[statId])
		stats = append(stats, ConvItemStat{
			StatType:  statId,
			StatValue: e.CalculatedMax,
			Budget:    int(math.Abs(math.Ceil(float64(e.CalculatedMax) * statMod))),
		})
	}

	// Handle special stat case where 189 is catch all for crit, dodge, parry, hit, haste, expertise
	if s.Effect1 != 0 && s.Effect1 == 6 && (s.EffectAura1 == 189 || s.EffectAura1 == 123) {
		log.Printf("Special case for spell aura effect: %v", s.Description)
		statId := parseStatDesc(s.Description)
		if statId == 0 {
			log.Printf("Could not determine stat for spell aura effect description: %v", s.Name)
		}

		calced := calcMaxValue(s.EffectBasePoints1, s.EffectDieSides1)
		log.Printf("StatId: %v Calced: %v", statId, calced)
		stats = append(stats, ConvItemStat{
			StatType:  statId,
			StatValue: calced,
			Budget:    int(math.Abs(math.Ceil(float64(calced) * float64(config.StatModifiers[statId])))),
		})
	}

	return stats, nil
}

// This spell can be converted fully into a stat and not needed on the item
func (s Spell) CanBeConverted() bool {

	// if there are any spell effects that are not aura effects, then it can be converted
	effects := s.GetSpellEffects()
	for _, e := range effects {
		if e.Effect != 0 && e.Effect != 6 {
			return false
		}
	}

	// Unfortunately if there are any mixed effects for auras, it is too difficult to split those so just
	// bail out
	auras := s.GetAuraEffects()
	auraFlag := false
	for _, a := range auras {
		if a.Effect == 0 {
			continue
		}

		if AuraEffectCanBeConv(a.Effect) {
			return true
		}
	}

	return auraFlag
}

// based on the description determine the stat to
func parseStatDesc(desc string) int {
	if strings.Contains(desc, "critical strike") {
		return 32
	}

	if strings.Contains(desc, "dodge") {
		return 13
	}

	if strings.Contains(desc, "parry") {
		return 14
	}

	if strings.Contains(desc, "hit rating") {
		return 31
	}

	if strings.Contains(desc, "haste rating") {
		return 36
	}

	if strings.Contains(desc, "expertise rating") {
		return 37
	}

	if strings.Contains(desc, "defense rating") {
		return 12
	}

	if strings.Contains(desc, "block rating") {
		return 15
	}

	if strings.Contains(desc, "armor penetration") {
		return 44
	}

	if strings.Contains(desc, "spell penetration") {
		return 47
	}

	return 0
}

// Scales a spell effect, means creating a new spell with the same effect but scaled to a new item level, then passing
// back the new spellId, In order to be predictable I will use 30000000 for rare, 31000000 for epic, 32000000 for legendary
// An example of this might on hit do $s1 nature damage over $d seconds.  We would just scale the $s1 value
// based on the formula below. This assumes that Blizzard has already balanced the spell bonus against the
// stats on the item level and quality.  This is a big assumption as the stats are not penalized
// from having the extra damage.  This could really create some unique sought after weapons that exploit this.
// modified ratio ((s1 / existing iLevel) * newIlevel) * (0.20 Rare or 0.30 Epic or 0.4 for Legendary).
func (s *Spell) ScaleSpell(fromItemLevel int, itemLevel int, itemQuality int) (int, error) {
	s.Scaled = false
	qualModifier := map[int]float64{
		3: 1.20,
		4: 1.30,
		5: 1.40,
	}

	idBump := 30000000
	if itemQuality == 4 {
		idBump = 31000000
	}
	if itemQuality == 5 {
		idBump = 32000000
	}

	// direct damage types
	dd := [...]int{2, 9, 10}

	didScale := false
	// Causes direct damage
	if s.Effect1 != 0 && funk.Contains(dd, s.Effect1) {
		s.EffectBasePoints1 = int(float64(s.EffectBasePoints1) / float64(fromItemLevel) * float64(itemLevel) * qualModifier[itemQuality] * 2.5)
		didScale = true
	}
	if s.Effect2 != 0 && funk.Contains(dd, s.Effect1) {
		s.EffectBasePoints2 = int(float64(s.EffectBasePoints2) / float64(fromItemLevel) * float64(itemLevel) * qualModifier[itemQuality] * 2.5)
		didScale = true
	}

	// Restores a Power / Mana
	if s.Effect1 != 0 && s.Effect1 == 30 {
		// skip anyhing else that is not mana as they are flat values
		if strings.Contains(s.Description, "Mana") || strings.Contains(s.Description, "mana") {
			s.EffectBasePoints1 = int(float64(s.EffectBasePoints1) / float64(fromItemLevel) * float64(itemLevel) * qualModifier[itemQuality])
			didScale = true
		}
	}

	// Scales a stat buff
	if s.Effect1 != 0 && s.Effect1 == 35 {
		s.EffectBasePoints1 = int(float64(s.EffectBasePoints1) / float64(fromItemLevel) * float64(itemLevel) * qualModifier[itemQuality])
		didScale = true
	}
	if s.Effect1 != 0 && s.Effect2 == 35 {
		s.EffectBasePoints2 = int(float64(s.EffectBasePoints2) / float64(fromItemLevel) * float64(itemLevel) * qualModifier[itemQuality])
		didScale = true
	}

	// Handle special aura effects
	if s.EffectAura1 != 0 && s.EffectAura1 == 3 && s.Effect1 == 6 {
		s.EffectBasePoints1 = int(float64(s.EffectBasePoints1) / float64(fromItemLevel) * float64(itemLevel) * qualModifier[itemQuality] * 2)
		didScale = true
	}

	// Damage Shield Increase Scale due to HP curve
	if s.EffectAura1 != 0 && s.EffectAura1 == 15 && s.Effect1 == 6 {
		s.EffectBasePoints1 = int(float64(s.EffectBasePoints1) / float64(fromItemLevel) * float64(itemLevel) * qualModifier[itemQuality] * 1.50)
		didScale = true
	}

	if !didScale {
		return 0, fmt.Errorf("did not qualify to be scaled in ScaleSpell %v (%v)", s.Name, s.ID)
	}
	s.Scaled = true
	return idBump + s.ID, nil
}
