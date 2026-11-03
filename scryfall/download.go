package scryfall

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Metadata struct {
	DownloadURI string
	UpdatedAt   time.Time
}

type InvalidTimestampError struct {
	Message string
}

func (e InvalidTimestampError) Error() string {
	return e.Message
}

// Change the directory path back to "data/" after creating the main workflow
func DownloadToTempFile(url string, directory string) (filePath string, err error) {
	gzipFile, err := os.CreateTemp(directory, "scryfall-*.jsonl.gz")
	if err != nil {
		return "", err
	}
	keepFile := false
	fileClosed := false
	defer func() {
		if fileClosed == false {
			closeErr := gzipFile.Close()
			if closeErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to close temporary file: %w", closeErr))
			}
		}
		if keepFile == false {
			cleanupErr := os.Remove(gzipFile.Name())
			if cleanupErr != nil && os.IsNotExist(cleanupErr) == false {
				err = errors.Join(err, fmt.Errorf("failed to remove temporary file: %w", cleanupErr))
			}
		}
	}()

	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("failed to download file: received status code %d", response.StatusCode)
	}
	_, err = io.Copy(gzipFile, response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to temporary file: %w", err)
	}
	err = gzipFile.Close()
	fileClosed = true
	if err != nil {
		return "", fmt.Errorf("failed to close temporary file: %w", err)
	}
	keepFile = true
	return gzipFile.Name(), nil
}

func GetOracleCardsGzipURI() (Metadata, error) {
	const endpoint = "https://api.scryfall.com/bulk-data"
	metadata, err := FindURI(endpoint)
	if err != nil {
		return Metadata{}, err
	}
	return metadata, nil
}

func FindURI(endpoint string) (Metadata, error) {
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to create bulk data request: %w", err)
	}

	request.Header.Set("User-Agent", "MTGDeckAnalyzer/1.0")
	request.Header.Set("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to fetch bulk data: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseBody, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return Metadata{}, fmt.Errorf("failed to fetch bulk data: status code %d; could not read error response: %w", response.StatusCode, readErr)
		}

		return Metadata{}, fmt.Errorf(
			"failed to fetch bulk data: received status code %d: %s",
			response.StatusCode,
			string(responseBody),
		)
	}
	jsonData, err := io.ReadAll(response.Body)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to read bulk data response: %w", err)
	}
	type BulkDataEntry struct {
		Type        string    `json:"type"`
		DownloadURI string    `json:"jsonl_download_uri"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	type BulkDataResponse struct {
		Data []BulkDataEntry `json:"data"`
	}
	var bulkData BulkDataResponse
	err = json.Unmarshal(jsonData, &bulkData)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to parse bulk data response: %w", err)
	}
	for _, entry := range bulkData.Data {
		if entry.Type == "oracle_cards" {
			if entry.DownloadURI == "" {
				return Metadata{}, fmt.Errorf("no download uri found for entry of type oracle_cards")
			}
			if entry.UpdatedAt.IsZero() == true {
				return Metadata{}, fmt.Errorf("no updated_at field found for entry of type oracle_cards")
			}
			var metadata Metadata
			metadata.DownloadURI = entry.DownloadURI
			metadata.UpdatedAt = entry.UpdatedAt
			return metadata, nil
		}
	}
	return Metadata{}, fmt.Errorf("no entry with type oracle_cards was found")
}
