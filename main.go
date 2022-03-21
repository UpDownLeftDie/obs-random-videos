package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/sparkdemcisin81/promptui" // using this instead of manifoldco because of weird bugs 🤷‍♂️
)

//go:embed template.gohtml
var templateHTML string

//go:embed js/main.js
var mainScript string

//go:embed js/body.js
var bodyScript string

// UserAnswers holds config answers from comand line prompt ui
type UserAnswers struct {
	MediaFolder         string
	MediaFiles          []string
	PlayOnlyOne         bool
	LoopFirstVideo      bool
	HaveTransitionVideo bool
	TransitionVideo     string
	HashKey             string
}

// Scripts stores javascript scripts that are later injected into templateHTML
type Scripts struct {
	MainScript string
	BodyScript string
}

var scripts = Scripts{
	MainScript: mainScript,
	BodyScript: bodyScript,
}

var (
	outputHTMLName = "obs-random-videos.html"
	audioFileExts  = []string{".mp3", ".ogg", ".aac"}
	videoFileExts  = []string{".mp4", ".webm", ".mpeg4", ".m4v", ".mov"}
	// imagesFileExts = []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	// mediaFileExts  = append(append(audioFileExts, videoFileExts...), imagesFileExts...)
	mediaFileExts = append(audioFileExts, videoFileExts...)
	promptDelay   = 100 * time.Millisecond // helps with race conditionsin promptui
)

func main() {
	mainDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("Failed to get current directory path: %v", err)
	}
	mediaFiles := getMediaFiles(mainDir)
	if len(mediaFiles) < 1 {
		fmt.Printf("No media files found in: %s", mainDir)
		fmt.Print("\n\nPress enter to exit...")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		return
	}

	answers, err := askQuestions(mediaFiles, mainDir)
	if err != nil {
		log.Fatalf("Something went wrong getting user input: %v", err)
	}
	if answers.TransitionVideo != "" {
		answers.MediaFiles = removeTransitionVideo(answers.TransitionVideo, answers.MediaFiles)
	}

	templateHTML = "<!--\nAUTO GENERATED FILE\nDON'T TOUCH\n-->\n" + templateHTML
	var outputHTML bytes.Buffer
	t := template.Must(template.New("HTML").Parse(templateHTML))
	err = t.Execute(&outputHTML, scripts)
	if err != nil {
		log.Fatalf("Failed compiling template: %v", err)
	}
	t = template.Must(template.New("HTML").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(outputHTML.String()))
	outputHTML.Reset()
	err = t.Execute(&outputHTML, answers)
	if err != nil {
		log.Fatalf("Failed compiling template final: %v", err)
	}
	outputHTMLFile, err := os.Create(outputHTMLName)
	if err != nil {
		log.Fatalf("Failed create output file: %v", err)
	}
	outputHTMLFile.WriteString(outputHTML.String())
	outputHTMLFile.Close()
	os.Exit(0)
}

func getMediaFiles(currentDir string) []string {
	mediaFiles := []string{}
	filepath.WalkDir(currentDir, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !file.IsDir() {
			if isValidFileType(file) {
				// TODO: move this fix to end of program before injection
				// 	Then clean up getTransitionVideo and show the subpath in the file selector
				fixedFilePath := fixFilePath(path)
				mediaFiles = append(mediaFiles, fixedFilePath)
			}
		}
		return nil
	})
	return mediaFiles
}

func fixFilePath(filePath string) string {
	separator := string(os.PathSeparator)
	if runtime.GOOS == "windows" {
		separatorEscaped := strings.Repeat(separator, 2)
		return strings.Replace(filePath, separator, separatorEscaped, -1)
	}
	return filePath
}

func askQuestions(mediaFiles []string, mainDir string) (UserAnswers, error) {
	answers := UserAnswers{
		MediaFiles:          mediaFiles,
		PlayOnlyOne:         false,
		LoopFirstVideo:      false,
		HaveTransitionVideo: false,
		TransitionVideo:     "",
		HashKey:             "",
	}

	answers.PlayOnlyOne = showQuestion("Do you only want to play one video? (The first random video will play once and then stop)", false)
	if !answers.PlayOnlyOne {
		answers.LoopFirstVideo = showQuestion("Do you want to loop the first video?", false)
		answers.HaveTransitionVideo = showQuestion("Do have a transition video? (This video plays after every other video)", false)
		if answers.HaveTransitionVideo {
			err := errors.New("")
			answers.TransitionVideo, err = getTransitionVideo(mediaFiles, mainDir)
			if err != nil {
				return answers, err
			}
		}
	}
	answers.HashKey = createHashFromUserAnswers(answers)
	return answers, nil
}

func createHashFromUserAnswers(answers UserAnswers) string {
	s := fmt.Sprintf(
		"%v%v%v%s%s",
		answers.PlayOnlyOne,
		answers.LoopFirstVideo,
		answers.HaveTransitionVideo,
		answers.TransitionVideo,
		strings.Join(answers.MediaFiles[:], ""))

	hasher := md5.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

func removeTransitionVideo(transitionVideo string, mediaFiles []string) []string {
	files := mediaFiles
	for i, file := range mediaFiles {
		if file == transitionVideo {
			files[i] = files[len(files)-1]
			return files[:len(files)-1]
		}
	}
	return files
}

func isValidFileType(file fs.DirEntry) bool {
	fileExts := mediaFileExts
	for _, ext := range fileExts {
		if strings.HasSuffix(file.Name(), ext) {
			return true
		}
	}
	return false
}

func getTransitionVideo(mediaFiles []string, mainDir string) (string, error) {
	items := []string{}
	// Strip full filepath when displaying options to users
	for _, filePath := range mediaFiles {
		fileParts := strings.Split(filePath, string(os.PathSeparator))
		fmt.Printf("%v", fileParts)
		fileName := fileParts[len(fileParts)-1]
		items = append(items, fileName)
	}
	items = append(items, "CANCEL")
	prompt := promptui.Select{
		Label: "Select your transition video",
		Items: items,
	}
	_, result, err := prompt.Run()
	time.Sleep(promptDelay)
	if err != nil {
		return "", err
	}
	if strings.ToLower(result) != "cancel" {
		// Need to match file name with full path again
		for _, filePath := range mediaFiles {
			if strings.Contains(filePath, result) {
				return filePath, nil
			}
		}
		log.Fatalf("Error occurred when trying to get transition video path for: %s", result)
	}
	return "", nil
}

func showQuestion(message string, defaultValueFlag bool) bool {
	defaultValue := "y"
	if !defaultValueFlag {
		defaultValue = "n"
	}
	allowedValues := [...]string{"y", "yes", "no", "n"}

	validate := func(input string) error {
		for _, value := range allowedValues {
			if strings.ToLower(input) == value {
				return nil
			}
		}
		return fmt.Errorf("number should be one of the values %v", allowedValues)
	}

	prompt := promptui.Prompt{
		Label:    message,
		Validate: validate,
		Default:  defaultValue,
	}

	result, err := prompt.Run()
	if err != nil {
		time.Sleep(promptDelay)
		return showQuestion(message, defaultValueFlag)
	}

	result = strings.ToLower(result)
	time.Sleep(promptDelay)
	return result == "y" || result == "yes"
}
