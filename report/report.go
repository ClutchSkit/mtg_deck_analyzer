package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ClutchSkit/mtg_deck_analyzer/analysis"
)

func GenerateReport(commanderName string, deckStats analysis.DeckStats, probabilities analysis.Probabilities) {
	fmt.Println("\nMTG DECK ANALYZER V1.0")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("")
	fmt.Println("Commander:", commanderName)
	fmt.Println("Total Cards in library:", deckStats.TotalCards)
	fmt.Println("")
	fmt.Println("PROBABILITIES")
	fmt.Println("Based on 7 card opening hand, plus 1 card drawn on turn 1(8 cards total).")
	fmt.Println("Assumes no mulligans.")
	fmt.Println("Percentages represent chance of drawing at least")
	fmt.Println("one card from the specified category, unless otherwise specified.")
	fmt.Println("")
	fmt.Println("CARD TYPES")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("")
	fmt.Printf("%-20s %5d\n", "Lands", deckStats.CardTypes.Lands)
	fmt.Printf("  %-20s %8.1f%%\n", "At least one", probabilities.Lands.AtLeastOne*100)
	fmt.Printf("  %-20s %8.1f%%\n", "At least two", probabilities.Lands.AtLeastTwo*100)
	fmt.Printf("  %-20s %8.1f%%\n", "At least three", probabilities.Lands.AtLeastThree*100)
	fmt.Printf("%-20s %5d\n", "Creatures", deckStats.CardTypes.Creatures)
	fmt.Printf("%-20s %5d\n", "Artifacts", deckStats.CardTypes.Artifacts)
	fmt.Printf("%-20s %5d\n", "Enchantments", deckStats.CardTypes.Enchantments)
	fmt.Printf("%-20s %5d\n", "Instants", deckStats.CardTypes.Instants)
	fmt.Printf("%-20s %5d\n", "Sorceries", deckStats.CardTypes.Sorceries)
	fmt.Printf("%-20s %5d\n", "Planeswalkers", deckStats.CardTypes.Planeswalkers)
	fmt.Printf("%-20s %5d\n\n", "Battles", deckStats.CardTypes.Battles)
	fmt.Println("MANA CURVE")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("")
	manaCosts := deckStats.Mana.ManaCurve.ManaCosts

	var manaValues []float64

	for manaValue := range manaCosts {
		manaValues = append(manaValues, manaValue)
	}
	sort.Float64s(manaValues)
	for _, manaValue := range manaValues {
		fmt.Printf("%-20s %5d\n", fmt.Sprintf("%g mana", manaValue), manaCosts[manaValue])
	}
	fmt.Println("")
	fmt.Println("MANA COLOR REQUIREMENTS")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("")
	fmt.Printf("%-20s %5d\n", "White", deckStats.Mana.ColorRequirements.White)
	fmt.Printf("%-20s %5d\n", "Blue", deckStats.Mana.ColorRequirements.Blue)
	fmt.Printf("%-20s %5d\n", "Black", deckStats.Mana.ColorRequirements.Black)
	fmt.Printf("%-20s %5d\n", "Red", deckStats.Mana.ColorRequirements.Red)
	fmt.Printf("%-20s %5d\n", "Green", deckStats.Mana.ColorRequirements.Green)
	fmt.Println("")
	fmt.Println("LAND MANA PRODUCTION")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("")
	fmt.Printf("%-20s %5d %8.1f%%\n", "White", deckStats.Mana.LandProduction.White, probabilities.LandColors.White*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Blue", deckStats.Mana.LandProduction.Blue, probabilities.LandColors.Blue*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Black", deckStats.Mana.LandProduction.Black, probabilities.LandColors.Black*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Red", deckStats.Mana.LandProduction.Red, probabilities.LandColors.Red*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Green", deckStats.Mana.LandProduction.Green, probabilities.LandColors.Green*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Colorless", deckStats.Mana.LandProduction.Colorless, probabilities.LandColors.Colorless*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Fetch", deckStats.Mana.LandProduction.Fetch, probabilities.LandColors.Fetch*100)
	fmt.Println("")
	fmt.Println("CARD ROLES")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("")
	fmt.Printf("%-20s %5d %8.1f%%\n", "Ramp", deckStats.Roles.Ramp, probabilities.Roles.Ramp*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Card Draw", deckStats.Roles.CardDraw, probabilities.Roles.CardDraw*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Tutors", deckStats.Roles.Tutors, probabilities.Roles.Tutors*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Creature Removal", deckStats.Roles.CreatureRemoval, probabilities.Roles.CreatureRemoval*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Artifact Removal", deckStats.Roles.ArtifactRemoval, probabilities.Roles.ArtifactRemoval*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Enchantment Removal", deckStats.Roles.EnchantmentRemoval, probabilities.Roles.EnchantmentRemoval*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Planeswalker Removal", deckStats.Roles.PlaneswalkerRemoval, probabilities.Roles.PlaneswalkerRemoval*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Protection", deckStats.Roles.Protection, probabilities.Roles.Protection*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Recursion", deckStats.Roles.Recursion, probabilities.Roles.Recursion*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Token Generation", deckStats.Roles.TokenGeneration, probabilities.Roles.TokenGeneration*100)
	fmt.Printf("%-20s %5d %8.1f%%\n", "Other", deckStats.Roles.Other, probabilities.Roles.Other*100)
	fmt.Println("")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("")
	fmt.Println("CONCLUSION")
	fmt.Println(strings.Repeat("-", 50))
	recomendations := analysis.GenerateRecomendations(deckStats)

	for _, recomendation := range recomendations {
		fmt.Println(recomendation)
	}
}
