package config

import (
	"testing"
)

func TestDefault(t *testing.T) {
	version := "1.0.0"
	cfg := Default(version)

	if cfg.Version != version {
		t.Errorf("Version = %v, want %v", cfg.Version, version)
	}

	if cfg.OutputFileName != "obs-random-videos.html" {
		t.Errorf("OutputFileName = %v, want 'obs-random-videos.html'", cfg.OutputFileName)
	}

	if cfg.ConcurrentScans != 4 {
		t.Errorf("ConcurrentScans = %v, want 4", cfg.ConcurrentScans)
	}

	if cfg.Recursive != true {
		t.Errorf("Recursive = %v, want true", cfg.Recursive)
	}

	if cfg.Verbose != false {
		t.Errorf("Verbose = %v, want false", cfg.Verbose)
	}

	if cfg.AutoSanitize != false {
		t.Errorf("AutoSanitize = %v, want false", cfg.AutoSanitize)
	}
}

func TestConfigFields(t *testing.T) {
	cfg := &Config{
		OutputFileName:  "test.html",
		Version:         "2.0.0",
		MaxFileSize:     1024,
		ConcurrentScans: 8,
		Directory:       "/test/dir",
		Recursive:       false,
		Verbose:         true,
		AutoSanitize:    true,
	}

	if cfg.OutputFileName != "test.html" {
		t.Errorf("OutputFileName = %v, want 'test.html'", cfg.OutputFileName)
	}

	if cfg.MaxFileSize != 1024 {
		t.Errorf("MaxFileSize = %v, want 1024", cfg.MaxFileSize)
	}

	if cfg.ConcurrentScans != 8 {
		t.Errorf("ConcurrentScans = %v, want 8", cfg.ConcurrentScans)
	}
}
