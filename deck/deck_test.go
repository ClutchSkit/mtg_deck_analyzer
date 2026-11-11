package deck

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ClutchSkit/mtg_deck_analyzer/scryfall"
)

func TestParseDeckLine(t *testing.T) {
	tests := []struct {
		line             string
		quantityExpected int
		nameExpected     string
		errorExpected    bool
	}{
		{line: "2 Sol Ring", quantityExpected: 2, nameExpected: "Sol Ring", errorExpected: false},
		{line: "abc Sol Ring", errorExpected: true},
		{line: "0 Sol Ring", errorExpected: true},
		{line: "-2 Sol Ring", errorExpected: true},
		{line: "0.5 Sol Ring", errorExpected: true},
		{line: "Sol Ring", errorExpected: true},
		{line: "2 ", errorExpected: true},
		{line: " Sol Ring", errorExpected: true},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			if tt.errorExpected == true {
				_, err := ParseDeckLine(tt.line)
				if err == nil {
					t.Fatalf("Expected error, instead line was parsed")
				}
			} else {
				got, err := ParseDeckLine(tt.line)
				if err != nil {
					t.Fatalf("Error thrown when it was not expected")
				}
				if got.Name != tt.nameExpected {
					t.Errorf("Name Expected was %s, instead got %s", tt.nameExpected, got.Name)
				}
				if got.Quantity != tt.quantityExpected {
					t.Errorf("Quantity Expected was %d, instead got %d", tt.quantityExpected, got.Quantity)
				}
			}
		})
	}
}

func TestParseDeckList(t *testing.T) {
	decklist, _, err := ParseDeckList("../testdata/sample.txt")

	if err != nil {
		t.Fatalf("Error parsing deck: %v", err)
	}
	if len(decklist) != 5 {
		t.Fatalf("Expected 5 cards, got %d", len(decklist))
	}
	if decklist[0].Name != "Sol Ring" {
		t.Errorf("Expected name of first card Sol Ring, got %s", decklist[0].Name)
	}
	if decklist[0].Quantity != 1 {
		t.Errorf("Expected quantity of first card 1, got %d", decklist[0].Quantity)
	}
}

func TestParseDeckListText(t *testing.T) {
	decklist, parseErrors, err := ParseDeckListText("1 Sol Ring\n1 Arcane Signet\n1 Cultivate\n1 Black Lotus\n1 Mana Crypt")

	if err != nil {
		t.Fatalf("Error parsing deck: %v", err)
	}
	if len(parseErrors) != 0 {
		t.Fatalf("Expected 0 lines witth parse errors, got: %d", len(parseErrors))
	}
	if len(decklist) != 5 {
		t.Fatalf("Expected 5 cards, got %d", len(decklist))
	}
	if decklist[0].Name != "Sol Ring" {
		t.Errorf("Expected name of first card Sol Ring, got %s", decklist[0].Name)
	}
	if decklist[0].Quantity != 1 {
		t.Errorf("Expected quantity of first card 1, got %d", decklist[0].Quantity)
	}
	if decklist[4].Name != "Mana Crypt" {
		t.Errorf("Expected name of last card Mana Crypt, got %s", decklist[4].Name)
	}
	if decklist[4].Quantity != 1 {
		t.Errorf("Expected quantity of last card 1, got %d", decklist[4].Quantity)
	}
}

func TestParseDeckListSkipsMalformedLines(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test_deck.txt")
	err := os.WriteFile(path, []byte("2 Sol Ring\nabc Sol Ring\n1 Lightning Bolt\n"), 0600)
	if err != nil {
		t.Fatalf("Error writing test file: %v", err)
	}
	decklist, parseErrors, err := ParseDeckList(path)
	if err != nil {
		t.Fatalf("Error parsing deck: %v", err)
	}
	if len(decklist) != 2 {
		t.Fatalf("Expected 2 cards, got %d", len(decklist))
	}
	if len(parseErrors) != 1 {
		t.Fatalf("Expected 1 parse error, got %d", len(parseErrors))
	}
	if !strings.Contains(parseErrors[0], "Line 2") {
		t.Errorf("Expected parse erroron Line 2, got %s", parseErrors[0])
	}
	if decklist[0].Name != "Sol Ring" || decklist[0].Quantity != 2 {
		t.Errorf("Expected first card to be 2 Sol Ring, got %d %s", decklist[0].Quantity, decklist[0].Name)
	}
	if decklist[1].Name != "Lightning Bolt" || decklist[1].Quantity != 1 {
		t.Errorf("Expected second card to be 1 Lightning Bolt, got %d %s", decklist[1].Quantity, decklist[1].Name)
	}
}

func TestResolveCards(t *testing.T) {
	cardDatabase := map[string]scryfall.Card{
		"Sol Ring": {
			Name:     "Sol Ring",
			ManaCost: "{1}",
			CMC:      1,
		},
		"Arcane Signet": {
			Name:     "Arcane Signet",
			ManaCost: "{2}",
			CMC:      2,
		},
	}
	deckList := []CardEntry{
		{
			Quantity: 1,
			Name:     "Sol Ring",
		},
		{
			Quantity: 2,
			Name:     "Arcane Signet",
		},
	}
	foundCards, err := ResolveCards(deckList, cardDatabase)
	if err != nil {
		t.Fatalf("Missing cards when it was not expected: %v", err)
	}
	if len(foundCards) != 2 {
		t.Fatalf("Expected to get 2 cards, got: %d", len(foundCards))
	}
	if foundCards[0].Card.Name != "Sol Ring" {
		t.Fatalf("Expected name of first card to be Sol Ring, got: %s", foundCards[0].Card.Name)
	}
	if foundCards[0].Card.ManaCost != "{1}" {
		t.Fatalf("Expected mana cost of first card to be {1}, got: %s", foundCards[0].Card.ManaCost)
	}
	if foundCards[0].Card.CMC != 1.0 {
		t.Fatalf("Expected cmc of first card  to be 1.0, got: %f", foundCards[0].Card.CMC)
	}
	if foundCards[0].Quantity != 1 {
		t.Fatalf("Expected quantity of first card to be 1, got: %d", foundCards[0].Quantity)
	}
	if foundCards[1].Quantity != 2 {
		t.Fatalf("Expected quantity of second card to be 2, got: %d", foundCards[1].Quantity)
	}
}

func TestResolveCardsMissingCards(t *testing.T) {
	var missingCardsErr *MissingCardsError
	cardDatabase := map[string]scryfall.Card{
		"Sol Ring": {
			Name:     "Sol Ring",
			ManaCost: "{1}",
			CMC:      1,
		},
		"Arcane Signet": {
			Name:     "Arcane Signet",
			ManaCost: "{2}",
			CMC:      2,
		},
	}
	deckList := []CardEntry{
		{
			Quantity: 1,
			Name:     "Sol Ring",
		},
		{
			Quantity: 2,
			Name:     "Mana Crypt",
		},
	}
	foundCards, err := ResolveCards(deckList, cardDatabase)
	if err == nil {
		t.Fatalf("No missing cards when it was expected")
	}
	if errors.As(err, &missingCardsErr) == false {
		t.Fatalf("Expected error to be of type MissingCardsError")
	}
	if len(missingCardsErr.Names) != 1 {
		t.Fatalf("Expected 1 missing card, got: %d", len(missingCardsErr.Names))
	}
	if missingCardsErr.Names[0] != "Mana Crypt" {
		t.Fatalf("Expected missing card to be Mana Crypt, got: %s", missingCardsErr.Names[0])
	}
	if len(foundCards) != 1 {
		t.Fatalf("Expected to get 1 card, got: %d", len(foundCards))
	}
	if foundCards[0].Card.Name != "Sol Ring" {
		t.Fatalf("Expected name of first card to be Sol Ring, got: %s", foundCards[0].Card.Name)
	}
	if foundCards[0].Card.ManaCost != "{1}" {
		t.Fatalf("Expected mana cost of first card to be {1}, got: %s", foundCards[0].Card.ManaCost)
	}
	if foundCards[0].Card.CMC != 1.0 {
		t.Fatalf("Expected cmc of first card  to be 1.0, got: %f", foundCards[0].Card.CMC)
	}
	if foundCards[0].Quantity != 1 {
		t.Fatalf("Expected quantity of first card to be 1, got: %d", foundCards[0].Quantity)
	}
}
