package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Uninstall removes files according to the mapping configuration
func Uninstall(config *Config) error {
	fmt.Println("Removing installation...")

	for _, mapping := range config.Mappings {
		if err := uninstallPath(mapping.To); err != nil {
			fmt.Printf("  Error processing %s: %v\n", mapping.To, err)
			// Continue with other files
		}
	}

	return nil
}

// uninstallPath removes a single path
func uninstallPath(mappingPath string) error {
	targetPath := expandPath(mappingPath)

	// Check if the target exists
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, skip silently
			return nil
		}
		return fmt.Errorf("failed to stat: %w", err)
	}

	// Handle directories
	if info.IsDir() {
		return uninstallDirectory(targetPath)
	}

	// Handle files
	return uninstallFile(targetPath)
}

// uninstallDirectory removes a directory if safe to do so
func uninstallDirectory(path string) error {
	if !canRemoveDirectory(path) {
		fmt.Printf("  Skipped %s (protected directory)\n", path)
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		fmt.Printf("  Failed to remove %s: %v\n", path, err)
		return err
	}

	fmt.Printf("  Removed %s (directory)\n", path)
	return nil
}

// uninstallFile removes a single file
func uninstallFile(path string) error {
	if err := os.Remove(path); err != nil {
		fmt.Printf("  Failed to remove %s: %v\n", path, err)
		return err
	}

	fmt.Printf("  Removed %s\n", path)
	return nil
}

// canRemoveDirectory checks if a directory can be safely removed
// Only directories under home directory (excluding home itself) are allowed to be removed
func canRemoveDirectory(path string) bool {
	homeDir := getRealHomeDir()
	if homeDir == "" {
		// If we can't get home directory, don't remove anything
		return false
	}

	// Ensure both paths are clean and absolute for proper comparison
	cleanPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}

	cleanHome, err := filepath.Abs(filepath.Clean(homeDir))
	if err != nil {
		return false
	}

	// Don't remove home directory itself
	if cleanPath == cleanHome {
		return false
	}

	// Check if the path is under home directory
	return strings.HasPrefix(cleanPath, cleanHome+string(filepath.Separator))
}
