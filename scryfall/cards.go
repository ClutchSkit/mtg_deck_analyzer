package scryfall

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Card struct {
	Name          string     `json:"name"`
	ManaCost      string     `json:"mana_cost"`
	CMC           float64    `json:"cmc"`
	TypeLine      string     `json:"type_line"`
	Colors        []string   `json:"colors"`
	ColorIdentity []string   `json:"color_identity"`
	OracleText    string     `json:"oracle_text"`
	Layout        string     `json:"layout"`
	Object        string     `json:"object"`
	CardFaces     []CardFace `json:"card_faces"`
	ProducedMana  []string   `json:"produced_mana"`
}

type CardFace struct {
	Name         string   `json:"name"`
	ManaCost     string   `json:"mana_cost"`
	TypeLine     string   `json:"type_line"`
	Colors       []string `json:"colors"`
	OracleText   string   `json:"oracle_text"`
	ProducedMana []string `json:"produced_mana"`
}

func DataSetExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) == true {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if fileInfo.Mode().IsRegular() == true {
		return true, nil
	}
	return false, fmt.Errorf("path exists but is not a regular file: %s", path)
}

// ExtractGzip extracts a gzip-compressed file from sourcePath to destinationPath.
// It returns an error if opening, decompression, writing, or closing fails.
// The destination is created with permissions 0600 and must not already exist.
// If extraction fails after the destination is created, a partial file may remain.
// The caller is responsible for cleaning up that partial file.
func ExtractGzip(sourcePath string, destinationPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(destinationFile, gzipReader)
	closeErr := destinationFile.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func ValidateJSONL(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	lineNumber := 0
	validRecords := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if json.Valid([]byte(line)) != true {
			return fmt.Errorf("invalid JSON on line %d", lineNumber)
		}
		validRecords++
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	if validRecords == 0 {
		return fmt.Errorf("no valid JSON records found")
	}
	return nil
}

// Extracts a compressed JSONL file and validates it.
// The dataset manager has to supply a temporary destination path for the extracted file, which must not already exist.
func ExtractAndValidateJSONL(sourcePath string, destinationPath string) error {
	destinationExists, err := DataSetExists(destinationPath)
	if err != nil {
		return err
	}
	if destinationExists {
		return fmt.Errorf("destination file already exists: %s", destinationPath)
	}
	err = ExtractGzip(sourcePath, destinationPath)
	if err != nil {
		cleanupErr := os.Remove(destinationPath)
		if cleanupErr != nil && os.IsNotExist(cleanupErr) == false {
			return fmt.Errorf("extraction failed: %v; additionally, failed to remove destination file: %v", err, cleanupErr)
		}
		return err
	}
	err = ValidateJSONL(destinationPath)
	if err != nil {
		cleanupErr := os.Remove(destinationPath)
		if cleanupErr != nil && os.IsNotExist(cleanupErr) == false {
			return fmt.Errorf("validation failed: %v; additionally, failed to remove destination file: %v", err, cleanupErr)
		}
		return err
	}
	return nil
}
