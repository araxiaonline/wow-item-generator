package spells

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/araxiaonline/endgame-item-generator/internal/config"
	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
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
	29,  // Modifies Strength
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
	29:  4,
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

type Spell struct {
	mysql.DbSpell
	Scaled        bool
	ItemSpellSlot int
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
	return funk.Contains(SpellAuraEffects, effect)
}

// Lookup details about the effect and return the stat type -1 indicates not found
func convertAuraEffect(effect int) int {
	if !funk.Contains(AuraEffectsStatMap, effect) {
		// log.Printf("effect %v not found in SpellEffectStatMap skipping", effect)
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
		// log.Printf("Special case for spell aura effect: %v", s.Description)
		statId := parseStatDesc(s.Description)
		// if statId == 0 {
		// 	// log.Printf("Could not determine stat for spell aura effect description: %v", s.Name)
		// }

		calced := calcMaxValue(s.EffectBasePoints1, s.EffectDieSides1)
		// log.Printf("StatId: %v Calced: %v", statId, calced)
		stats = append(stats, ConvItemStat{
			StatType:  statId,
			StatValue: calced,
			Budget:    int(math.Abs(math.Ceil(float64(calced) * float64(config.StatModifiers[statId])))),
		})
	}

	// if len(stats) == 0 {
	// 	// log.Printf("Failed to Convert Spell to Stats: %v AuraEffect1: %v AuraEffect2: %v AuraEffect3: %v", s.Name, s.EffectAura1, s.EffectAura2, s.EffectAura3)
	// } else {
	// 	// log.Printf("Converted Spell to Stats: %v AuraEffect1: %v AuraEffect2: %v AuraEffect3: %v", s.Name, s.EffectAura1, s.EffectAura2, s.EffectAura3)
	// }

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

func SpellToSql(spell Spell, quality int) string {

	entryBump := 30000000
	if quality == 4 {
		entryBump = 31000000
	}
	if quality == 5 {
		entryBump = 32000000
	}

	insert := fmt.Sprintf(`
	INSERT IGNORE INTO acore_world.spell_dbc (
		ID, Category, DispelType, Mechanic, Attributes, AttributesEx, AttributesEx2, AttributesEx3, AttributesEx4,
		AttributesEx5, AttributesEx6, AttributesEx7, ShapeshiftMask, unk_320_2, ShapeshiftExclude, unk_320_3, Targets,
		TargetCreatureType, RequiresSpellFocus, FacingCasterFlags, CasterAuraState, TargetAuraState, ExcludeCasterAuraState,
		ExcludeTargetAuraState, CasterAuraSpell, TargetAuraSpell, ExcludeCasterAuraSpell, ExcludeTargetAuraSpell, CastingTimeIndex,
		RecoveryTime, CategoryRecoveryTime, InterruptFlags, AuraInterruptFlags, ChannelInterruptFlags, ProcTypeMask, ProcChance,
		ProcCharges, MaxLevel, BaseLevel, SpellLevel, DurationIndex, PowerType, ManaCost, ManaCostPerLevel, ManaPerSecond,
		ManaPerSecondPerLevel, RangeIndex, Speed, ModalNextSpell, CumulativeAura, Totem_1, Totem_2, Reagent_1, Reagent_2, Reagent_3,
		Reagent_4, Reagent_5, Reagent_6, Reagent_7, Reagent_8, ReagentCount_1, ReagentCount_2, ReagentCount_3, ReagentCount_4,
		ReagentCount_5, ReagentCount_6, ReagentCount_7, ReagentCount_8, EquippedItemClass, EquippedItemSubclass, EquippedItemInvTypes,
		Effect_1, Effect_2, Effect_3, EffectDieSides_1, EffectDieSides_2, EffectDieSides_3, EffectRealPointsPerLevel_1,
		EffectRealPointsPerLevel_2, EffectRealPointsPerLevel_3, EffectBasePoints_1, EffectBasePoints_2, EffectBasePoints_3,
		EffectMechanic_1, EffectMechanic_2, EffectMechanic_3, ImplicitTargetA_1, ImplicitTargetA_2, ImplicitTargetA_3, ImplicitTargetB_1,
		ImplicitTargetB_2, ImplicitTargetB_3, EffectRadiusIndex_1, EffectRadiusIndex_2, EffectRadiusIndex_3, EffectAura_1,
		EffectAura_2, EffectAura_3, EffectAuraPeriod_1, EffectAuraPeriod_2, EffectAuraPeriod_3, EffectMultipleValue_1, EffectMultipleValue_2,
		EffectMultipleValue_3, EffectChainTargets_1, EffectChainTargets_2, EffectChainTargets_3, EffectItemType_1, EffectItemType_2,
		EffectItemType_3, EffectMiscValue_1, EffectMiscValue_2, EffectMiscValue_3, EffectMiscValueB_1, EffectMiscValueB_2, EffectMiscValueB_3,
		EffectTriggerSpell_1, EffectTriggerSpell_2, EffectTriggerSpell_3, EffectPointsPerCombo_1, EffectPointsPerCombo_2, EffectPointsPerCombo_3,
		EffectSpellClassMaskA_1, EffectSpellClassMaskA_2, EffectSpellClassMaskA_3, EffectSpellClassMaskB_1, EffectSpellClassMaskB_2,
		EffectSpellClassMaskB_3, EffectSpellClassMaskC_1, EffectSpellClassMaskC_2, EffectSpellClassMaskC_3, SpellVisualID_1, SpellVisualID_2,
		SpellIconID, ActiveIconID, SpellPriority, Name_Lang_enUS, Name_Lang_enGB, Name_Lang_koKR, Name_Lang_frFR, Name_Lang_deDE,
		Name_Lang_enCN, Name_Lang_zhCN, Name_Lang_enTW, Name_Lang_zhTW, Name_Lang_esES, Name_Lang_esMX, Name_Lang_ruRU, Name_Lang_ptPT,
		Name_Lang_ptBR, Name_Lang_itIT, Name_Lang_Unk, Name_Lang_Mask, NameSubtext_Lang_enUS, NameSubtext_Lang_enGB, NameSubtext_Lang_koKR,
		NameSubtext_Lang_frFR, NameSubtext_Lang_deDE, NameSubtext_Lang_enCN, NameSubtext_Lang_zhCN, NameSubtext_Lang_enTW, NameSubtext_Lang_zhTW,
		NameSubtext_Lang_esES, NameSubtext_Lang_esMX, NameSubtext_Lang_ruRU, NameSubtext_Lang_ptPT, NameSubtext_Lang_ptBR, NameSubtext_Lang_itIT,
		NameSubtext_Lang_Unk, NameSubtext_Lang_Mask, Description_Lang_enUS, Description_Lang_enGB, Description_Lang_koKR, Description_Lang_frFR,
		Description_Lang_deDE, Description_Lang_enCN, Description_Lang_zhCN, Description_Lang_enTW, Description_Lang_zhTW, Description_Lang_esES,
		Description_Lang_esMX, Description_Lang_ruRU, Description_Lang_ptPT, Description_Lang_ptBR, Description_Lang_itIT, Description_Lang_Unk,
		Description_Lang_Mask, AuraDescription_Lang_enUS, AuraDescription_Lang_enGB, AuraDescription_Lang_koKR, AuraDescription_Lang_frFR,
		AuraDescription_Lang_deDE, AuraDescription_Lang_enCN, AuraDescription_Lang_zhCN, AuraDescription_Lang_enTW, AuraDescription_Lang_zhTW,
		AuraDescription_Lang_esES, AuraDescription_Lang_esMX, AuraDescription_Lang_ruRU, AuraDescription_Lang_ptPT, AuraDescription_Lang_ptBR,
		AuraDescription_Lang_itIT, AuraDescription_Lang_Unk, AuraDescription_Lang_Mask, ManaCostPct, StartRecoveryCategory, StartRecoveryTime,
		MaxTargetLevel, SpellClassSet, SpellClassMask_1, SpellClassMask_2, SpellClassMask_3, MaxTargets, DefenseType, PreventionType, StanceBarOrder,
		EffectChainAmplitude_1, EffectChainAmplitude_2, EffectChainAmplitude_3, MinFactionID, MinReputation, RequiredAuraVision, RequiredTotemCategoryID_1,
		RequiredTotemCategoryID_2, RequiredAreasID, SchoolMask, RuneCostID, SpellMissileID, PowerDisplayID, EffectBonusMultiplier_1, EffectBonusMultiplier_2,
		EffectBonusMultiplier_3, SpellDescriptionVariableID, SpellDifficultyID
	) SELECT 
	ID + %v, Category, DispelType, Mechanic, Attributes, AttributesEx, AttributesEx2, AttributesEx3, AttributesEx4,
	AttributesEx5, AttributesEx6, AttributesEx7, ShapeshiftMask, unk_320_2, ShapeshiftExclude, unk_320_3, Targets,
	TargetCreatureType, RequiresSpellFocus, FacingCasterFlags, CasterAuraState, TargetAuraState, ExcludeCasterAuraState,
	ExcludeTargetAuraState, CasterAuraSpell, TargetAuraSpell, ExcludeCasterAuraSpell, ExcludeTargetAuraSpell, CastingTimeIndex,
	RecoveryTime, CategoryRecoveryTime, InterruptFlags, AuraInterruptFlags, ChannelInterruptFlags, ProcTypeMask, ProcChance,
	ProcCharges, MaxLevel, BaseLevel, SpellLevel, DurationIndex, PowerType, ManaCost, ManaCostPerLevel, ManaPerSecond,
	ManaPerSecondPerLevel, RangeIndex, Speed, ModalNextSpell, CumulativeAura, Totem_1, Totem_2, Reagent_1, Reagent_2, Reagent_3,
	Reagent_4, Reagent_5, Reagent_6, Reagent_7, Reagent_8, ReagentCount_1, ReagentCount_2, ReagentCount_3, ReagentCount_4,
	ReagentCount_5, ReagentCount_6, ReagentCount_7, ReagentCount_8, EquippedItemClass, EquippedItemSubclass, EquippedItemInvTypes,
	Effect_1, Effect_2, Effect_3, EffectDieSides_1, EffectDieSides_2, EffectDieSides_3, EffectRealPointsPerLevel_1,
	EffectRealPointsPerLevel_2, EffectRealPointsPerLevel_3, EffectBasePoints_1, EffectBasePoints_2, EffectBasePoints_3,
	EffectMechanic_1, EffectMechanic_2, EffectMechanic_3, ImplicitTargetA_1, ImplicitTargetA_2, ImplicitTargetA_3, ImplicitTargetB_1,
	ImplicitTargetB_2, ImplicitTargetB_3, EffectRadiusIndex_1, EffectRadiusIndex_2, EffectRadiusIndex_3, EffectAura_1,
	EffectAura_2, EffectAura_3, EffectAuraPeriod_1, EffectAuraPeriod_2, EffectAuraPeriod_3, EffectMultipleValue_1, EffectMultipleValue_2,
	EffectMultipleValue_3, EffectChainTargets_1, EffectChainTargets_2, EffectChainTargets_3, EffectItemType_1, EffectItemType_2,
	EffectItemType_3, EffectMiscValue_1, EffectMiscValue_2, EffectMiscValue_3, EffectMiscValueB_1, EffectMiscValueB_2, EffectMiscValueB_3,
	EffectTriggerSpell_1, EffectTriggerSpell_2, EffectTriggerSpell_3, EffectPointsPerCombo_1, EffectPointsPerCombo_2, EffectPointsPerCombo_3,
	EffectSpellClassMaskA_1, EffectSpellClassMaskA_2, EffectSpellClassMaskA_3, EffectSpellClassMaskB_1, EffectSpellClassMaskB_2,
	EffectSpellClassMaskB_3, EffectSpellClassMaskC_1, EffectSpellClassMaskC_2, EffectSpellClassMaskC_3, SpellVisualID_1, SpellVisualID_2,
	SpellIconID, ActiveIconID, SpellPriority, Name_Lang_enUS, Name_Lang_enGB, Name_Lang_koKR, Name_Lang_frFR, Name_Lang_deDE,
	Name_Lang_enCN, Name_Lang_zhCN, Name_Lang_enTW, Name_Lang_zhTW, Name_Lang_esES, Name_Lang_esMX, Name_Lang_ruRU, Name_Lang_ptPT,
	Name_Lang_ptBR, Name_Lang_itIT, Name_Lang_Unk, Name_Lang_Mask, NameSubtext_Lang_enUS, NameSubtext_Lang_enGB, NameSubtext_Lang_koKR,
	NameSubtext_Lang_frFR, NameSubtext_Lang_deDE, NameSubtext_Lang_enCN, NameSubtext_Lang_zhCN, NameSubtext_Lang_enTW, NameSubtext_Lang_zhTW,
	NameSubtext_Lang_esES, NameSubtext_Lang_esMX, NameSubtext_Lang_ruRU, NameSubtext_Lang_ptPT, NameSubtext_Lang_ptBR, NameSubtext_Lang_itIT,
	NameSubtext_Lang_Unk, NameSubtext_Lang_Mask, Description_Lang_enUS, Description_Lang_enGB, Description_Lang_koKR, Description_Lang_frFR,
	Description_Lang_deDE, Description_Lang_enCN, Description_Lang_zhCN, Description_Lang_enTW, Description_Lang_zhTW, Description_Lang_esES,
	Description_Lang_esMX, Description_Lang_ruRU, Description_Lang_ptPT, Description_Lang_ptBR, Description_Lang_itIT, Description_Lang_Unk,
	Description_Lang_Mask, AuraDescription_Lang_enUS, AuraDescription_Lang_enGB, AuraDescription_Lang_koKR, AuraDescription_Lang_frFR,
	AuraDescription_Lang_deDE, AuraDescription_Lang_enCN, AuraDescription_Lang_zhCN, AuraDescription_Lang_enTW, AuraDescription_Lang_zhTW,
	AuraDescription_Lang_esES, AuraDescription_Lang_esMX, AuraDescription_Lang_ruRU, AuraDescription_Lang_ptPT, AuraDescription_Lang_ptBR,
	AuraDescription_Lang_itIT, AuraDescription_Lang_Unk, AuraDescription_Lang_Mask, ManaCostPct, StartRecoveryCategory, StartRecoveryTime,
	MaxTargetLevel, SpellClassSet, SpellClassMask_1, SpellClassMask_2, SpellClassMask_3, MaxTargets, DefenseType, PreventionType, StanceBarOrder,
	EffectChainAmplitude_1, EffectChainAmplitude_2, EffectChainAmplitude_3, MinFactionID, MinReputation, RequiredAuraVision, RequiredTotemCategoryID_1,
	RequiredTotemCategoryID_2, RequiredAreasID, SchoolMask, RuneCostID, SpellMissileID, PowerDisplayID, EffectBonusMultiplier_1, EffectBonusMultiplier_2,
	EffectBonusMultiplier_3, SpellDescriptionVariableID, SpellDifficultyID from acore_world.spell_dbc as src
	WHERE src.ID = %v ON DUPLICATE KEY UPDATE ID = src.ID + %v;`, entryBump, spell.ID, entryBump)

	update := fmt.Sprintf(`
	UPDATE acore_world.spell_dbc
	SET EffectBasePoints_1 = %v, EffectBasePoints_2 = %v
	WHERE ID = %v;`, spell.EffectBasePoints1, spell.EffectBasePoints2, entryBump+spell.ID)

	return fmt.Sprintf("\n %s \n %s \n", insert, update)
}
