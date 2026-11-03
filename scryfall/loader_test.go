package scryfall

import (
	"testing"
)

func TestLoadCards(t *testing.T) {
	cards, err := LoadCards("../testdata/sample_cards.jsonl")
	if err != nil {
		t.Fatalf("Error loading cards: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("Expected 2 entries, got: %v", len(cards))
	}
	card, exists := cards["Nissa, Worldsoul Speaker"]
	if exists == false {
		t.Fatalf("Nissa, Worldsoul Speaker not found in map")
	}
	if card.Name != "Nissa, Worldsoul Speaker" {
		t.Errorf("Expected card name to be Nissa, Worldsoul Speaker, got: %s", card.Name)
	}
	if card.CMC != 4.0 {
		t.Errorf("Expected cmc to be 4.0, got %f", card.CMC)
	}
	if card.ManaCost != "{3}{G}" {
		t.Errorf("Expected mana cost to be {3}{G}, got: %s", card.ManaCost)
	}
}
