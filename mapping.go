package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Mapping represents a single file/directory mapping
type Mapping struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// Config represents the complete mapping configuration
type Config struct {
	Mappings []Mapping `yaml:"mappings"`
}

// LoadMapping loads and parses the mapping configuration file
func LoadMapping(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapping file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate that mappings exist
	if len(config.Mappings) == 0 {
		return nil, fmt.Errorf("no mappings defined in configuration")
	}

	// Validate each mapping
	for i, mapping := range config.Mappings {
		if mapping.From == "" {
			return nil, fmt.Errorf("mapping %d: 'from' field is empty", i)
		}
		if mapping.To == "" {
			return nil, fmt.Errorf("mapping %d: 'to' field is empty", i)
		}
	}

	return &config, nil
}
