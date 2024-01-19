package utils

import (
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	tempFileName := tempFile.Name()

	// Close the file immediately
	tempFile.Close()

	// Check if the file exists
	if !FileExists(tempFileName) {
		t.Errorf("FileExists(%s) = false; want true", tempFileName)
	}

	// Remove the file
	if err := os.Remove(tempFileName); err != nil {
		t.Fatal(err)
	}

	// Check again after removal
	if FileExists(tempFileName) {
		t.Errorf("FileExists(%s) = true; want false after file removal", tempFileName)
	}
}
