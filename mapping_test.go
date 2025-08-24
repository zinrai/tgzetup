package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMapping(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		check   func(t *testing.T, config *Config)
	}{
		{
			name: "valid mapping",
			yaml: `mappings:
  - from: "bin/limactl"
    to: "/usr/local/bin/limactl"
  - from: "share/lima/templates"
    to: "~/.lima/_templates"`,
			wantErr: false,
			check: func(t *testing.T, config *Config) {
				if len(config.Mappings) != 2 {
					t.Errorf("expected 2 mappings, got %d", len(config.Mappings))
				}
				if config.Mappings[0].From != "bin/limactl" {
					t.Errorf("expected first mapping from 'bin/limactl', got %s", config.Mappings[0].From)
				}
				if config.Mappings[0].To != "/usr/local/bin/limactl" {
					t.Errorf("expected first mapping to '/usr/local/bin/limactl', got %s", config.Mappings[0].To)
				}
			},
		},
		{
			name:    "empty mappings",
			yaml:    `mappings: []`,
			wantErr: true,
		},
		{
			name: "missing from field",
			yaml: `mappings:
  - to: "/usr/local/bin/limactl"`,
			wantErr: true,
		},
		{
			name: "missing to field",
			yaml: `mappings:
  - from: "bin/limactl"`,
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			yaml:    `mappings: [invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary YAML file
			tmpDir := t.TempDir()
			yamlPath := filepath.Join(tmpDir, "test-mapping.yaml")
			err := os.WriteFile(yamlPath, []byte(tt.yaml), 0644)
			if err != nil {
				t.Fatalf("failed to write test yaml: %v", err)
			}

			// Test LoadMapping
			config, err := LoadMapping(yamlPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadMapping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Run additional checks if provided
			if !tt.wantErr && tt.check != nil {
				tt.check(t, config)
			}
		})
	}
}

func TestLoadMapping_FileNotFound(t *testing.T) {
	_, err := LoadMapping("/non/existent/file.yaml")
	if err == nil {
		t.Error("LoadMapping() expected error for non-existent file, got nil")
	}
}
