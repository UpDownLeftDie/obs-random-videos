// Package config provides configuration management for the OBS Random Videos application.
// It handles parsing command-line flags and maintaining application settings.
package config

import (
	"flag"
	"os"
)

// Config holds the application configuration
type Config struct {
	// OutputFileName is the name of the generated HTML file
	OutputFileName string
	// Version is the application version
	Version string
	// MaxFileSize is the maximum file size to consider (in bytes, 0 = no limit)
	MaxFileSize int64
	// ConcurrentScans is the number of concurrent goroutines for directory scanning
	ConcurrentScans int
	// Directory is the directory to scan for media files
	Directory string
	// Recursive determines if subdirectories should be scanned
	Recursive bool
	// Verbose enables verbose output
	Verbose bool
	// AutoSanitize determines if file names should be automatically sanitized
	AutoSanitize bool
}

// Default returns a Config with default values
func Default(version string) *Config {
	return &Config{
		OutputFileName:  "obs-random-videos.html",
		Version:         version,
		MaxFileSize:     0, // No limit by default
		ConcurrentScans: 4,
		Directory:       ".",
		Recursive:       true,
		Verbose:         false,
		AutoSanitize:    false,
	}
}

// ParseFlags parses command line flags and returns a Config
func ParseFlags(version string) *Config {
	cfg := Default(version)

	flag.StringVar(&cfg.Directory, "dir", cfg.Directory, "Directory to scan for media files")
	flag.StringVar(&cfg.OutputFileName, "output", cfg.OutputFileName, "Output HTML filename")
	flag.BoolVar(&cfg.Recursive, "recursive", cfg.Recursive, "Scan subdirectories")
	flag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "Verbose output")
	flag.BoolVar(&cfg.AutoSanitize, "sanitize", cfg.AutoSanitize, "Automatically sanitize problematic file names")
	flag.Int64Var(&cfg.MaxFileSize, "max-size", cfg.MaxFileSize, "Maximum file size in bytes (0 = no limit)")
	flag.IntVar(&cfg.ConcurrentScans, "concurrent", cfg.ConcurrentScans, "Number of concurrent file scanners")

	// Check for DEBUG environment variable
	if os.Getenv("DEBUG") != "" {
		cfg.Verbose = true
	}

	flag.Parse()

	return cfg
}
