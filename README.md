# MTG DECK ANALYZER
#### Description:
MTG DECK ANALYZER is a simple terminal run tool that analyzes a plain-text exported decklist(from websites such as [Moxfield](https://moxfield.com/) or https://edhrec.com/) and return an analysis of:\
* The number of cards of each card type (Lands, Instants, Creatures, etc.)\
* The number of cards that require a certain mana type to be cast (White, Blue, etc.)\
* The number of lands that can produce each type of mana (fetch lands counted separately)\
* The number of nonland cards that fit into certain roles(ramp, card draw, removal)
* The probability to have at least one card from that specific role in your opening 8 cards (7 + 1 from turn 1 draw)
* The probability to have at least 2 and at least 3 lands in your opening 8 cards
* A short conclusion that detects issues with removal, the total land count, and the land production count for each color\

## Usage

The application requires Go to be installed when running from source.

1. Place a Commander decklist exported as a plain-text file inside the `decklists/` directory.
2. Open a terminal in the project's root directory.
3. Run the application using:

   ```bash
   go run .
   ```

4. Enter the decklist's filename without the `.txt` extension.
5. Enter the name of the commander when prompted.
6. The application will load the Scryfall dataset, analyze the deck, and display the results in the terminal.

An internet connection is required for the initial dataset download and for checking Scryfall's metadata when the application starts.

## Files
### **analysis/analyzer.go**
This file contains the main deck analysis logic. It defines the structs used to store the statistics about card types, mana requirements, mana curve and card roles.

The analyzer processes each card using its Scryfall data. Card roles are identified using patterns made with regular expressions, applied to each card's oracle text. Some complex card effects may not be categorized perfectly due to the limitations of go's regular expression system.
### **deck/deck.go**
This file contains the main card search and deck parsing logic. It defines structs that hold card quantities and card data.

The parser receives a decklist imported as plain text and returns a slice containing each card's name and their quantity. The card that is identified as the commander is removed from the decklist, then the cards are resolved and matched by name to their counterpart from the dataset.
### **scryfall/cards.go**
This file defines the structs used to store card data from Scryfall's JSON dataset. It also contains the extraction and validation logic for the downloaded card data.

The extractor decompresses the downloaded gzip file into a JSONL file, which is then validated line by line to ensure that the data is correctly formatted before being loaded into the application.
### **scryfall/dataset.go**
This file has multiple helper functions that read the timestamp of when the local dataset was updated, compare it with the timestamp of the remote dataset, and decide whether the local version is up-to-date.

The "ensurer"(EnsureDataset function) downloads the remote dataset to a temporary file, backs up the local dataset if it exists, then moves the new dataset to its permanent location. If moving the new dataset fails, the rollback mechanism moves the backup back to its original location.

This took the longest to implement in order to make sure that no data was lost, either by downloading directly over an existing data set, or due to failures in moving the files.
### **scryfall/download.go**
This file contains the main download logic. It has helpers that find the download URI from the metadata provided by Scryfall's API.

The downloader makes a request to a known bulk data endpoint and returns a download URI as well as metadata to check if the local dataset needs updating. After getting the URI, it downloads the compressed JSONL file to a temporary location.

### **scryfall/loader.go**
This file contains the dataset loading logic.

The loader receives the dataset as a JSONL file, then using the Card structs defined in [cards.go](#scryfallcardsgo), loads the dataset into a map using the card's name as the key.

Ran into an issue with multi-faced cards where at first they would not be [resolved](#deckdeckgo) properly due to differences in Scryfall's naming convention and other websites' naming conventions. Solved this by adding additional map entries for individual card-faces names.

### **report/report.go**
This file contains the terminal reporting logic. 

It receives the statistics and probabilities calculated by the [analyzer](#analysisanalyzergo) and prints them to the terminal using a simple formatting style. The report organizes the results into sections for readability, and displays simple recomendations based on the statistics. 

Used Go's built-in fmt package instead of an external terminal UI library because the application's output is relatively simple and doesn't require complex visual elements or interactive components.

### **main.go**
This file is the entry point of the application and connects the different packages.

It coordinates the process downloading or updating the Scryfall dataset, parsing through the user's decklist, removing the commander from the decklist for analysis purposes, resolving the cards, calculating statistics and probabilities and generating the final report.

### **data/**
This folder stores the downloaded Scryfall dataset and its last-update timestamp. The application expects this directory to exist because it uses a relative path to access the dataset.

The dataset is downloaded and updated automatically when necessary.

### **decklists/**
This folder contains the user's decklists as .txt files. Users can add their own decklists and provide the filename, without the .txt extension, when prompted by the application.

### **testdata/**
This folder contains a few sample decklists, created with the sole purpose of testing the accuracy of the [analyzer's](#analysisanalyzergo) role counting.

It also contains a sample of dataset consisting in 2 card entries, used for checking if cards are properly loaded.

### **Tests(*_test.go)**
All the files ending with _test.go contain automated tests for the application. Most test files correspond to a specific source file, and they verify both succesful inputs and various error conditions.

The only exceptions are probabilty_test.go and integration_test.go which both test functions from [analyzer.go](#analysisanalyzergo). 

The probability test was separated from the normal test file because it serves a largely different function, checking a hypergeometric probability calculator. 

Integration_test.go tests the integration of multiple functions from different packages together by loading the dataset, and parsing through [decklists](#testdata) specifically made to the test the analyzer.

All automated tests can be executed at once by using:
``` bash
go test ./...
```

## Design Decisions
### **Bulk data instead of API requests**
The decision of using bulk data instead of calling Scryfall's card API for each card was made for two reasons:
1. Speed\
Scryfall's API has rate limits, meaning that retrieving information for every card in a Commander deck would require multiple network requests and introduce unnecessary delays.\
By downloading the dataset and loading the cards into a Go map, the application can retrieve card information by name without making additional API requests. This makes repeated card lookups much faster.

2. Offline availability\
Making an API request for every card would make the analysis dependent on a stable internet connection throughout the process.\
By storing the dataset locally, card information remains available even when individual API requests cannot be made. However, the current version still checks Scryfall's metadata when starting, so a network connection is required for normal operation.\
This approach also allows the application to reuse previously downloaded data instead of downloading the entire dataset every time it runs.

### **Hypergeometric probability**
Consistency is one of the most important deck-building aspects in trading card games. Having an opening hand that allows you to make plays is essential.\

By implementing a hypergeometric probability calculator alongside our [role analysis](#analysisanalyzergo) we can see how consistently a deck starts the game with the cards it needs in hand.\

I chose hypergeometric distribution because cards are drawn without replacement, and is the standard tool for calculating deck consistency in trading card games.

## Limitations and future improvements
### **Single Commanders**
Some Commander decks can have multiple commanders under specific conditions, such as the Partner mechanic.

The current version supports only one commander because supporting multiple commanders would require additional logic in the [main application](#maingo) for user input, as well as changes to the [commander extraction](#deckdeckgo) process.

### **Alternate Art Cards**
Some cards have special printed versions that use different names from their original Oracle names.

These alternate printed names are not always included in the Scryfall Oracle dataset, which contains one canonical record per unique Oracle card. As a result, the application may fail to resolve these cards unless their standard Oracle names are used.
### **Regular Expression based detection**
The analyzer uses regular expressions to identify cards that fulfill certain roles based on their Oracle text.

This approach works well when cards use recognizable wording, but not all cards with similar effects are worded the same way. Because regular expressions match text patterns rather than interpret gameplay mechanics, some cards may be incorrectly categorized or not detected at all.