package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// VerifyArchiveStructure verifies that all expected files exist in the extracted archive
func VerifyArchiveStructure(extractedDir string, config *Config) error {
	fmt.Println("Verifying archive structure...")

	allValid := true
	for _, mapping := range config.Mappings {
		sourcePath := filepath.Join(extractedDir, mapping.From)

		// Check if the source file/directory exists
		if _, err := os.Stat(sourcePath); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("  [FAIL] %s not found\n", mapping.From)
				allValid = false
			} else {
				return fmt.Errorf("failed to check %s: %w", mapping.From, err)
			}
		} else {
			fmt.Printf("  [OK] %s found\n", mapping.From)
		}
	}

	if !allValid {
		return fmt.Errorf("archive structure does not match mapping file")
	}

	return nil
}
