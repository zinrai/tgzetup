package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	var installURL string
	var uninstall bool
	var keepTemp bool
	var showVersion bool
	var mappingFile string

	flag.StringVar(&installURL, "install", "", "URL of tar.gz archive to install")
	flag.BoolVar(&uninstall, "uninstall", false, "Uninstall")
	flag.BoolVar(&keepTemp, "keep-temp", false, "Keep temporary directory after installation")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.StringVar(&mappingFile, "mapping", "", "Path to mapping configuration file (required)")
	flag.Parse()

	if showVersion {
		fmt.Printf("tgzetup %s\n", version)
		os.Exit(0)
	}

	// Check mutually exclusive options
	if installURL != "" && uninstall {
		fmt.Fprintf(os.Stderr, "Error: -install and -uninstall cannot be used together\n")
		os.Exit(1)
	}

	// Require at least one action
	if installURL == "" && !uninstall {
		flag.Usage()
		os.Exit(1)
	}

	// Require mapping file
	if mappingFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -mapping option is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load mapping configuration
	config, err := LoadMapping(mappingFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading mapping file: %v\n", err)
		os.Exit(1)
	}

	// Execute the requested action
	var actionErr error
	if uninstall {
		actionErr = Uninstall(config)
		if actionErr == nil {
			fmt.Println("Uninstallation completed.")
		}
	} else {
		actionErr = Install(installURL, config, keepTemp)
		if actionErr == nil {
			fmt.Println("Installation completed successfully.")
		}
	}

	if actionErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", actionErr)
		os.Exit(1)
	}
}
