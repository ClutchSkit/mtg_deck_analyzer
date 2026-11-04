package analysis

import (
	"testing"

	"github.com/ClutchSkit/mtg_deck_analyzer/deck"
	"github.com/ClutchSkit/mtg_deck_analyzer/scryfall"
)

func makeFoundCard(quantity int, name string, typeLine string, colors []string, production []string, cmc float64) deck.FoundCard {
	return deck.FoundCard{
		Quantity: quantity,
		Card: scryfall.Card{
			Name:         name,
			TypeLine:     typeLine,
			Colors:       colors,
			ProducedMana: production,
			CMC:          cmc,
		},
	}
}

func TestDeckAnalysis(t *testing.T) {
	deckList := []deck.FoundCard{
		makeFoundCard(9, "Special Island", "Artifact Land — Island", nil, []string{"U"}, 0),
		makeFoundCard(1, "Dimir Guildgate", "Land", nil, []string{"U", "B"}, 0),
		makeFoundCard(1, "Wastes", "Basic Land", nil, []string{"C"}, 0),
		makeFoundCard(1, "Sol Ring", "Legendary Artifact", nil, []string{"C", "C"}, 2),
		makeFoundCard(1, "Llanowar Elves", "Creature — Elf Druid", []string{"G"}, []string{"G"}, 1),
		makeFoundCard(1, "Phyrexian Arena", "Enchantment", []string{"B"}, nil, 3),
		makeFoundCard(1, "Assassin's Trophy", "Instant", []string{"B", "G"}, nil, 2),
		makeFoundCard(1, "Cultivate", "Sorcery", []string{"G"}, nil, 3),
		makeFoundCard(1, "Jace, the Mind Sculptor", "Legendary Planeswalker — Jace", []string{"U"}, nil, 4),
		makeFoundCard(1, "Invasion of Zendikar", "Battle — Siege", []string{"G"}, nil, 4),
	}

	deckStats := DeckAnalysis(deckList)

	tests := []struct {
		name     string
		got      int
		expected int
	}{
		// General statistics
		{"total cards", deckStats.TotalCards, 18},
		{"total nonlands", deckStats.TotalNonlands, 7},

		// Card types
		{"lands", deckStats.CardTypes.Lands, 11},
		{"artifacts", deckStats.CardTypes.Artifacts, 10},
		{"creatures", deckStats.CardTypes.Creatures, 1},
		{"enchantments", deckStats.CardTypes.Enchantments, 1},
		{"instants", deckStats.CardTypes.Instants, 1},
		{"sorceries", deckStats.CardTypes.Sorceries, 1},
		{"planeswalkers", deckStats.CardTypes.Planeswalkers, 1},
		{"battles", deckStats.CardTypes.Battles, 1},

		// Color requirements
		{"black requirements", deckStats.Mana.ColorRequirements.Black, 2},
		{"blue requirements", deckStats.Mana.ColorRequirements.Blue, 1},
		{"green requirements", deckStats.Mana.ColorRequirements.Green, 4},
		{"red requirements", deckStats.Mana.ColorRequirements.Red, 0},
		{"white requirements", deckStats.Mana.ColorRequirements.White, 0},

		// Land production
		{"blue production", deckStats.Mana.LandProduction.Blue, 10},
		{"black production", deckStats.Mana.LandProduction.Black, 1},
		{"colorless production", deckStats.Mana.LandProduction.Colorless, 1},
		{"red production", deckStats.Mana.LandProduction.Red, 0},
		{"green production", deckStats.Mana.LandProduction.Green, 0},
		{"white production", deckStats.Mana.LandProduction.White, 0},

		//CMC
		{"cmc of 0", deckStats.Mana.ManaCurve.ManaCosts[0], 11},
		{"cmc of 1", deckStats.Mana.ManaCurve.ManaCosts[1], 1},
		{"cmc of 2", deckStats.Mana.ManaCurve.ManaCosts[2], 2},
		{"cmc of 3", deckStats.Mana.ManaCurve.ManaCosts[3], 2},
		{"cmc of 4", deckStats.Mana.ManaCurve.ManaCosts[4], 2},

		//Roles
		{"ramp", deckStats.Roles.Ramp, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("expected %s to be %d, got %d",
					tt.name, tt.expected, tt.got)
			}
		})
	}
}
