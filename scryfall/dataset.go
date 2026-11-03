package scryfall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EnsureDataset checks if the dataset exists at it's intended permanent path, and if not, downloads it from Scryfall and extracts it.
// Pass the downloadURI field using the GetOracleCardsGzipURI function
func EnsureDataset(permanentPath string, metadata Metadata) (err error) {
	exists, err := DataSetExists(permanentPath)
	if err != nil {
		return err
	}
	rollbackFailed := false
	tempDownloadPath, err := DownloadToTempFile(metadata.DownloadURI, filepath.Dir(permanentPath))
	if err != nil {
		return err
	}
	defer func() {
		cleanupErr := os.Remove(tempDownloadPath)
		if cleanupErr != nil && os.IsNotExist(cleanupErr) == false {
			err = errors.Join(err, fmt.Errorf("failed to remove temporary download file: %w", cleanupErr))
		}
	}()
	tempDir, err := os.MkdirTemp(filepath.Dir(permanentPath), "scryfall-tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		if rollbackFailed == false {
			tempCleanupErr := os.RemoveAll(tempDir)
			if tempCleanupErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to remove temporary directory: %w", tempCleanupErr))
			}
		}
	}()
	backupDataPath := filepath.Join(tempDir, "backup-oracle-cards.jsonl")
	backupTimePath := filepath.Join(tempDir, "backup-oracle-cards-updated-at.txt")
	tempExtractPath := filepath.Join(tempDir, "oracle-cards.jsonl")
	err = ExtractAndValidateJSONL(tempDownloadPath, tempExtractPath)
	if err != nil {
		return err
	}
	timestamp := metadata.UpdatedAt.Format(time.RFC3339Nano)
	tempTimestampPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	timestampPath := filepath.Join(filepath.Dir(permanentPath), "oracle-cards-updated-at.txt")
	err = os.WriteFile(tempTimestampPath, []byte(timestamp), 0600)
	if err != nil {
		return fmt.Errorf("failed to write temporary dataset update timestamp file: %w", err)
	}
	succeeded := false
	datasetBackedUp := false
	timestampBackedUp := false
	defer func() {
		if succeeded == false {
			if datasetBackedUp == true {
				deleteErr := os.Remove(permanentPath)
				if deleteErr != nil && errors.Is(deleteErr, os.ErrNotExist) == false {
					err = errors.Join(err, fmt.Errorf("failed to delete dataset from permanentPath, moving the backup back may fail: %w", deleteErr))
					rollbackFailed = true
				}
				backupMoveErr := os.Rename(backupDataPath, permanentPath)
				if backupMoveErr != nil {
					err = errors.Join(err, fmt.Errorf("failed to move dataset backup back to permanent path, dataset was lost: %w", backupMoveErr))
					rollbackFailed = true
				}
			}
			if timestampBackedUp == true {
				deleteErr := os.Remove(timestampPath)
				if deleteErr != nil && errors.Is(deleteErr, os.ErrNotExist) == false {
					err = errors.Join(err, fmt.Errorf("failed to delete timestamp file from permanentPath, moving the timestamp back may fail: %w", deleteErr))
					rollbackFailed = true
				}
				backupMoveErr := os.Rename(backupTimePath, timestampPath)
				if backupMoveErr != nil {
					err = errors.Join(err, fmt.Errorf("failed to move timestamp backup back to permanent path, timestamp was lost: %w", backupMoveErr))
					rollbackFailed = true
				}
			}
		}
	}()
	if exists == true {
		err = os.Rename(permanentPath, backupDataPath)
		if err != nil {
			return fmt.Errorf("failed to move existing data set to backup path: %w", err)
		}
		datasetBackedUp = true
		err = os.Rename(timestampPath, backupTimePath)
		if err != nil && errors.Is(err, os.ErrNotExist) == false {
			return fmt.Errorf("failed to move existing timestamp file to backup path: %w", err)
		} else if err == nil {
			timestampBackedUp = true
		}
	}

	err = os.Rename(tempExtractPath, permanentPath)
	if err != nil {
		return err
	}
	err = os.Rename(tempTimestampPath, timestampPath)
	if err != nil {
		return err
	}

	succeeded = true
	return nil
}

func ReadDatasetTimestamp(path string) (time.Time, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read timestamp file; %w", err)
	}
	parsedTime, err := time.Parse(time.RFC3339Nano, string(content))
	if err != nil {
		return time.Time{}, InvalidTimestampError{
			Message: "Invalid timestamp format",
		}
	}
	return parsedTime, nil
}

func CompareDatasetTimestamps(localTime time.Time, remoteTime time.Time) (time.Duration, error) {
	if remoteTime.After(localTime) == true {
		return remoteTime.Sub(localTime), nil
	} else if remoteTime.Equal(localTime) {
		return remoteTime.Sub(localTime), nil
	}
	return time.Duration(0), fmt.Errorf("scryfall dataset appears to be older than installed dataset")
}

func ShouldUpdateDataset(timestampPath string, metadata Metadata) (bool, error) {
	localTime, err := ReadDatasetTimestamp(timestampPath)
	var invalidTimestampErr InvalidTimestampError
	if err != nil && errors.Is(err, os.ErrNotExist) == true {
		return true, nil
	} else if err != nil && errors.As(err, &invalidTimestampErr) == true {
		return true, nil
	} else if err != nil {
		return false, err
	}
	remoteTime := metadata.UpdatedAt
	difference, err := CompareDatasetTimestamps(localTime, remoteTime)
	if err != nil {
		return false, err
	}
	const updateThreshold = 7 * 24 * time.Hour
	if difference > updateThreshold {
		return true, nil
	}
	return false, nil
}
