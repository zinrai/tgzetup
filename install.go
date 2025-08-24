package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// Install downloads, extracts, verifies and installs from the given URL
func Install(url string, config *Config, keepTemp bool) error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "tgzetup-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clean up temp directory unless keepTemp is set
	if !keepTemp {
		defer func() {
			fmt.Printf("Cleaning up temporary directory...\n")
			os.RemoveAll(tempDir)
		}()
	} else {
		fmt.Printf("Temporary directory: %s\n", tempDir)
	}

	// Download archive
	archivePath := filepath.Join(tempDir, "archive.tar.gz")
	if err := DownloadArchive(url, archivePath); err != nil {
		return err
	}

	// Extract archive
	extractDir := filepath.Join(tempDir, "extracted")
	if err := ExtractTarGz(archivePath, extractDir); err != nil {
		return err
	}

	// Verify structure
	if err := VerifyArchiveStructure(extractDir, config); err != nil {
		return err
	}

	// Install files
	fmt.Println("Installing files...")
	for _, mapping := range config.Mappings {
		if err := installMapping(extractDir, mapping); err != nil {
			return fmt.Errorf("failed to install %s: %w", mapping.From, err)
		}
	}

	if keepTemp {
		fmt.Printf("\nTemporary directory kept at: %s\n", tempDir)
	}

	return nil
}

// installMapping installs a single mapping entry
func installMapping(extractDir string, mapping Mapping) error {
	sourcePath := filepath.Join(extractDir, mapping.From)
	targetPath := expandPath(mapping.To)

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if sourceInfo.IsDir() {
		return installDirectory(sourcePath, targetPath)
	}

	return installFile(sourcePath, targetPath)
}

// installFile installs a single file
func installFile(sourcePath, targetPath string) error {
	// Handle gzipped files
	if filepath.Ext(sourcePath) == ".gz" {
		if err := extractGzipFile(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to extract gzip file: %w", err)
		}
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
		// Fix ownership if needed
		if err := fixOwnership(targetPath); err != nil {
			return fmt.Errorf("failed to fix ownership: %w", err)
		}
		fmt.Printf("  Installed %s (extracted from gzip)\n", targetPath)
		return nil
	}

	// Handle regular files
	if err := copyFile(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Make binary files executable
	if isBinary(targetPath) {
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
	}

	// Fix ownership if needed
	if err := fixOwnership(targetPath); err != nil {
		return fmt.Errorf("failed to fix ownership: %w", err)
	}

	fmt.Printf("  Installed %s\n", targetPath)
	return nil
}

// installDirectory installs a directory
func installDirectory(sourcePath, targetPath string) error {
	if err := copyDirectory(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to copy directory: %w", err)
	}

	// Fix ownership recursively if needed
	if err := fixOwnershipRecursive(targetPath); err != nil {
		return fmt.Errorf("failed to fix ownership: %w", err)
	}

	fmt.Printf("  Installed %s (directory)\n", targetPath)
	return nil
}

// copyFile copies a single file from source to destination
func copyFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// Fix ownership of parent directories if they were just created
	if err := fixOwnershipPath(dstDir); err != nil {
		return err
	}

	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy contents
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Fix ownership of the destination directory itself if it's in home
	if err := fixOwnership(dst); err != nil {
		return err
	}

	// Walk through source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return err
			}
			// Fix ownership immediately after creating
			return fixOwnership(dstPath)
		}

		// Copy file
		if err := copyFile(path, dstPath); err != nil {
			return err
		}
		// Fix ownership of copied file
		return fixOwnership(dstPath)
	})
}

// extractGzipFile extracts a gzip compressed file
func extractGzipFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open gzip file
	gzFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer gzFile.Close()

	// Create gzip reader
	gz, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}
	defer gz.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy uncompressed content
	_, err = io.Copy(dstFile, gz)
	return err
}

// isBinary checks if the file path indicates it's a binary executable
func isBinary(path string) bool {
	// Check if file is in /usr/local/bin
	return filepath.Dir(path) == "/usr/local/bin"
}

// fixOwnership fixes file ownership when running with sudo
func fixOwnership(path string) error {
	// Only fix ownership when running with sudo
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" {
		return nil
	}

	// Only fix ownership for files in home directory
	if !isInHomeDirectory(path) {
		return nil
	}

	// Get the original user's UID and GID
	u, err := user.Lookup(sudoUser)
	if err != nil {
		return fmt.Errorf("failed to lookup user %s: %w", sudoUser, err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("failed to parse UID: %w", err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("failed to parse GID: %w", err)
	}

	// Change ownership
	return os.Chown(path, uid, gid)
}

// fixOwnershipRecursive fixes ownership recursively for directories
func fixOwnershipRecursive(path string) error {
	// Only fix ownership when running with sudo
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" {
		return nil
	}

	// Only fix ownership for directories in home directory
	if !isInHomeDirectory(path) {
		return nil
	}

	// Get the original user's UID and GID
	u, err := user.Lookup(sudoUser)
	if err != nil {
		return fmt.Errorf("failed to lookup user %s: %w", sudoUser, err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("failed to parse UID: %w", err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("failed to parse GID: %w", err)
	}

	// Walk through directory and change ownership
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(p, uid, gid)
	})
}

// fixOwnershipPath fixes ownership for a path and all parent directories up to home
func fixOwnershipPath(path string) error {
	// Only fix ownership when running with sudo
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" {
		return nil
	}

	homeDir := getRealHomeDir()
	if homeDir == "" {
		return nil
	}

	// Get the original user's UID and GID
	u, err := user.Lookup(sudoUser)
	if err != nil {
		return fmt.Errorf("failed to lookup user %s: %w", sudoUser, err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("failed to parse UID: %w", err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("failed to parse GID: %w", err)
	}

	// Fix ownership of the path and parent directories up to home
	current := path
	for {
		if isInHomeDirectory(current) {
			if err := os.Chown(current, uid, gid); err != nil {
				// Ignore errors for directories we don't own
				if !os.IsPermission(err) {
					return err
				}
			}
		}

		// Stop at home directory
		if current == homeDir {
			break
		}

		// Move to parent
		parent := filepath.Dir(current)
		if parent == current {
			break // Reached root
		}
		current = parent
	}

	return nil
}

// isInHomeDirectory checks if a path is within any user's home directory
func isInHomeDirectory(path string) bool {
	homeDir := getRealHomeDir()
	if homeDir == "" {
		return false
	}

	// Clean paths for comparison
	cleanPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}

	cleanHome, err := filepath.Abs(filepath.Clean(homeDir))
	if err != nil {
		return false
	}

	// Check if path is within home directory
	return strings.HasPrefix(cleanPath, cleanHome+string(filepath.Separator)) || cleanPath == cleanHome
}
