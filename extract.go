package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTarGz extracts a tar.gz archive to the specified directory
func ExtractTarGz(archivePath string, destDir string) error {
	fmt.Printf("Extracting archive to %s...\n", destDir)

	// Open the archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct the full path
		target := filepath.Join(destDir, header.Name)

		// Security check: ensure the target path is within destDir
		if !isPathWithinDir(target, destDir) {
			return fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		// Extract based on type
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				// Continue on error - not fatal
				continue
			}
		case tar.TypeReg:
			// Extract regular file
			if err := extractRegularFile(tr, header, target); err != nil {
				// Continue on error - not fatal
				continue
			}
		default:
			// Skip other types silently (symlinks, hard links, etc.)
			continue
		}
	}

	fmt.Println("Extraction completed")
	return nil
}

// extractRegularFile extracts a regular file from tar
func extractRegularFile(tr *tar.Reader, header *tar.Header, target string) error {
	// Create directory for the file
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	// Create the file
	file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy file contents
	_, err = io.Copy(file, tr)
	return err
}

// isPathWithinDir checks if a path is within a directory (security check)
func isPathWithinDir(path, dir string) bool {
	cleanPath := filepath.Clean(path)
	cleanDir := filepath.Clean(dir)

	// Use HasPrefix with proper path separator handling
	return strings.HasPrefix(cleanPath, cleanDir+string(filepath.Separator)) || cleanPath == cleanDir
}
