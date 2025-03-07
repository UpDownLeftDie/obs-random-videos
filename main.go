package main

import (
	"bufio"
	_ "embed"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"crypto/md5"

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
var outputHTMLName = "obs-random-videos.html"

func main() {
	fmt.Printf("OBS Random Video: %s\n\n", version)

	// Get the current directory
	mainDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("Failed to get current directory path: %v", err)
	}

	// Get supported file types
	fileTypes := media.DefaultFileTypes()

	// Get media files
	mediaFiles := media.GetMediaFiles(mainDir, fileTypes)
	if len(mediaFiles) < 1 {
		fmt.Printf("No media files found in: %s", mainDir)
		fmt.Print("\n\nPress enter to exit...")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		return
	}

	// Ask user for configuration
	answers, err := askQuestions(mediaFiles)
	if err != nil {
		log.Fatalf("Something went wrong getting user input: %v", err)
	}

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
	outputHTMLFilePath := filepath.Join(exPath, outputHTMLName)
	outputHTMLFile, err := os.Create(outputHTMLFilePath)
	if err != nil {
		log.Fatalf("Failed create output file: %v", err)
	}
	outputHTMLFile.WriteString(outputHTML)
	outputHTMLFile.Close()

	fmt.Printf("\nCreated %s successfully!\n", outputHTMLName)
	os.Exit(0)
}

func createHashFromUserAnswers(answers ui.UserAnswers) string {
	s := fmt.Sprintf(
		"%v%v%v%s%s",
		answers.PlayOnlyOne,
		answers.LoopFirstVideo,
		answers.HaveTransitionVideo,
		answers.TransitionVideo,
		strings.Join(answers.MediaFiles[:], ""))

	// Not using this hash for anything sensitive
	hasher := md5.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

func askQuestions(mediaFiles []string) (ui.UserAnswers, error) {
	answers := ui.UserAnswers{
		MediaFiles:          mediaFiles,
		PlayOnlyOne:         false,
		LoopFirstVideo:      false,
		HaveTransitionVideo: false,
		TransitionVideo:     "",
		HashKey:             "",
	}

	// Display a few sample files to verify
	if len(mediaFiles) > 0 && os.Getenv("DEBUG") != "" {
		fmt.Printf("Found %d media files\n", len(mediaFiles))
		fmt.Println("Sample media files:")
		maxSamples := 5
		if len(mediaFiles) < maxSamples {
			maxSamples = len(mediaFiles)
		}
		for i := 0; i < maxSamples; i++ {
			fmt.Printf("  %d: %s\n", i+1, mediaFiles[i])
		}
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

	answers.HashKey = createHashFromUserAnswers(answers)
	return answers, nil
}
