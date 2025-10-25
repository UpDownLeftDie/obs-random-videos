package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/UpDownLeftDie/obs-random-videos/v2/internal/config"
	"github.com/UpDownLeftDie/obs-random-videos/v2/internal/media"
	"github.com/UpDownLeftDie/obs-random-videos/v2/internal/ui"
)

//go:embed web/templates/template.gohtml
var templateHTML string

//go:embed web/js/main.js
var mainScript string

//go:embed web/js/body.js
var bodyScript string

var version = "DEV"

func main() {
	// Parse configuration from command line flags
	cfg := config.ParseFlags(version)

	fmt.Printf("OBS Random Video: %s\n\n", cfg.Version)

	// Get the directory to scan
	var mainDir string
	var err error

	if cfg.Directory == "." {
		// Use executable directory
		mainDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatalf("Failed to get current directory path: %v", err)
		}
	} else {
		// Use specified directory
		mainDir, err = filepath.Abs(cfg.Directory)
		if err != nil {
			log.Fatalf("Failed to resolve directory path: %v", err)
		}
	}

	if cfg.Verbose {
		fmt.Printf("Scanning directory: %s\n", mainDir)
	}

	// Get supported file types
	fileTypes := media.DefaultFileTypes()

	// Get media files (use concurrent scanning if configured with multiple workers)
	var mediaFiles []string
	if cfg.ConcurrentScans > 1 {
		mediaFiles = media.GetMediaFilesConcurrent(mainDir, fileTypes, cfg.Verbose, cfg.ConcurrentScans)
	} else {
		mediaFiles = media.GetMediaFiles(mainDir, fileTypes, cfg.Verbose)
	}

	if len(mediaFiles) < 1 {
		fmt.Printf("No media files found in: %s", mainDir)
		fmt.Print("\n\nPress enter to exit...")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		return
	}

	// Auto-sanitize file names if requested
	if cfg.AutoSanitize {
		fmt.Println("\nSanitizing file names...")
		renamedCount, errors := media.SanitizeMediaFiles(mediaFiles)
		if len(errors) > 0 {
			fmt.Printf("\n⚠️  Encountered %d errors during sanitization:\n", len(errors))
			for _, err := range errors {
				fmt.Printf("   • %v\n", err)
			}
		}
		if renamedCount > 0 {
			fmt.Printf("\n✓ Successfully renamed %d files\n", renamedCount)
			// Re-scan to get updated file names
			if cfg.ConcurrentScans > 1 {
				mediaFiles = media.GetMediaFilesConcurrent(mainDir, fileTypes, false, cfg.ConcurrentScans)
			} else {
				mediaFiles = media.GetMediaFiles(mainDir, fileTypes, false)
			}
		}
	}

	// Validate all file paths
	for _, file := range mediaFiles {
		if err := media.ValidatePath(mainDir, file); err != nil {
			log.Fatalf("Security validation failed: %v", err)
		}
	}

	// Ask user for configuration
	answers, err := askQuestions(mediaFiles, cfg.Verbose)
	if err != nil {
		log.Fatalf("Something went wrong getting user input: %v", err)
	}

	// Set the version in answers
	answers.Version = cfg.Version

	// Remove transition video from media files if needed
	if answers.TransitionVideo != "" {
		answers.MediaFiles = media.RemoveTransitionVideo(answers.TransitionVideo, answers.MediaFiles)
	}

	// Create scripts object
	scripts := ui.Scripts{
		MainScript: mainScript,
		BodyScript: bodyScript,
	}

	// Generate HTML
	outputHTML, err := ui.GenerateHTML(templateHTML, scripts, answers)
	if err != nil {
		log.Fatalf("Failed generating HTML: %v", err)
	}

	// Write HTML to file
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	outputHTMLFilePath := filepath.Join(exPath, cfg.OutputFileName)
	outputHTMLFile, err := os.Create(outputHTMLFilePath)
	if err != nil {
		log.Fatalf("Failed create output file: %v", err)
	}
	defer outputHTMLFile.Close()
	outputHTMLFile.WriteString(outputHTML)

	fmt.Printf("\nCreated %s successfully!\n", cfg.OutputFileName)

	if cfg.Verbose {
		fmt.Printf("Output location: %s\n", outputHTMLFilePath)
	}

	os.Exit(0)
}

// askQuestions prompts the user for configuration options using an interactive UI.
//
// Parameters:
//   - mediaFiles: List of media file paths found in the scan
//   - verbose: Whether to output verbose information
//
// Returns:
//   - UserAnswers: The user's configuration choices
//   - error: Any error encountered during the questioning process
func askQuestions(mediaFiles []string, verbose bool) (ui.UserAnswers, error) {
	answers := ui.UserAnswers{
		MediaFiles:          mediaFiles,
		PlayOnlyOne:         false,
		LoopFirstVideo:      false,
		HaveTransitionVideo: false,
		TransitionVideo:     "",
		HashKey:             "",
		Version:             "",
	}

	// Display a few sample files to verify
	if len(mediaFiles) > 0 && verbose {
		fmt.Printf("\nFound %d media files\n", len(mediaFiles))
		fmt.Println("Sample media files:")
		maxSamples := 5
		if len(mediaFiles) < maxSamples {
			maxSamples = len(mediaFiles)
		}
		for i := 0; i < maxSamples; i++ {
			fmt.Printf("  %d: %s\n", i+1, mediaFiles[i])
		}
		fmt.Println()
	}

	// Ask questions using Bubble Tea
	answers.PlayOnlyOne = ui.ShowQuestion("Do you only want to play one video? (The first random video will play once and then stop)", false)

	if !answers.PlayOnlyOne {
		answers.LoopFirstVideo = ui.ShowQuestion("Do you want to loop the first video?", false)
		answers.HaveTransitionVideo = ui.ShowQuestion("Do you have a transition video? (This video plays after every other video)", false)

		if answers.HaveTransitionVideo {
			// Make sure we have media files to select from
			if len(mediaFiles) == 0 {
				fmt.Println("No media files available for selection")
				answers.HaveTransitionVideo = false
			} else {
				var err error
				answers.TransitionVideo, err = ui.SelectTransitionVideo(mediaFiles)
				if err != nil {
					return answers, err
				}

				if answers.TransitionVideo == "" {
					answers.HaveTransitionVideo = false
				}
			}
		}
	}

	// Use the hash function from ui package (removing duplicate)
	answers.HashKey = ui.CreateHashFromUserAnswers(answers)
	return answers, nil
}
