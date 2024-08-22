package mysql

import (
	"fmt"

	"github.com/araxiaonline/endgame-item-generator/internal/items"
)

func (db *MySqlDb) GetItem(entry int) (items.Item, error) {
	if entry == 0 {
		return items.Item{}, fmt.Errorf("entry cannot be 0")
	}

	item := items.Item{}
	sql := "SELECT " + GetItemFields("") + " FROM item_template WHERE entry = ?"
	err := db.Get(&item, sql, entry)
	if err != nil {
		return items.Item{}, err
	}

	return item, nil
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
	spelltrigger_1, spelltrigger_2, spelltrigger_3`
}

func ItemToSql(item Item, reqLevel int, difficulty int) string {

	entryBump := 20000000
	spellBump := 30000000
	if difficulty == 4 {
		entryBump = 21000000
	}
	if difficulty == 5 {
		entryBump = 22000000
	}

	if *item.Quality == 4 {
		spellBump = 31000000
	}
	if *item.Quality == 5 {
		spellBump = 32000000
	}

	spells := ""
	if len(item.Spells) > 0 {
		for i, spell := range item.Spells {
			spells.
				spells += spells.SpellToSql(spell, *item.Quality)
			item.UpdateField(fmt.Sprintf("SpellId%v", i), spellBump+spell.ID)
		}
	}

	delete := fmt.Sprintf("DELETE FROM acore_world.item_template WHERE entry = %v;", entryBump+item.Entry)

	clone := fmt.Sprintf(`
	INSERT INTO acore_world.item_template  (
		entry, class, subclass, SoundOverrideSubclass, name, displayid, Quality, Flags, FlagsExtra, BuyCount, 
		BuyPrice, SellPrice, InventoryType, AllowableClass, AllowableRace, ItemLevel, RequiredLevel, 
		RequiredSkill, RequiredSkillRank, requiredspell, requiredhonorrank, RequiredCityRank, 
		RequiredReputationFaction, RequiredReputationRank, maxcount, stackable, ContainerSlots, StatsCount, 
		stat_type1, stat_value1, stat_type2, stat_value2, stat_type3, stat_value3, stat_type4, stat_value4, 
		stat_type5, stat_value5, stat_type6, stat_value6, stat_type7, stat_value7, stat_type8, stat_value8, 
		stat_type9, stat_value9, stat_type10, stat_value10, ScalingStatDistribution, ScalingStatValue, 
		dmg_min1, dmg_max1, dmg_type1, dmg_min2, dmg_max2, dmg_type2, armor, holy_res, fire_res, nature_res, 
		frost_res, shadow_res, arcane_res, delay, ammo_type, RangedModRange, spellid_1, spelltrigger_1, 
		spellcharges_1, spellppmRate_1, spellcooldown_1, spellcategory_1, spellcategorycooldown_1, spellid_2, 
		spelltrigger_2, spellcharges_2, spellppmRate_2, spellcooldown_2, spellcategory_2, spellcategorycooldown_2, 
		spellid_3, spelltrigger_3, spellcharges_3, spellppmRate_3, spellcooldown_3, spellcategory_3, 
		spellcategorycooldown_3, spellid_4, spelltrigger_4, spellcharges_4, spellppmRate_4, spellcooldown_4, 
		spellcategory_4, spellcategorycooldown_4, spellid_5, spelltrigger_5, spellcharges_5, spellppmRate_5, 
		spellcooldown_5, spellcategory_5, spellcategorycooldown_5, bonding, description, PageText, LanguageID, 
		PageMaterial, startquest, lockid, Material, sheath, RandomProperty, RandomSuffix, block, itemset, 
		MaxDurability, area, Map, BagFamily, TotemCategory, socketColor_1, socketContent_1, socketColor_2, 
		socketContent_2, socketColor_3, socketContent_3, socketBonus, GemProperties, RequiredDisenchantSkill, 
		ArmorDamageModifier, duration, ItemLimitCategory, HolidayId, ScriptName, DisenchantID, FoodType, 
		minMoneyLoot, maxMoneyLoot, flagsCustom, VerifiedBuild
	  )
	  SELECT 
		entry + %v, class, subclass, SoundOverrideSubclass, name, displayid, Quality, Flags, FlagsExtra, BuyCount, 
		BuyPrice, SellPrice, InventoryType, AllowableClass, AllowableRace, ItemLevel, RequiredLevel, 
		RequiredSkill, RequiredSkillRank, requiredspell, requiredhonorrank, RequiredCityRank, 
		RequiredReputationFaction, RequiredReputationRank, maxcount, stackable, ContainerSlots, StatsCount, 
		stat_type1, stat_value1, stat_type2, stat_value2, stat_type3, stat_value3, stat_type4, stat_value4, 
		stat_type5, stat_value5, stat_type6, stat_value6, stat_type7, stat_value7, stat_type8, stat_value8, 
		stat_type9, stat_value9, stat_type10, stat_value10, ScalingStatDistribution, ScalingStatValue, 
		dmg_min1, dmg_max1, dmg_type1, dmg_min2, dmg_max2, dmg_type2, armor, holy_res, fire_res, nature_res, 
		frost_res, shadow_res, arcane_res, delay, ammo_type, RangedModRange, spellid_1, spelltrigger_1, 
		spellcharges_1, spellppmRate_1, spellcooldown_1, spellcategory_1, spellcategorycooldown_1, spellid_2, 
		spelltrigger_2, spellcharges_2, spellppmRate_2, spellcooldown_2, spellcategory_2, spellcategorycooldown_2, 
		spellid_3, spelltrigger_3, spellcharges_3, spellppmRate_3, spellcooldown_3, spellcategory_3, 
		spellcategorycooldown_3, spellid_4, spelltrigger_4, spellcharges_4, spellppmRate_4, spellcooldown_4, 
		spellcategory_4, spellcategorycooldown_4, spellid_5, spelltrigger_5, spellcharges_5, spellppmRate_5, 
		spellcooldown_5, spellcategory_5, spellcategorycooldown_5, bonding, description, PageText, LanguageID, 
		PageMaterial, startquest, lockid, Material, sheath, RandomProperty, RandomSuffix, block, itemset, 
		MaxDurability, area, Map, BagFamily, TotemCategory, socketColor_1, socketContent_1, socketColor_2, 
		socketContent_2, socketColor_3, socketContent_3, socketBonus, GemProperties, RequiredDisenchantSkill, 
		ArmorDamageModifier, duration, ItemLimitCategory, HolidayId, ScriptName, DisenchantID, FoodType, 
		minMoneyLoot, maxMoneyLoot, flagsCustom, VerifiedBuild
	  FROM acore_world.item_template as src
	  WHERE src.entry = %v ON DUPLICATE KEY UPDATE entry = src.entry + %v;	  
	`, entryBump, item.Entry, entryBump)

	update := fmt.Sprintf(`
	UPDATE acore_world.item_template
	SET 
	  Quality = %v,
	  ItemLevel = %v,
	  RequiredLevel = %v,
	  dmg_min1 = %v,
	  dmg_max1 = %v,
	  dmg_min2 = %v,
	  dmg_max2 = %v,
	  StatsCount = %v,
	  stat_type1 = %v,
	  stat_value1 = %v,
	  stat_type2 = %v,
	  stat_value2 = %v,
	  stat_type3 = %v,
	  stat_value3 = %v,
	  stat_type4 = %v,
	  stat_value4 = %v,
	  stat_type5 = %v,
	  stat_value5 = %v,
	  stat_type6 = %v,
	  stat_value6 = %v,
	  stat_type7 = %v,
	  stat_value7 = %v,
	  stat_type8 = %v,
	  stat_value8 = %v,
	  stat_type9 = %v,
	  stat_value9 = %v,
	  stat_type10 = %v,
	  stat_value10 = %v,
	  spellid_1 = %v,
	  spellid_2 = %v,
	  spellid_3 = %v,
	  RequiredDisenchantSkill = %v,
	  DisenchantID = %v,
	  SellPrice = FLOOR(100000 + (RAND() * 400001)),
	  Armor = %v
	WHERE entry = %v;
	`, *item.Quality, *item.ItemLevel, reqLevel, *item.MinDmg1, *item.MaxDmg1, *item.MinDmg2, *item.MaxDmg2, *item.StatsCount,
		*item.StatType1, *item.StatValue1, *item.StatType2, *item.StatValue2, *item.StatType3, *item.StatValue3, *item.StatType4, *item.StatValue4,
		*item.StatType5, *item.StatValue5, *item.StatType6, *item.StatValue6, *item.StatType7, *item.StatValue7, *item.StatType8, *item.StatValue8,
		*item.StatType9, *item.StatValue9, *item.StatType10, *item.StatValue10, *item.SpellId1, *item.SpellId2, *item.SpellId3, 375,
		68, *item.Armor, entryBump+item.Entry)

	return fmt.Sprintf("%s %s \n %s \n %s", spells, delete, clone, update)
}
