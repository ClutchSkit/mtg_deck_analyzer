package scryfall

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestServer(statusCode int, body string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
	return server
}

func TestDownloadToTempFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	},
	))
	defer server.Close()
	tempDir := t.TempDir()
	filePath, err := DownloadToTempFile(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Error downloading: %v", err)
	}
	if filepath.Dir(filePath) != tempDir {
		t.Fatalf("Expected file in %s, got %s", tempDir, filepath.Dir(filePath))
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Error reading downloaded file: %v", err)
	}
	if string(contents) != "hello" {
		t.Fatalf("Downloaded file contents mismatch: expected 'hello', got '%s'", string(contents))
	}
}

func TestDownloadToTempFileInternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("hello"))
	},
	))
	defer server.Close()
	tempDir := t.TempDir()
	_, err := DownloadToTempFile(server.URL, tempDir)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
	matches, matchErr := filepath.Glob(filepath.Join(tempDir, "scryfall-*.jsonl.gz"))
	if matchErr != nil {
		t.Fatalf("Error checking for leftover files: %v", matchErr)
	}
	if len(matches) > 0 {
		t.Fatalf("Temporary file was not cleaned up after failed download: %v", matches)
	}

}

func TestDownloadToTempFileNetworkError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	ip := listener.Addr().String()
	err = listener.Close()
	if err != nil {
		t.Fatalf("Failed to close listener: %v", err)
	}
	tempDir := t.TempDir()
	_, err = DownloadToTempFile("http://"+ip, tempDir)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
	matches, matchErr := filepath.Glob(filepath.Join(tempDir, "scryfall-*.jsonl.gz"))
	if matchErr != nil {
		t.Fatalf("Error checking for leftover files: %v", matchErr)
	}
	if len(matches) > 0 {
		t.Fatalf("Temporary file was not cleaned up after failed download: %v", matches)
	}
}

func TestFindURI(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"data":[{"type":"oracle_cards","jsonl_download_uri":"placeholder.url", "updated_at": "2026-10-29T09:02:02.225+00:00"}]}`)
	defer server.Close()
	metadata, err := FindURI(server.URL)
	if err != nil {
		t.Fatalf("Error finding download uri %v", err)
	}
	expectedTime, err := time.Parse(time.RFC3339Nano, "2026-10-29T09:02:02.225+00:00")
	if err != nil {
		t.Fatalf("Failed to parse expected timestamp: %v", err)
	}
	if metadata.DownloadURI != "placeholder.url" {
		t.Fatalf("Expected placeholder.url, got %s", metadata.DownloadURI)
	}
	if metadata.UpdatedAt != expectedTime {
		t.Fatalf("Expected %v, got:%v", expectedTime, metadata.UpdatedAt)
	}
}

func TestFindURINoOracleCards(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"data":[{"type":"oracle","jsonl_download_uri":"placeholder.url", "updated_at": "2026-10-29T09:02:02.225+00:00"}]}`)
	defer server.Close()
	metadata, err := FindURI(server.URL)
	if err == nil {
		t.Fatalf("Expected error due to missing oracle_cards entry but got none")
	}
	if metadata != (Metadata{}) {
		t.Fatalf("Expected no metadata due to missing oracle_cards but got: %+v", metadata)
	}
}

func TestFindURIEmptyDownloadURI(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"data":[{"type":"oracle_cards","jsonl_download_uri":"", "updated_at": "2026-10-29T09:02:02.225+00:00"}]}`)
	defer server.Close()
	metadata, err := FindURI(server.URL)
	if err == nil {
		t.Fatalf("Expected error due to missing uri but got none")
	}
	if metadata != (Metadata{}) {
		t.Fatalf("Expected no metadata due to missing uri but got: %+v", metadata)
	}
}

func TestFindURIEmptyUpdatedAt(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"data":[{"type":"oracle_cards","jsonl_download_uri":"placeholder.url", "updated_at": ""}]}`)
	defer server.Close()
	metadata, err := FindURI(server.URL)
	if err == nil {
		t.Fatalf("Expected error due to missing update time but got none")
	}
	if metadata != (Metadata{}) {
		t.Fatalf("Expected no metadata due to missing update time but got: %+v", metadata)
	}
}

func TestFindURIInvalidStatusCode(t *testing.T) {
	server := newTestServer(http.StatusInternalServerError, `{"data":[{"type":"oracle_cards","jsonl_download_uri":"placeholder.url", "updated_at": "2026-10-29T09:02:02.225+00:00"}]}`)
	defer server.Close()
	metadata, err := FindURI(server.URL)
	if err == nil {
		t.Fatalf("Expected error due to invalid status code")
	}
	if metadata != (Metadata{}) {
		t.Fatalf("Expected no metadata but got: %+v", metadata)
	}
}

func TestFindURIInvalidJSON(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"data":}`)
	defer server.Close()
	metadata, err := FindURI(server.URL)
	if err == nil {
		t.Fatalf("Expected error due to invalid json")
	}
	if metadata != (Metadata{}) {
		t.Fatalf("Expected no metadata but got: %+v", metadata)
	}
}

func TestFindURINetworkError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	ip := listener.Addr().String()
	err = listener.Close()
	if err != nil {
		t.Fatalf("Failed to close listener: %v", err)
	}
	metadata, err := FindURI("http://" + ip)
	if err == nil {
		t.Fatalf("Expected error due to network error but got none")
	}
	if metadata != (Metadata{}) {
		t.Fatalf("Expected no metadata but got: %+v", metadata)
	}
}
