package scryfall

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDataSetExists(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing_file.txt")
	err := os.WriteFile(existingFile, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("Error writing test file: %v", err)
	}
	exists, err := DataSetExists(existingFile)
	if err != nil {
		t.Fatalf("Error checking if file exists: %v", err)
	}
	if exists != true {
		t.Fatalf("Expected file to exist: %s", existingFile)
	}
	exists, err = DataSetExists(tempDir)
	if err == nil {
		t.Fatalf("Expected error when path is a directory: %v", err)
	}
	if exists != false {
		t.Fatalf("Expected directory not to count as a dataset file: %s", tempDir)
	}
}

func TestDataSetDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	nonExistingFile := filepath.Join(tempDir, "non_existing_file.txt")
	exists, err := DataSetExists(nonExistingFile)
	if err != nil {
		t.Fatalf("Error checking if file exists: %v", err)
	}
	if exists != false {
		t.Fatalf("Expected file to not exist: %s", nonExistingFile)
	}
}

func TestExtractGzip(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "test.json.gz")
	destinationPath := filepath.Join(tempDir, "test.json")
	sourceFile, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Error creating gzip file: %v", err)
	}
	gzipWriter := gzip.NewWriter(sourceFile)
	expected := "{\"name\": \"Sol Ring\"}\n"
	_, err = gzipWriter.Write([]byte(expected))
	if err != nil {
		t.Fatalf("Error writing to gzip file: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatalf("Error closing gzip writer: %v", err)
	}
	err = sourceFile.Close()
	if err != nil {
		t.Fatalf("Error closing gzip file: %v", err)
	}
	err = ExtractGzip(sourcePath, destinationPath)
	if err != nil {
		t.Fatalf("Error extracting gzip file: %v", err)
	}
	content, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf("Error reading extracted file: %v", err)
	}
	if string(content) != expected {
		t.Fatalf("Expected content %q, but got %q", expected, content)
	}
}

func TestExtractGzipInvalidSource(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "notgzip.txt")
	destinationPath := filepath.Join(tempDir, "test.json")
	sourceFile, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Error creating source file: %v", err)
	}
	err = os.WriteFile(sourcePath, []byte("This is not a gzip file"), 0600)
	if err != nil {
		t.Fatalf("Error writing to source file: %v", err)
	}
	err = sourceFile.Close()
	if err != nil {
		t.Fatalf("Error closing source file: %v", err)
	}
	err = ExtractGzip(sourcePath, destinationPath)
	if err == nil {
		t.Fatalf("Expected error when extracting non-gzip file")
	}
}

func TestExtractGzipNonExistentSource(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "nonexistent.gz")
	destinationPath := filepath.Join(tempDir, "test.json")
	err := ExtractGzip(sourcePath, destinationPath)
	if err == nil {
		t.Fatalf("Expected error when extracting from non-existent source file")
	}
}

func TestExtractGzipDestinationExists(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "notgzip.json.gz")
	destinationPath := filepath.Join(tempDir, "test.json")
	sourceFile, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Error creating source file: %v", err)
	}
	gzipWriter := gzip.NewWriter(sourceFile)
	expected := "{\"name\": \"Sol Ring\"}\n"
	_, err = gzipWriter.Write([]byte(expected))
	if err != nil {
		t.Fatalf("Error writing to gzip file: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatalf("Error closing gzip writer: %v", err)
	}
	err = sourceFile.Close()
	if err != nil {
		t.Fatalf("Error closing source file: %v", err)
	}
	err = os.WriteFile(destinationPath, []byte("Existing content"), 0600)
	if err != nil {
		t.Fatalf("Error creating destination file: %v", err)
	}
	err = ExtractGzip(sourcePath, destinationPath)
	if err == nil {
		t.Fatalf("Expected error when extracting to an existing destination file")
	}
	content, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf("Error reading destination file: %v", err)
	}
	if string(content) != "Existing content" {
		t.Fatalf("Expected destination file content to remain unchanged, but got %q", content)
	}

}

func TestExtractGzipPartialWrite(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "test.json.gz")
	destinationPath := filepath.Join(tempDir, "test.json")
	sourceFile, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Error creating gzip file: %v", err)
	}
	gzipWriter := gzip.NewWriter(sourceFile)
	expected := "{\"name\": \"Sol Ring\"}\n"
	_, err = gzipWriter.Write([]byte(expected))
	if err != nil {
		t.Fatalf("Error writing to gzip file: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatalf("Error closing gzip writer: %v", err)
	}
	err = sourceFile.Close()
	if err != nil {
		t.Fatalf("Error closing source file: %v", err)
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("Error reading source file: %v", err)
	}
	if len(data) <= 0 {
		t.Fatalf("Source file is empty")
	}
	newsourcePath := filepath.Join(tempDir, "test_partial.json.gz")
	err = os.WriteFile(newsourcePath, data[:len(data)-1], 0600)
	if err != nil {
		t.Fatalf("Error writing partial gzip file: %v", err)
	}
	err = ExtractGzip(newsourcePath, destinationPath)
	if err == nil {
		t.Fatalf("Expected error when extracting from a partially written gzip file")
	}
}

func TestValidateJSONL(t *testing.T) {
	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "valid.jsonl")
	validContent := "{\"name\": \"Sol Ring\"}\n{\"name\": \"Black Lotus\"}\n"
	err := os.WriteFile(validFile, []byte(validContent), 0600)
	if err != nil {
		t.Fatalf("Error writing valid JSONL file: %v", err)
	}
	err = ValidateJSONL(validFile)
	if err != nil {
		t.Fatalf("Error validating valid JSONL file: %v", err)
	}
}

func TestValidateJSONLInvalid(t *testing.T) {
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.jsonl")
	invalidContent := "{\"name\": \"Sol Ring\"}\n{\"name\":}\n{\"name\": \"Black Lotus\"}\n"
	err := os.WriteFile(invalidFile, []byte(invalidContent), 0600)
	if err != nil {
		t.Fatalf("Error writing invalid JSONL file: %v", err)
	}
	err = ValidateJSONL(invalidFile)
	if err == nil {
		t.Fatalf("Expected error when validating invalid JSONL file")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("Expected error on line 2, but got: %v", err.Error())
	}
}

func TestValidateJSONLEmpty(t *testing.T) {
	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.jsonl")
	err := os.WriteFile(emptyFile, []byte(""), 0600)
	if err != nil {
		t.Fatalf("Error writing empty JSONL file: %v", err)
	}
	err = ValidateJSONL(emptyFile)
	if err == nil {
		t.Fatalf("Expected error when validating empty JSONL file")
	}
}
func TestValidateJSONLBlank(t *testing.T) {
	tempDir := t.TempDir()
	blankFile := filepath.Join(tempDir, "blank.jsonl")
	err := os.WriteFile(blankFile, []byte("\n\n\n"), 0600)
	if err != nil {
		t.Fatalf("Error writing blank JSONL file: %v", err)
	}
	err = ValidateJSONL(blankFile)
	if err == nil {
		t.Fatalf("Expected error when validating blank JSONL file")
	}
}

func TestValidateJSONLNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.jsonl")
	err := ValidateJSONL(nonExistentFile)
	if err == nil {
		t.Fatalf("Expected error when validating non-existent JSONL file")
	}
}

func TestExtractAndValidateJSONL(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "test.json.gz")
	destinationPath := filepath.Join(tempDir, "test.json")
	sourceFile, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Error creating gzip file: %v", err)
	}
	gzipWriter := gzip.NewWriter(sourceFile)
	expected := "{\"name\": \"Sol Ring\"}\n{\"name\": \"Black Lotus\"}\n"
	_, err = gzipWriter.Write([]byte(expected))
	if err != nil {
		t.Fatalf("Error writing to gzip file: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatalf("Error closing gzip writer: %v", err)
	}
	err = sourceFile.Close()
	if err != nil {
		t.Fatalf("Error closing gzip file: %v", err)
	}
	err = ExtractAndValidateJSONL(sourcePath, destinationPath)
	if err != nil {
		t.Fatalf("Error extracting gzip file: %v", err)
	}
	content, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf("Error reading extracted file: %v", err)
	}
	if string(content) != expected {
		t.Fatalf("Expected content %q, but got %q", expected, content)
	}
}

func TestExtractAndValidateJSONLInvalid(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "test.json.gz")
	destinationPath := filepath.Join(tempDir, "test.json")
	sourceFile, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Error creating gzip file: %v", err)
	}
	gzipWriter := gzip.NewWriter(sourceFile)
	expected := "{\"name\": \"Sol Ring\"}\n{\"name\":}\n"
	_, err = gzipWriter.Write([]byte(expected))
	if err != nil {
		t.Fatalf("Error writing to gzip file: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatalf("Error closing gzip writer: %v", err)
	}
	err = sourceFile.Close()
	if err != nil {
		t.Fatalf("Error closing gzip file: %v", err)
	}
	err = ExtractAndValidateJSONL(sourcePath, destinationPath)
	if err == nil {
		t.Fatalf("Expected error when extracting invalid JSONL file")
	}
	_, err = os.Stat(destinationPath)
	if os.IsNotExist(err) == true {
		//Expected
	} else if err != nil {
		t.Fatalf("Error checking if destination file exists: %v", err)
	} else {
		t.Fatalf("Expected destination file to be removed after validation failure, but it still exists")
	}

}

func TestExtractAndValidateJSONLDestinationExists(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "test.json.gz")
	destinationPath := filepath.Join(tempDir, "test.json")
	sourceFile, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Error creating gzip file: %v", err)
	}
	gzipWriter := gzip.NewWriter(sourceFile)
	expected := "{\"name\": \"Sol Ring\"}\n{\"name\": \"Black Lotus\"}\n"
	_, err = gzipWriter.Write([]byte(expected))
	if err != nil {
		t.Fatalf("Error writing to gzip file: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatalf("Error closing gzip writer: %v", err)
	}
	err = sourceFile.Close()
	if err != nil {
		t.Fatalf("Error closing gzip file: %v", err)
	}
	err = os.WriteFile(destinationPath, []byte("Existing content"), 0600)
	if err != nil {
		t.Fatalf("Error creating destination file: %v", err)
	}
	err = ExtractAndValidateJSONL(sourcePath, destinationPath)
	if err == nil {
		t.Fatalf("Expected error when extracting to an existing destination file")
	}
	existingContent, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf("Error reading destination file: %v", err)
	}
	if string(existingContent) != "Existing content" {
		t.Fatalf("Expected destination file content to remain unchanged, but got %q", existingContent)
	}
}
func TestExtractAndValidateJSONLInvalidGzip(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "notgzip.txt")
	destinationPath := filepath.Join(tempDir, "test.json")
	err := os.WriteFile(sourcePath, []byte("This is not a gzip file"), 0600)
	if err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}
	err = ExtractAndValidateJSONL(sourcePath, destinationPath)
	if err == nil {
		t.Fatalf("Expected error when extracting non-gzip file")
	}
	_, err = os.Stat(destinationPath)
	if os.IsNotExist(err) == true {
		//Expected
	} else if err != nil {
		t.Fatalf("Error checking if destination file exists: %v", err)
	} else {
		t.Fatalf("Expected destination file to be removed after extraction failure, but it still exists")
	}
}
