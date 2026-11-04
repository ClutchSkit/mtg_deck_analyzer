package deck

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ClutchSkit/mtg_deck_analyzer/scryfall"
)

type CardEntry struct {
	Quantity int
	Name     string
}

type FoundCard struct {
	Quantity int
	Card     scryfall.Card
}

type MissingCardsError struct {
	Names []string
}

func (e *MissingCardsError) Error() string {
	return fmt.Sprintf("the following cards were not found: %v ", e.Names)
}

func ParseDeckLine(line string) (CardEntry, error) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return CardEntry{}, fmt.Errorf("Bad line formatting")
	}

	quantity, err := strconv.Atoi(parts[0])
	if err != nil {
		return CardEntry{}, fmt.Errorf("Quantity is not an integer")
	}
	if quantity <= 0 {
		return CardEntry{}, fmt.Errorf("Card quantity must be a positive integer")
	}
	name := strings.TrimSpace(parts[1])
	if name == "" {
		return CardEntry{}, fmt.Errorf("Card name not supplied")
	}
	var card CardEntry
	card.Quantity = quantity
	card.Name = name

	return card, nil
}

func ParseDeckList(path string) ([]CardEntry, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	var cards []CardEntry
	var parseErrors []string
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		card, err := ParseDeckLine(line)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("Line %d: %v", lineNumber, err))
			continue
		} else {
			cards = append(cards, card)
		}
	}
	if scanner.Err() != nil {
		return nil, nil, scanner.Err()
	}

	return cards, parseErrors, nil

}

func ResolveCards(deckList []CardEntry, cardDatabaseMap map[string]scryfall.Card) ([]FoundCard, error) {
	var missingErr MissingCardsError
	foundCards := make([]FoundCard, 0)
	for _, entry := range deckList {
		card, exists := cardDatabaseMap[entry.Name]
		if exists == false {
			missingErr.Names = append(missingErr.Names, entry.Name)
			continue
		}
		foundCards = append(foundCards, FoundCard{
			Quantity: entry.Quantity,
			Card:     card,
		})
	}

	if len(missingErr.Names) > 0 {
		return foundCards, &missingErr
	}
	return foundCards, nil
}

func ExtractCommander(deckList []CardEntry, commanderName string) ([]CardEntry, CardEntry, error) {
	for i, entry := range deckList {
		if entry.Name == commanderName {
			commander := entry
			deckList = append(deckList[:i], deckList[i+1:]...)
			return deckList, commander, nil
		}
	}
	return deckList, CardEntry{}, fmt.Errorf("Commander not found")
}
