package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ClutchSkit/mtg_deck_analyzer/analysis"
	"github.com/ClutchSkit/mtg_deck_analyzer/deck"
	"github.com/ClutchSkit/mtg_deck_analyzer/report"
	"github.com/ClutchSkit/mtg_deck_analyzer/scryfall"
)

func main() {
	datasetMetadata, err := scryfall.GetOracleCardsGzipURI()
	if err != nil {
		fmt.Printf("Could not retrieve metadata: %v", err)
		return
	}
	err = scryfall.EnsureDataset("data/oracle-cards.jsonl", datasetMetadata)
	if err != nil {
		fmt.Printf("Retrieving the dataset from scryfall failed: %v", err)
		return
	}
	dataset, err := scryfall.LoadCards("data/oracle-cards.jsonl")
	if err != nil {
		fmt.Printf("Loading the dataset into a map failed: %v", err)
		return
	}
	fmt.Printf("Loaded %d cards from dataset\n", len(dataset))

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\nEnter deck name without file extension (Alternate Art Cards will make the program fail): ")
	deckName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read deck name: %v\n", err)
		return
	}

	deckName = strings.TrimSpace(deckName)
	deckPath := filepath.Join("decklists", deckName+".txt")

	deckList, parseErr, err := deck.ParseDeckList(deckPath)
	if err != nil {
		fmt.Printf("Failed to parse deckList: %v", err)
		return
	}
	if len(parseErr) != 0 {
		for _, parse := range parseErr {
			fmt.Printf("%s\n", parse)
		}
		return
	}

	fmt.Printf("Loaded %d decklist entries\n", len(deckList))

	fmt.Print("\nEnter commander name (V1.0 ONLY SUPPORTS SINGLE COMMANDERS): ")
	commanderName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read commander name: %v\n", err)
		return
	}

	commanderName = strings.TrimSpace(commanderName)

	deckList, _, err = deck.ExtractCommander(deckList, commanderName)
	if err != nil {
		fmt.Printf("Failed to remove commander from decklist: %v", err)
		return
	}
	foundCards, missingErr := deck.ResolveCards(deckList, dataset)
	if missingErr != nil {
		fmt.Printf("Failed to resolve decklist cards: %v\n", missingErr)
		return
	}

	deckStats := analysis.DeckAnalysis(foundCards)

	probabilities, err := analysis.CalculateProbabilities(deckStats)
	if err != nil {
		fmt.Printf("Failed to calculate probabilities: %v", err)
		return
	}

	report.GenerateReport(commanderName, deckStats, probabilities)
}
