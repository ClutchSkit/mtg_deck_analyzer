package analysis

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ClutchSkit/mtg_deck_analyzer/deck"
)

const OpeningHandSize = 8

//does not consider card faces for role analysis, cards with multiple faces will not be counted for roles, since oracle_text will exist within the card_faces array

type DeckStats struct {
	TotalCards    int
	TotalNonlands int

	CardTypes CardTypeStats
	Mana      ManaStats
	Roles     RoleStats
}

type CardTypeStats struct {
	Lands         int
	Creatures     int
	Artifacts     int
	Enchantments  int
	Instants      int
	Sorceries     int
	Planeswalkers int
	Battles       int
}

type ManaStats struct {
	ColorRequirements ColorRequirementStats
	LandProduction    LandProductionStats
	ManaCurve         ManaCurveStats
}

type ColorRequirementStats struct {
	White int
	Blue  int
	Black int
	Red   int
	Green int
}

type LandProductionStats struct {
	White     int
	Blue      int
	Black     int
	Red       int
	Green     int
	Colorless int
	Fetch     int
}

type ManaCurveStats struct {
	ManaCosts map[float64]int
}

type RoleStats struct {
	Ramp                int
	CardDraw            int
	Tutors              int
	CreatureRemoval     int
	ArtifactRemoval     int
	EnchantmentRemoval  int
	PlaneswalkerRemoval int
	Protection          int
	Recursion           int
	TokenGeneration     int
	Other               int
}

type Probabilities struct {
	Lands      LandProbabilities
	LandColors LandColorProbabilities
	Roles      RoleProbabilities
}

type LandProbabilities struct {
	AtLeastOne   float64
	AtLeastTwo   float64
	AtLeastThree float64
}
type LandColorProbabilities struct {
	White     float64
	Blue      float64
	Black     float64
	Red       float64
	Green     float64
	Colorless float64
	Fetch     float64
}
type RoleProbabilities struct {
	Ramp                float64
	CardDraw            float64
	Tutors              float64
	CreatureRemoval     float64
	ArtifactRemoval     float64
	EnchantmentRemoval  float64
	PlaneswalkerRemoval float64
	Protection          float64
	Recursion           float64
	TokenGeneration     float64
	Other               float64
}

// fetch land detection patterns
var fetchLandPattern = regexp.MustCompile(`(?i)search your library for .*(lands?|basic lands?|islands?|mountains?|swamps?|forests?|plains?).*put .*battlefield`)

// ramp detection patterns
var rampLandPattern = regexp.MustCompile(`(?i)search your library for .*lands?[ ,].*put .*battlefield`)
var rampTreasurePattern = regexp.MustCompile(`(?i)create .*treasure tokens?`)
var rampReductionPattern = regexp.MustCompile(`(?i)spells? you cast costs? .*less to cast`)
var rampAdditionalLandPattern = regexp.MustCompile(`(?i)play .* additional lands?`)

// card draw detection patterns
var cardDrawPattern = regexp.MustCompile(`(?i)draw .*cards?`)

// tutor detection patterns
var tutorPattern = regexp.MustCompile(`(?i)search your library for .*cards?[ ,] .*(put|add)`)

// removal detection patterns
// creature removal patterns
var creatureDestroyRemovalPattern = regexp.MustCompile(`(?i)destroys? (target|all|each|up to) .*(creatures?|nonland permanents?).*`)
var creatureExileRemovalPattern = regexp.MustCompile(`(?i)exiles? (target|all|each|up to) .*(creatures?|nonland permanents?).*`)
var creatureDamageRemovalPattern = regexp.MustCompile(`(?i)deals? .*damage .*to (target|all|each|any|up to) .*(creatures?|targets?).*`)
var creatureGetsMinusXMinusXRemovalPattern = regexp.MustCompile(`(?i)gets? (-.*|0|\+.*)/-[^0].*`)
var creatureFightRemovalPattern = regexp.MustCompile(`(?i)fights? .*creatures?`)
var creatureSacrificeRemovalPattern = regexp.MustCompile(`(?i)(targets?|each|all|up to) .*sacrifices? .*creatures?`)
var creatureBounceRemovalPattern = regexp.MustCompile(`(?i)(returns?|puts?) .*(target|all|each|any|up to) .*(creatures?|cards?|nonland permanents?) .*(hand|library)`)

// artifact removal patterns
var artifactDestroyRemovalPattern = regexp.MustCompile(`(?i)destroys? .*(target|all|each|up to) .*(artifacts?|nonland permanents?)`)
var artifactExileRemovalPattern = regexp.MustCompile(`(?i)exiles? .*(target|all|each|up to) .*(artifacts?|nonland permanents?)`)
var artifactBounceRemovalPattern = regexp.MustCompile(`(?i)(returns?|puts?) .*(target|all|each|any|up to) .*(artifacts?|cards?|nonland permanents?) .*(hand|library)`)

// enchantment removal patterns
var enchantmentDestroyRemovalPattern = regexp.MustCompile(`(?i)destroys? .*(target|all|each|up to) .*(enchantments?|nonland permanents?)`)
var enchantmentExileRemovalPattern = regexp.MustCompile(`(?i)exiles? .*(target|all|each|up to) .*(enchantments?|nonland permanents?)`)
var enchantmentBounceRemovalPattern = regexp.MustCompile(`(?i)(returns?|puts?) .*(target|all|each|any|up to) .*(enchantments?|cards?|nonland permanents?) .*(hand|library)`)

// planeswalker removal patterns
var planeswalkerDestroyRemovalPattern = regexp.MustCompile(`(?i)destroys? .*(target|all|each|up to) .*(planeswalkers?|nonland permanents?)`)
var planeswalkerExileRemovalPattern = regexp.MustCompile(`(?i)exiles? .*(target|all|each|up to) .*(planeswalkers?|nonland permanents?)`)
var planeswalkerDamageRemovalPattern = regexp.MustCompile(`(?i)deals? .*damage .*to (target|all|each|any|up to) .*(planeswalkers?|targets?).*`)
var planeswalkerBounceRemovalPattern = regexp.MustCompile(`(?i)(returns?|puts?) .*(target|all|each|any|up to) .*(planeswalkers?|cards?|nonland permanents?) .*(hand|library)`)

//hexproof and indestructible circumvention patterns EVENTUALY

// protection patterns
var counterspellProtectionPattern = regexp.MustCompile(`(?i)counters? .*spells?`)
var permanentProtectionPattern = regexp.MustCompile(`(?i)(gains?|have|has|are) .*(indestructible|hexproof|protection from .*|shroud)`)

// recursion patterns
var recursionPattern = regexp.MustCompile(`(?i)(returns?|puts?) .*(target|all|each|any|a|up to) .*(creatures?|artifacts?|enchantments?|planeswalkers?|instants?|sorcery|sorceries|cards?|nonland permanents?) .*graveyards? .*(hand|library|battlefield)`)

// token generation
var tokenGenerationPattern = regexp.MustCompile(`(?i)creates? .*tokens?`)

func DeckAnalysis(deck []deck.FoundCard) DeckStats {
	var stats DeckStats
	var remainingText string
	stats.Mana.ManaCurve.ManaCosts = make(map[float64]int)
	for _, card := range deck {
		typeLine := card.Card.TypeLine
		roleFound := false
		if strings.Contains(typeLine, "Land") == true {
			stats.CardTypes.Lands += card.Quantity
			if fetchLandPattern.MatchString(card.Card.OracleText) == true {
				stats.Mana.LandProduction.Fetch += card.Quantity
			}
			roleFound = true
			for _, production := range card.Card.ProducedMana {
				switch production {
				case "W":
					stats.Mana.LandProduction.White += card.Quantity
				case "U":
					stats.Mana.LandProduction.Blue += card.Quantity
				case "B":
					stats.Mana.LandProduction.Black += card.Quantity
				case "R":
					stats.Mana.LandProduction.Red += card.Quantity
				case "G":
					stats.Mana.LandProduction.Green += card.Quantity
				case "C":
					stats.Mana.LandProduction.Colorless += card.Quantity
				}
			}
		} else {
			stats.TotalNonlands += card.Quantity
			if len(card.Card.ProducedMana) > 0 ||
				rampLandPattern.MatchString(card.Card.OracleText) == true ||
				rampTreasurePattern.MatchString(card.Card.OracleText) == true ||
				rampReductionPattern.MatchString(card.Card.OracleText) == true ||
				rampAdditionalLandPattern.MatchString(card.Card.OracleText) == true {
				stats.Roles.Ramp += card.Quantity
				//card ramp
				roleFound = true
			}
			if cardDrawPattern.MatchString(card.Card.OracleText) == true {
				stats.Roles.CardDraw += card.Quantity
				//card draw
				roleFound = true
			}
			if tutorPattern.MatchString(card.Card.OracleText) == true &&
				rampLandPattern.MatchString(card.Card.OracleText) == false {
				stats.Roles.Tutors += card.Quantity
				//tutors
				roleFound = true
			}

			indexes := creatureBounceRemovalPattern.FindStringIndex(card.Card.OracleText)

			if indexes != nil {
				remainingText = card.Card.OracleText[:indexes[0]] + card.Card.OracleText[indexes[1]:]
			} else {
				remainingText = ""
			}

			if creatureDestroyRemovalPattern.MatchString(card.Card.OracleText) == true ||
				creatureExileRemovalPattern.MatchString(card.Card.OracleText) == true ||
				creatureDamageRemovalPattern.MatchString(card.Card.OracleText) == true ||
				creatureGetsMinusXMinusXRemovalPattern.MatchString(card.Card.OracleText) == true ||
				creatureFightRemovalPattern.MatchString(card.Card.OracleText) == true ||
				creatureSacrificeRemovalPattern.MatchString(card.Card.OracleText) == true ||
				(indexes != nil && strings.Contains(strings.ToLower(remainingText), "graveyard") == false) {
				stats.Roles.CreatureRemoval += card.Quantity
				//creature removal
				roleFound = true
			}

			indexes = artifactBounceRemovalPattern.FindStringIndex(card.Card.OracleText)

			if indexes != nil {
				remainingText = card.Card.OracleText[:indexes[0]] + card.Card.OracleText[indexes[1]:]
			} else {
				remainingText = ""
			}

			if artifactDestroyRemovalPattern.MatchString(card.Card.OracleText) == true ||
				artifactExileRemovalPattern.MatchString(card.Card.OracleText) == true ||
				(indexes != nil && strings.Contains(strings.ToLower(remainingText), "graveyard") == false) {
				stats.Roles.ArtifactRemoval += card.Quantity
				//artifact removal
				roleFound = true
			}

			indexes = enchantmentBounceRemovalPattern.FindStringIndex(card.Card.OracleText)

			if indexes != nil {
				remainingText = card.Card.OracleText[:indexes[0]] + card.Card.OracleText[indexes[1]:]
			} else {
				remainingText = ""
			}

			if enchantmentDestroyRemovalPattern.MatchString(card.Card.OracleText) == true ||
				enchantmentExileRemovalPattern.MatchString(card.Card.OracleText) == true ||
				(indexes != nil && strings.Contains(strings.ToLower(remainingText), "graveyard") == false) {
				stats.Roles.EnchantmentRemoval += card.Quantity
				//enchantment removal
				roleFound = true
			}

			indexes = planeswalkerBounceRemovalPattern.FindStringIndex(card.Card.OracleText)

			if indexes != nil {
				remainingText = card.Card.OracleText[:indexes[0]] + card.Card.OracleText[indexes[1]:]
			} else {
				remainingText = ""
			}

			if planeswalkerDestroyRemovalPattern.MatchString(card.Card.OracleText) == true ||
				planeswalkerExileRemovalPattern.MatchString(card.Card.OracleText) == true ||
				planeswalkerDamageRemovalPattern.MatchString(card.Card.OracleText) == true ||
				(indexes != nil && strings.Contains(strings.ToLower(remainingText), "graveyard") == false) {
				stats.Roles.PlaneswalkerRemoval += card.Quantity
				//planeswalker removal
				roleFound = true
			}
			if counterspellProtectionPattern.MatchString(card.Card.OracleText) == true ||
				permanentProtectionPattern.MatchString(card.Card.OracleText) == true {
				stats.Roles.Protection += card.Quantity
				//protection
				roleFound = true
			}
			if recursionPattern.MatchString(card.Card.OracleText) == true {
				stats.Roles.Recursion += card.Quantity
				//recursion
				roleFound = true
			}
			if tokenGenerationPattern.MatchString(card.Card.OracleText) == true {
				stats.Roles.TokenGeneration += card.Quantity
				//token generation
				roleFound = true
			}
			if roleFound == false {
				stats.Roles.Other += card.Quantity
			}

		}
		if strings.Contains(typeLine, "Creature") == true {
			stats.CardTypes.Creatures += card.Quantity
		}

		if strings.Contains(typeLine, "Artifact") == true {
			stats.CardTypes.Artifacts += card.Quantity
		}

		if strings.Contains(typeLine, "Enchantment") == true {
			stats.CardTypes.Enchantments += card.Quantity
		}

		if strings.Contains(typeLine, "Instant") == true {
			stats.CardTypes.Instants += card.Quantity
		}

		if strings.Contains(typeLine, "Sorcery") == true {
			stats.CardTypes.Sorceries += card.Quantity
		}

		if strings.Contains(typeLine, "Planeswalker") == true {
			stats.CardTypes.Planeswalkers += card.Quantity
		}

		if strings.Contains(typeLine, "Battle") == true {
			stats.CardTypes.Battles += card.Quantity
		}
		stats.TotalCards += card.Quantity

		stats.Mana.ManaCurve.ManaCosts[card.Card.CMC] += card.Quantity

		for _, color := range card.Card.Colors {
			switch color {
			case "W":
				stats.Mana.ColorRequirements.White += card.Quantity
			case "U":
				stats.Mana.ColorRequirements.Blue += card.Quantity
			case "B":
				stats.Mana.ColorRequirements.Black += card.Quantity
			case "R":
				stats.Mana.ColorRequirements.Red += card.Quantity
			case "G":
				stats.Mana.ColorRequirements.Green += card.Quantity
			}
		}
	}
	return stats
}

func HypergeometricProbability(LibrarySize int, CardsInCategory int, NumberOfHits int) (float64, error) {
	if NumberOfHits > CardsInCategory {
		return 0, nil
	}
	if OpeningHandSize > LibrarySize {
		return 0, fmt.Errorf("OpeningHandSize cannot be greater than LibrarySize")
	}
	probability := 0.0
	for k := 0; k < NumberOfHits; k++ {
		probability += combinations(CardsInCategory, k) * combinations(LibrarySize-CardsInCategory, OpeningHandSize-k) / combinations(LibrarySize, OpeningHandSize)
	}
	return 1 - probability, nil
}

func combinations(n int, k int) float64 {
	if k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}
	if k > n-k {
		k = n - k
	}

	numerator := 1.0
	denominator := 1.0

	for i := 1; i <= k; i++ {
		denominator *= float64(i)
	}
	for i := n - k + 1; i <= n; i++ {
		numerator *= float64(i)
	}

	return numerator / denominator
}

func CalculateProbabilities(stats DeckStats) (Probabilities, error) {
	if stats.TotalCards < OpeningHandSize {
		return Probabilities{}, fmt.Errorf("TotalCards cannot be less than OpeningHandSize")
	}
	var probabilities Probabilities
	probabilities.Lands.AtLeastOne, _ = HypergeometricProbability(stats.TotalCards, stats.CardTypes.Lands, 1)
	probabilities.Lands.AtLeastTwo, _ = HypergeometricProbability(stats.TotalCards, stats.CardTypes.Lands, 2)
	probabilities.Lands.AtLeastThree, _ = HypergeometricProbability(stats.TotalCards, stats.CardTypes.Lands, 3)

	probabilities.LandColors.White, _ = HypergeometricProbability(stats.TotalCards, stats.Mana.LandProduction.White, 1)
	probabilities.LandColors.Blue, _ = HypergeometricProbability(stats.TotalCards, stats.Mana.LandProduction.Blue, 1)
	probabilities.LandColors.Black, _ = HypergeometricProbability(stats.TotalCards, stats.Mana.LandProduction.Black, 1)
	probabilities.LandColors.Red, _ = HypergeometricProbability(stats.TotalCards, stats.Mana.LandProduction.Red, 1)
	probabilities.LandColors.Green, _ = HypergeometricProbability(stats.TotalCards, stats.Mana.LandProduction.Green, 1)
	probabilities.LandColors.Colorless, _ = HypergeometricProbability(stats.TotalCards, stats.Mana.LandProduction.Colorless, 1)
	probabilities.LandColors.Fetch, _ = HypergeometricProbability(stats.TotalCards, stats.Mana.LandProduction.Fetch, 1)

	probabilities.Roles.Ramp, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.Ramp, 1)
	probabilities.Roles.CardDraw, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.CardDraw, 1)
	probabilities.Roles.Tutors, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.Tutors, 1)
	probabilities.Roles.CreatureRemoval, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.CreatureRemoval, 1)
	probabilities.Roles.ArtifactRemoval, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.ArtifactRemoval, 1)
	probabilities.Roles.EnchantmentRemoval, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.EnchantmentRemoval, 1)
	probabilities.Roles.PlaneswalkerRemoval, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.PlaneswalkerRemoval, 1)
	probabilities.Roles.Protection, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.Protection, 1)
	probabilities.Roles.Recursion, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.Recursion, 1)
	probabilities.Roles.TokenGeneration, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.TokenGeneration, 1)
	probabilities.Roles.Other, _ = HypergeometricProbability(stats.TotalCards, stats.Roles.Other, 1)

	return probabilities, nil
}

func GenerateRecomendations(stats DeckStats) []string {
	var recomendations []string
	if stats.Roles.CreatureRemoval == 0 {
		recomendations = append(recomendations, "No creature removal detected")
	}
	if stats.Roles.ArtifactRemoval == 0 {
		recomendations = append(recomendations, "No artifact removal detected")
	}
	if stats.Roles.EnchantmentRemoval == 0 {
		recomendations = append(recomendations, "No enchantment removal detected")
	}
	if stats.Roles.PlaneswalkerRemoval == 0 {
		recomendations = append(recomendations, "No planeswalker removal detected")
	}
	if stats.CardTypes.Lands < 33 {
		recomendations = append(recomendations, fmt.Sprintf("Only %d lands detected, these may be insufficient", stats.CardTypes.Lands))
	}
	if stats.CardTypes.Lands > 45 {
		recomendations = append(recomendations, fmt.Sprintf("%d lands detected, these may be excessive", stats.CardTypes.Lands))
	}

	if stats.TotalNonlands != 0 && stats.CardTypes.Lands != 0 {
		requirementNotice := false
		//White requirements
		requirementRatio := float64(stats.Mana.ColorRequirements.White) / float64(stats.TotalNonlands)
		landRatio := float64(stats.Mana.LandProduction.White) / float64(stats.CardTypes.Lands)

		if requirementRatio > 0 && requirementRatio-landRatio > 0.15 {
			recomendations = append(recomendations, fmt.Sprintf("White Mana production may be insufficient\n %.1f%% of nonlands require white, but only\n %.1f%% of lands produce white", requirementRatio*100, landRatio*100))
			requirementNotice = true
		}

		//Blue requirements
		requirementRatio = float64(stats.Mana.ColorRequirements.Blue) / float64(stats.TotalNonlands)
		landRatio = float64(stats.Mana.LandProduction.Blue) / float64(stats.CardTypes.Lands)

		if requirementRatio > 0 && requirementRatio-landRatio > 0.15 {
			recomendations = append(recomendations, fmt.Sprintf("Blue Mana production may be insufficient\n %.1f%% of nonlands require blue, but only\n %.1f%% of lands produce blue", requirementRatio*100, landRatio*100))
			requirementNotice = true
		}

		//Black requirements
		requirementRatio = float64(stats.Mana.ColorRequirements.Black) / float64(stats.TotalNonlands)
		landRatio = float64(stats.Mana.LandProduction.Black) / float64(stats.CardTypes.Lands)

		if requirementRatio > 0 && requirementRatio-landRatio > 0.15 {
			recomendations = append(recomendations, fmt.Sprintf("Black Mana production may be insufficient\n %.1f%% of nonlands require black, but only\n %.1f%% of lands produce black", requirementRatio*100, landRatio*100))
			requirementNotice = true
		}

		//Red requirements
		requirementRatio = float64(stats.Mana.ColorRequirements.Red) / float64(stats.TotalNonlands)
		landRatio = float64(stats.Mana.LandProduction.Red) / float64(stats.CardTypes.Lands)

		if requirementRatio > 0 && requirementRatio-landRatio > 0.15 {
			recomendations = append(recomendations, fmt.Sprintf("Red Mana production may be insufficient\n %.1f%% of nonlands require red, but only\n %.1f%% of lands produce red", requirementRatio*100, landRatio*100))
			requirementNotice = true
		}

		//Green requirements
		requirementRatio = float64(stats.Mana.ColorRequirements.Green) / float64(stats.TotalNonlands)
		landRatio = float64(stats.Mana.LandProduction.Green) / float64(stats.CardTypes.Lands)

		if requirementRatio > 0 && requirementRatio-landRatio > 0.15 {
			recomendations = append(recomendations, fmt.Sprintf("Green Mana production may be insufficient\n %.1f%% of nonlands require green, but only\n %.1f%% of lands produce green", requirementRatio*100, landRatio*100))
			requirementNotice = true
		}

		if requirementNotice == true {
			recomendations = append(recomendations, "Note: Fetchlands are not included in color production calculation")
		}

	} else if stats.TotalNonlands == 0 {
		recomendations = append(recomendations, "No nonland cards detected.")
	} else if stats.CardTypes.Lands == 0 {
		recomendations = append(recomendations, "No Land cards detected.")
	}
	if len(recomendations) == 0 {
		recomendations = append(recomendations, "No major problems detected")
	}
	return recomendations
}
