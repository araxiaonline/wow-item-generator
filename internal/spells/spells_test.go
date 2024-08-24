package spells

import (
	"testing"

	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
)

func TestCanBeConverted(t *testing.T) {
	tests := []struct {
		name     string
		spell    Spell
		expected bool
	}{
		{
			name: "Spell with non-aura effect",
			spell: Spell{
				DbSpell: mysql.DbSpell{
					Effect1: 1,
				},
			},
			expected: false,
		},
		{
			name: "Spell with aura effect that can be converted",
			spell: Spell{
				DbSpell: mysql.DbSpell{
					EffectAura1: 8,
				},
			},
			expected: true,
		},
		{
			name: "Spell with mixed effects",
			spell: Spell{
				DbSpell: mysql.DbSpell{
					Effect1:     1,
					EffectAura1: 8,
				},
			},
			expected: false,
		},
		{
			name: "Spell with no effects",
			spell: Spell{
				DbSpell: mysql.DbSpell{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spell.CanBeConverted()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
