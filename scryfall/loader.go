package scryfall

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func LoadCards(path string) (map[string]Card, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	cards := make(map[string]Card)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		var card Card
		err := json.Unmarshal([]byte(line), &card)
		if card.Object != "card" {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal line number %d: %w", lineNumber, err)
		}
		if card.Name == "" {
			return nil, fmt.Errorf("missing card name on line %d", lineNumber)
		}
		cards[card.Name] = card

		for _, face := range card.CardFaces {
			if face.Name == "" {
				continue
			}

			if _, exists := cards[face.Name]; exists == false {
				cards[face.Name] = card
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cards, nil
}
