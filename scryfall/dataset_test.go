package scryfall

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func gzipTestData(jsonContent string) ([]byte, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	_, err := gzipWriter.Write([]byte(jsonContent))
	if err != nil {
		return nil, fmt.Errorf("error writing the gzip content: %w", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzipwriter: %w", err)
	}
	return buffer.Bytes(), nil
}

func newGzipTestServer(statusCode int, body []byte) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write(body)
	}))
	return server
}

func TestEnsureDatasetUpdateSuccess(t *testing.T) {
	tempDir := t.TempDir()
	datasetPath := filepath.Join(tempDir, "oracle-cards.jsonl")
	timestampPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")

	oldDataset := `{"name":"Old Card"}` + "\n"
	err := os.WriteFile(datasetPath, []byte(oldDataset), 0600)
	if err != nil {
		t.Fatalf("Could not write existing dataset: %v", err)
	}

	oldTimestamp := time.Now().AddDate(0, 0, -10).Format(time.RFC3339Nano)
	err = os.WriteFile(timestampPath, []byte(oldTimestamp), 0600)
	if err != nil {
		t.Fatalf("Could not write existing timestamp: %v", err)
	}

	newDataset := `{"name":"Sol Ring"}` + "\n"
	compressedData, err := gzipTestData(newDataset)
	if err != nil {
		t.Fatalf("Could not write gzip file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(compressedData)
	}))
	defer server.Close()

	metadata := Metadata{
		DownloadURI: server.URL,
		UpdatedAt:   time.Now(),
	}

	err = EnsureDataset(datasetPath, metadata)
	if err != nil {
		t.Fatalf("Expected dataset update to succeed, got: %v", err)
	}

	content, err := os.ReadFile(datasetPath)
	if err != nil {
		t.Fatalf("Could not read updated dataset: %v", err)
	}
	if string(content) != newDataset {
		t.Fatalf("Expected updated dataset %q, got %q", newDataset, string(content))
	}

	timestamp, err := ReadDatasetTimestamp(timestampPath)
	if err != nil {
		t.Fatalf("Could not read updated timestamp: %v", err)
	}
	if timestamp.Equal(metadata.UpdatedAt) == false {
		t.Fatalf("Expected timestamp %v, got %v", metadata.UpdatedAt, timestamp)
	}
}

func TestEnsureDatasetSuccess(t *testing.T) {
	tempDir := t.TempDir()
	destinationPath := filepath.Join(tempDir, "oracle-cards.jsonl")
	compressed, err := gzipTestData("{\"name\": \"Sol Ring\"}\n{\"name\": \"Black Lotus\"}\n")
	if err != nil {
		t.Fatalf("Error writing gzip file: %v", err)
	}
	server := newGzipTestServer(http.StatusOK, compressed)
	defer server.Close()
	metadata := Metadata{
		DownloadURI: server.URL,
		UpdatedAt:   time.Now(),
	}
	err = EnsureDataset(destinationPath, metadata)
	if err != nil {
		t.Fatalf("Error ensuring dataset: %v", err)
	}
	content, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf("Error reading downloaded dataset: %v", err)
	}
	if string(content) != "{\"name\": \"Sol Ring\"}\n{\"name\": \"Black Lotus\"}\n" {
		t.Fatalf("Expected dataset content to be {\"name\": \"Sol Ring\"}\n{\"name\": \"Black Lotus\"}\n, got :%s", string(content))
	}
	timeStampPath := filepath.Join(filepath.Dir(destinationPath), "oracle-cards-updated-at.txt")
	content, err = os.ReadFile(timeStampPath)
	if err != nil {
		t.Fatalf("Error reading dataset update timestamp file: %v", err)
	}
	if string(content) != metadata.UpdatedAt.Format(time.RFC3339Nano) {
		t.Fatalf("Expected timestamp found to be %v, got: %v", metadata.UpdatedAt.Format(time.RFC3339Nano), string(content))
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Error reading temporary directory contents: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries remaining, got: %d", len(entries))
	}

	foundDataset := false
	foundTimestamp := false

	for _, entry := range entries {
		if entry.Name() == "oracle-cards.jsonl" {
			foundDataset = true
		}
		if entry.Name() == "oracle-cards-updated-at.txt" {
			foundTimestamp = true
		}
	}

	if foundDataset == false || foundTimestamp == false {
		t.Fatalf("Expected both dataset and timestamp files, got: %v", entries)
	}

}

func TestEnsureDatasetDownloadFailure(t *testing.T) {
	tempDir := t.TempDir()
	destinationPath := filepath.Join(tempDir, "oracle-cards.jsonl")
	compressed, err := gzipTestData("{\"name\": \"Sol Ring\"}\n{\"name\": \"Black Lotus\"}\n")
	if err != nil {
		t.Fatalf("Error writing gzip file: %v", err)
	}
	server := newGzipTestServer(http.StatusInternalServerError, compressed)
	defer server.Close()
	metadata := Metadata{
		DownloadURI: server.URL,
		UpdatedAt:   time.Now(),
	}
	err = EnsureDataset(destinationPath, metadata)
	if err == nil {
		t.Fatalf("Expected error due to http Internal Server Error, got none")
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Error reading temporary directory contents: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("Expected no entries after extraction failure, got: %d", len(entries))
	}
}

func TestEnsureDatasetInvalidContent(t *testing.T) {
	tempDir := t.TempDir()
	destinationPath := filepath.Join(tempDir, "oracle-cards.jsonl")
	server := newGzipTestServer(http.StatusOK, []byte("{\"name\": \"Sol Ring\"}\n"))
	defer server.Close()
	metadata := Metadata{
		DownloadURI: server.URL,
		UpdatedAt:   time.Now(),
	}
	err := EnsureDataset(destinationPath, metadata)
	if err == nil {
		t.Fatalf("Expected error due to passing uncompressed data as gzip, got none")
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Error reading temporary directory contents: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("Expected no entries after extraction failure, got: %d", len(entries))
	}
}

func TestEnsureDatasetInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	destinationPath := filepath.Join(tempDir, "oracle-cards.jsonl")
	compressed, err := gzipTestData("{\"name\": }\n{\"Black Lotus\"}\n")
	if err != nil {
		t.Fatalf("Error writing gzip file: %v", err)
	}
	server := newGzipTestServer(http.StatusOK, compressed)
	defer server.Close()
	metadata := Metadata{
		DownloadURI: server.URL,
		UpdatedAt:   time.Now(),
	}
	err = EnsureDataset(destinationPath, metadata)
	if err == nil {
		t.Fatalf("Expected error due to invalid JSON content, got none")
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Error reading temporary directory contents: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("Expected no entries after JSON validation failure, got: %d", len(entries))
	}
}

func TestReadDatasetTimestamp(t *testing.T) {
	timeNow := time.Now()
	timeStamp := timeNow.Format(time.RFC3339Nano)
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	err := os.WriteFile(tempPath, []byte(timeStamp), 0600)
	if err != nil {
		t.Fatalf("Could not write timestamp file: %v", err)
	}
	gotTime, err := ReadDatasetTimestamp(tempPath)
	if err != nil {
		t.Fatalf("Could not read dataset timestamp file: %v", err)
	}
	if timeNow.Equal(gotTime) == false {
		t.Fatalf("Expected to read time %v, got: %v", timeNow, gotTime)
	}
}

func TestReadDatasetTimeStampMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	gotTime, err := ReadDatasetTimestamp(tempPath)
	if err == nil {
		t.Fatalf("Expected error when file is missing, got none")
	}
	if gotTime.Equal(time.Time{}) == false {
		t.Fatalf("Expected to get time %v, got: %v", time.Time{}, gotTime)
	}
}

func TestReadDatasetTimeStampInvalidTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	err := os.WriteFile(tempPath, []byte("not-a-timestamp"), 0600)
	if err != nil {
		t.Fatalf("Could not write timestamp file: %v", err)
	}
	gotTime, err := ReadDatasetTimestamp(tempPath)
	if err == nil {
		t.Fatalf("Expected error when file does not contain a valid timestamp, got none")
	}
	if gotTime.Equal(time.Time{}) == false {
		t.Fatalf("Expected to get time %v, got: %v", time.Time{}, gotTime)
	}
}

func TestCompareDatasetTimestamps(t *testing.T) {
	localTime := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	remoteTime := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	difference, err := CompareDatasetTimestamps(localTime, remoteTime)
	if err != nil {
		t.Fatalf("Got error when comparing timestamps: %v", err)
	}
	if difference != 48*time.Hour {
		t.Fatalf("Expected difference to be 48 hours, got: %v", difference)
	}
}

func TestCompareDatasetTimestampsEqual(t *testing.T) {
	localTime := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	remoteTime := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	difference, err := CompareDatasetTimestamps(localTime, remoteTime)
	if err != nil {
		t.Fatalf("Got error when comparing timestamps: %v", err)
	}
	if difference != 0 {
		t.Fatalf("Expected difference to be 0, got: %v", difference)
	}
}

func TestCompareDatasetTimestampsError(t *testing.T) {
	localTime := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	remoteTime := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	difference, err := CompareDatasetTimestamps(localTime, remoteTime)
	if err == nil {
		t.Fatalf("Expected error when remote time is older than local, got none")
	}
	if difference != 0 {
		t.Fatalf("Expected difference to be 0, got: %v", difference)
	}
}

func TestShouldUpdateDatasetMissingTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	remoteTime := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	tempPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	update, err := ShouldUpdateDataset(tempPath, Metadata{DownloadURI: "placeholder.url", UpdatedAt: remoteTime})
	if err != nil {
		t.Fatalf("Expected no error if timestamp file doesn't exist")
	}
	if update == false {
		t.Fatalf("Expected return value true, got false")
	}
}

func TestShouldUpdateDatasetInvalidTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	remoteTime := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	tempPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	err := os.WriteFile(tempPath, []byte("not-a-timestamp"), 0600)
	if err != nil {
		t.Fatalf("Failed to write temporary timestamp file: %v", err)
	}
	update, err := ShouldUpdateDataset(tempPath, Metadata{DownloadURI: "placeholder.url", UpdatedAt: remoteTime})
	if err != nil {
		t.Fatalf("Expected no error if timestamp is invalid")
	}
	if update == false {
		t.Fatalf("Expected return value true, got false")
	}
}

func TestShouldUpdateDatasetMoreThan7DaysNewer(t *testing.T) {
	tempDir := t.TempDir()
	remoteTime := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	localTime := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	tempPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	err := os.WriteFile(tempPath, []byte(localTime.Format(time.RFC3339Nano)), 0600)
	if err != nil {
		t.Fatalf("Failed to write temporary timestamp file: %v", err)
	}
	update, err := ShouldUpdateDataset(tempPath, Metadata{DownloadURI: "placeholder.url", UpdatedAt: remoteTime})
	if err != nil {
		t.Fatalf("Expected no error when both timestamps are valid")
	}
	if update == false {
		t.Fatalf("Expected return value true, got false")
	}
}

func TestShouldUpdateDatasetLessThan7DaysNewer(t *testing.T) {
	tempDir := t.TempDir()
	remoteTime := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	localTime := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	tempPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	err := os.WriteFile(tempPath, []byte(localTime.Format(time.RFC3339Nano)), 0600)
	if err != nil {
		t.Fatalf("Failed to write temporary timestamp file: %v", err)
	}
	update, err := ShouldUpdateDataset(tempPath, Metadata{DownloadURI: "placeholder.url", UpdatedAt: remoteTime})
	if err != nil {
		t.Fatalf("Expected no error when both timestamps are valid")
	}
	if update == true {
		t.Fatalf("Expected return value false, got true")
	}
}

func TestShouldUpdateDatasetExactly7DaysNewer(t *testing.T) {
	tempDir := t.TempDir()
	remoteTime := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	localTime := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	tempPath := filepath.Join(tempDir, "oracle-cards-updated-at.txt")
	err := os.WriteFile(tempPath, []byte(localTime.Format(time.RFC3339Nano)), 0600)
	if err != nil {
		t.Fatalf("Failed to write temporary timestamp file: %v", err)
	}
	update, err := ShouldUpdateDataset(tempPath, Metadata{DownloadURI: "placeholder.url", UpdatedAt: remoteTime})
	if err != nil {
		t.Fatalf("Expected no error when both timestamps are valid")
	}
	if update == true {
		t.Fatalf("Expected return value false, got true")
	}
}
