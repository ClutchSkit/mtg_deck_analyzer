package analysis

import (
	"path/filepath"
	"testing"

	"github.com/ClutchSkit/mtg_deck_statistics/deck"
	"github.com/ClutchSkit/mtg_deck_statistics/scryfall"
)

func TestDeckAnalysisRolesRampCreatureRemoval(t *testing.T) {
	dataSet, err := scryfall.LoadCards(filepath.Join("..", "data", "oracle-cards.jsonl"))
	if err != nil {
		t.Fatalf("Error loading cards: %v", err)
	}
	deckList, _, err := deck.ParseDeckList(filepath.Join("..", "testdata", "ramp_creature_removal_decklist.txt"))
	if err != nil {
		t.Fatalf("Error parsing deck: %v", err)
	}
	resolvedDeck, err := deck.ResolveCards(deckList, dataSet)
	if err != nil {
		t.Fatalf("Error resolving deck: %v", err)
	}
	stats := DeckAnalysis(resolvedDeck)
	if stats.TotalCards != 30 {
		t.Fatalf("Expected total cards to be 30, got %d", stats.TotalCards)
	}
	if stats.Roles.Ramp != 9 {
		t.Fatalf("Expected ramp to be 9, got %d", stats.Roles.Ramp)
	}
	if stats.Roles.CardDraw != 2 {
		t.Fatalf("Expected card draw to be 2, got %d", stats.Roles.CardDraw)
	}
	if stats.Roles.Tutors != 3 {
		t.Fatalf("Expected tutors to be 3, got %d", stats.Roles.Tutors)
	}
	if stats.Roles.CreatureRemoval != 15 {
		t.Fatalf("Expected creature removal to be 15, got %d", stats.Roles.CreatureRemoval)
	}
}
func TestDeckAnalysisRolesArtifactEnchantmentRemoval(t *testing.T) {
	dataSet, err := scryfall.LoadCards(filepath.Join("..", "data", "oracle-cards.jsonl"))
	if err != nil {
		t.Fatalf("Error loading cards: %v", err)
	}
	deckList, _, err := deck.ParseDeckList(filepath.Join("..", "testdata", "artifact_enchantment_removal_decklist.txt"))
	if err != nil {
		t.Fatalf("Error parsing deck: %v", err)
	}
	resolvedDeck, err := deck.ResolveCards(deckList, dataSet)
	if err != nil {
		t.Fatalf("Error resolving deck: %v", err)
	}
	stats := DeckAnalysis(resolvedDeck)
	if stats.TotalCards != 7 {
		t.Fatalf("Expected total cards to be 7, got %d", stats.TotalCards)
	}
	if stats.Roles.ArtifactRemoval != 5 {
		t.Fatalf("Expected artifact removal to be 5, got %d", stats.Roles.ArtifactRemoval)
	}
	if stats.Roles.EnchantmentRemoval != 2 {
		t.Fatalf("Expected enchantment removal to be 2, got %d", stats.Roles.EnchantmentRemoval)
	}
	if stats.Roles.PlaneswalkerRemoval != 3 {
		t.Fatalf("Expected planeswalker removal to be 3, got %d", stats.Roles.PlaneswalkerRemoval)
	}
}

func TestDeckAnalysisRolesProtection(t *testing.T) {
	dataSet, err := scryfall.LoadCards(filepath.Join("..", "data", "oracle-cards.jsonl"))
	if err != nil {
		t.Fatalf("Error loading cards: %v", err)
	}
	deckList, _, err := deck.ParseDeckList(filepath.Join("..", "testdata", "protection_decklist.txt"))
	if err != nil {
		t.Fatalf("Error parsing deck: %v", err)
	}
	resolvedDeck, err := deck.ResolveCards(deckList, dataSet)
	if err != nil {
		t.Fatalf("Error resolving deck: %v", err)
	}
	stats := DeckAnalysis(resolvedDeck)
	if stats.TotalCards != 13 {
		t.Fatalf("Expected total cards to be 13, got %d", stats.TotalCards)
	}
	if stats.Roles.Protection != 7 {
		t.Fatalf("Expected protection to be 7, got %d", stats.Roles.Protection)
	}
	if stats.Roles.Recursion != 2 {
		t.Fatalf("Expected recursion to be 2, got %d", stats.Roles.Recursion)
	}
	if stats.Roles.TokenGeneration != 3 {
		t.Fatalf("Expected token generation to be 3, got %d", stats.Roles.TokenGeneration)
	}
	if stats.Roles.Other != 1 {
		t.Fatalf("Expected other to be 1, got %d", stats.Roles.Other)
	}
}
