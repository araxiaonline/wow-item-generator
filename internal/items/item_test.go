package items

import (
	"io"
	"log"
	"math"
	"reflect"
	"testing"

	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"golang.org/x/exp/rand"
)

func TestGetPrimaryStat(t *testing.T) {

	originalLog := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalLog)

	tests := []struct {
		name        string
		item        Item
		wantStat    int
		wantValue   int
		expectError bool
	}{
		{
			name: "No primary stat found",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     1,
					Name:      "Test Item",
					StatType1: ptrInt(1), StatValue1: ptrInt(10),
					StatType2: ptrInt(2), StatValue2: ptrInt(20),
					StatType3: ptrInt(12), StatValue3: ptrInt(15),
				},
			},
			wantStat:    0,
			wantValue:   0,
			expectError: false,
		},
		{
			name: "Primary stat found with higher value",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     1,
					Name:      "Test Item",
					StatType1: ptrInt(3), StatValue1: ptrInt(10), // Agility
					StatType2: ptrInt(4), StatValue2: ptrInt(20), // Strength
					StatType3: ptrInt(5), StatValue3: ptrInt(15), // Intellect
				},
			},
			wantStat:    4, // Strength
			wantValue:   20,
			expectError: false,
		},
		{
			name: "Primary stat found with lower value",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     1,
					Name:      "Test Item",
					StatType1: ptrInt(3), StatValue1: ptrInt(30), // Agility
					StatType2: ptrInt(4), StatValue2: ptrInt(20), // Strength
					StatType3: ptrInt(5), StatValue3: ptrInt(15), // Intellect
				},
			},
			wantStat:    3, // Agility
			wantValue:   30,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStat, gotValue, err := tt.item.GetPrimaryStat()
			if (err != nil) != tt.expectError {
				t.Errorf("GetPrimaryStat() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if gotStat != tt.wantStat {
				t.Errorf("GetPrimaryStat() gotStat = %v, want %v", gotStat, tt.wantStat)
			}
			if gotValue != tt.wantValue {
				t.Errorf("GetPrimaryStat() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestGetStatList(t *testing.T) {
	originalLog := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalLog)

	tests := []struct {
		name        string
		item        Item
		want        []int
		expectError bool
	}{
		{
			name: "No stats available",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     1,
					Name:      "Test Item",
					StatType1: ptrInt(0), StatValue1: ptrInt(0),
					StatType2: ptrInt(0), StatValue2: ptrInt(0),
				},
			},
			want:        []int{},
			expectError: false,
		},
		{
			name: "Multiple stats available",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     1,
					Name:      "Test Item",
					StatType1: ptrInt(3), StatValue1: ptrInt(10), // Agility
					StatType2: ptrInt(4), StatValue2: ptrInt(20), // Strength
					StatType3: ptrInt(5), StatValue3: ptrInt(15), // Intellect
				},
			},
			want:        []int{3, 4, 5},
			expectError: false,
		},
		{
			name: "Stats are ordered correctly",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     1,
					Name:      "Test Item",
					StatType1: ptrInt(7), StatValue1: ptrInt(10), // Agility
					StatType2: ptrInt(4), StatValue2: ptrInt(20), // Strength
					StatType3: ptrInt(31), StatValue3: ptrInt(15), // Intellect
				},
			},
			want:        []int{4, 7, 31},
			expectError: false,
		},
		{
			name: "Some stats are zero",
			item: Item{
				DbItem: mysql.DbItem{
					Entry:     1,
					Name:      "Test Item",
					StatType1: ptrInt(3), StatValue1: ptrInt(0), // Agility
					StatType2: ptrInt(4), StatValue2: ptrInt(20), // Strength
					StatType3: ptrInt(5), StatValue3: ptrInt(0), // Intellect
				},
			},
			want:        []int{4},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.item.GetStatList()
			if (err != nil) != tt.expectError {
				t.Errorf("GetStatList() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStatList() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDPS(t *testing.T) {
	tests := []struct {
		name        string
		item        Item
		wantDPS     float64
		expectError bool
	}{
		{
			name: "Valid DPS calculation",
			item: Item{
				DbItem: mysql.DbItem{
					MinDmg1: ptrFloat64(50),
					MaxDmg1: ptrFloat64(70),
					Delay:   ptrFloat64(3000),
				},
			},
			wantDPS:     20.00,
			expectError: false,
		},
		{
			name: "High damage DPS calculation",
			item: Item{
				DbItem: mysql.DbItem{
					MinDmg1: ptrFloat64(100),
					MaxDmg1: ptrFloat64(150),
					Delay:   ptrFloat64(2000),
				},
			},
			wantDPS:     62.50,
			expectError: false,
		},
		{
			name: "Low damage DPS calculation",
			item: Item{
				DbItem: mysql.DbItem{
					MinDmg1: ptrFloat64(10),
					MaxDmg1: ptrFloat64(15),
					Delay:   ptrFloat64(1500),
				},
			},
			wantDPS:     8.33,
			expectError: false,
		},
		{
			name: "Missing MinDmg1",
			item: Item{
				DbItem: mysql.DbItem{
					MaxDmg1: ptrFloat64(70),
					Delay:   ptrFloat64(3000),
				},
			},
			wantDPS:     0,
			expectError: true,
		},
		{
			name: "Missing MaxDmg1",
			item: Item{
				DbItem: mysql.DbItem{
					MinDmg1: ptrFloat64(50),
					Delay:   ptrFloat64(3000),
				},
			},
			wantDPS:     0,
			expectError: true,
		},
		{
			name: "Missing Delay",
			item: Item{
				DbItem: mysql.DbItem{
					MinDmg1: ptrFloat64(50),
					MaxDmg1: ptrFloat64(70),
				},
			},
			wantDPS:     0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDPS, err := tt.item.GetDPS()
			if (err != nil) != tt.expectError {
				t.Errorf("GetDPS() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && !almostEqual(gotDPS, tt.wantDPS, 0.01) {
				t.Errorf("GetDPS() = %v, want %v", gotDPS, tt.wantDPS)
			}
		})
	}
}

func TestScaleDPS(t *testing.T) {
	tests := []struct {
		name        string
		item        Item
		level       int
		wantDPSMin  float64
		wantDPSMax  float64
		expectError bool
	}{
		{
			name: "Valid Scale DPS calculation",
			item: Item{
				DbItem: mysql.DbItem{
					ItemLevel: ptrInt(60),
					Delay:     ptrFloat64(3000),
					MinDmg1:   ptrFloat64(50),
					MaxDmg1:   ptrFloat64(70),
					Subclass:  ptrInt(4), // One-handed weapon
					Quality:   ptrInt(3), // Rare
				},
			},
			level:       70,
			wantDPSMin:  53.0, // Expected DPS range due to randomness
			wantDPSMax:  107.0,
			expectError: false,
		},
		{
			name: "High level Scale DPS calculation",
			item: Item{
				DbItem: mysql.DbItem{
					ItemLevel: ptrInt(80),
					Delay:     ptrFloat64(2000),
					MinDmg1:   ptrFloat64(150),
					MaxDmg1:   ptrFloat64(200),
					Subclass:  ptrInt(17), // Two-handed weapon
					Quality:   ptrInt(4),  // Epic
				},
			},
			level:       100,
			wantDPSMin:  120.0, // Expected DPS range due to randomness
			wantDPSMax:  240.0,
			expectError: false,
		},
		{
			name: "Low level Scale DPS calculation",
			item: Item{
				DbItem: mysql.DbItem{
					ItemLevel: ptrInt(20),
					Delay:     ptrFloat64(1000),
					MinDmg1:   ptrFloat64(30),
					MaxDmg1:   ptrFloat64(50),
					Subclass:  ptrInt(2), // Ranged weapon
					Quality:   ptrInt(2), // Uncommon
				},
			},
			level:       25,
			wantDPSMin:  21.0, // Expected DPS range due to randomness
			wantDPSMax:  42.0,
			expectError: false,
		},
		{
			name: "Missing ItemLevel",
			item: Item{
				DbItem: mysql.DbItem{
					Delay:    ptrFloat64(3000),
					MinDmg1:  ptrFloat64(50),
					MaxDmg1:  ptrFloat64(70),
					Subclass: ptrInt(4), // One-handed weapon
					Quality:  ptrInt(3), // Rare
				},
			},
			level:       70,
			wantDPSMin:  0,
			wantDPSMax:  0,
			expectError: true,
		},
		{
			name: "Missing Delay",
			item: Item{
				DbItem: mysql.DbItem{
					ItemLevel: ptrInt(60),
					MinDmg1:   ptrFloat64(50),
					MaxDmg1:   ptrFloat64(70),
					Subclass:  ptrInt(4), // One-handed weapon
					Quality:   ptrInt(3), // Rare
				},
			},
			level:       70,
			wantDPSMin:  0,
			wantDPSMax:  0,
			expectError: true,
		},
		{
			name: "Secondary damage scaling",
			item: Item{
				DbItem: mysql.DbItem{
					ItemLevel: ptrInt(60),
					Delay:     ptrFloat64(3000),
					MinDmg1:   ptrFloat64(50),
					MaxDmg1:   ptrFloat64(70),
					MinDmg2:   ptrFloat64(25),
					MaxDmg2:   ptrFloat64(35),
					Subclass:  ptrInt(4), // One-handed weapon
					Quality:   ptrInt(3), // Rare
				},
			},
			level:       70,
			wantDPSMin:  53.0, // Expected DPS range due to randomness
			wantDPSMax:  107.0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Seed the random number generator for consistent test results
			rand.Seed(1)

			gotDPS, err := tt.item.ScaleDPS(tt.level)
			if (err != nil) != tt.expectError {
				t.Errorf("ScaleDPS() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && (gotDPS < tt.wantDPSMin || gotDPS > tt.wantDPSMax) {
				t.Errorf("ScaleDPS() = %v, want between %v and %v", gotDPS, tt.wantDPSMin, tt.wantDPSMax)
			}
		})
	}
}

func TestGetDpsModifier(t *testing.T) {
	tests := []struct {
		name         string
		item         Item
		wantModifier float64
		expectError  bool
	}{
		{
			name: "Valid one-handed weapon modifier",
			item: Item{
				DbItem: mysql.DbItem{
					Subclass: ptrInt(4), // One-handed weapon
					Quality:  ptrInt(3), // Rare
				},
			},
			wantModifier: 0.64 * 1.38,
			expectError:  false,
		},
		{
			name: "Valid two-handed weapon modifier",
			item: Item{
				DbItem: mysql.DbItem{
					Subclass: ptrInt(17), // Two-handed weapon
					Quality:  ptrInt(4),  // Epic
				},
			},
			wantModifier: 0.80 * 1.5,
			expectError:  false,
		},
		{
			name: "Valid ranged weapon modifier",
			item: Item{
				DbItem: mysql.DbItem{
					Subclass: ptrInt(2), // Ranged weapon
					Quality:  ptrInt(2), // Uncommon
				},
			},
			wantModifier: 0.70 * 1.25,
			expectError:  false,
		},
		{
			name: "Valid wand modifier",
			item: Item{
				DbItem: mysql.DbItem{
					Subclass: ptrInt(19), // Wand
					Quality:  ptrInt(4),  // Epic
				},
			},
			wantModifier: 0.70 * 1.5,
			expectError:  false,
		},
		{
			name: "Invalid subclass",
			item: Item{
				DbItem: mysql.DbItem{
					Subclass: ptrInt(99), // Invalid subclass
					Quality:  ptrInt(3),  // Rare
				},
			},
			wantModifier: 0,
			expectError:  true,
		},
		{
			name: "Missing subclass",
			item: Item{
				DbItem: mysql.DbItem{
					Quality: ptrInt(3), // Rare
				},
			},
			wantModifier: 0,
			expectError:  true,
		},
		{
			name: "Missing quality",
			item: Item{
				DbItem: mysql.DbItem{
					Subclass: ptrInt(4), // One-handed weapon
				},
			},
			wantModifier: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModifier, err := tt.item.GetDpsModifier()
			if (err != nil) != tt.expectError {
				t.Errorf("GetDpsModifier() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && !almostEqual(gotModifier, tt.wantModifier, 0.01) {
				t.Errorf("GetDpsModifier() = %v, want %v", gotModifier, tt.wantModifier)
			}
		})
	}
}

// Helper function to return a pointer to an int
func ptrInt(i int) *int {
	return &i
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
