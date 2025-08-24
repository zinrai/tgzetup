package main

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir := getRealHomeDir()
		if homeDir == "" {
			// Fallback to current user's home
			homeDir, _ = os.UserHomeDir()
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// getRealHomeDir returns the actual user's home directory, even when running with sudo
func getRealHomeDir() string {
	// Check if running under sudo
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		// Get the original user's information
		u, err := user.Lookup(sudoUser)
		if err == nil {
			return u.HomeDir
		}
	}

	// Not running under sudo, return current user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return homeDir
}
