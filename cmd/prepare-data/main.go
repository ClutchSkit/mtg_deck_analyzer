package main

import (
	"fmt"
	"os"

	"github.com/ClutchSkit/mtg_deck_analyzer/scryfall"
)

func main() {
	err := scryfall.PrepareDataset("data/oracle-cards.jsonl")
	if err != nil {
		fmt.Printf("Preparing dataset failed: %v\n", err)
		os.Exit(1)
	}
}
