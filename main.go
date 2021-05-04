package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/sparkdemcisin81/promptui" // using this instead of manifoldco because of weird bugs 🤷‍♂️
)

//go:embed template.gohtml
var templateHtml string

type UserAnswers struct {
	VideoFolder         string
	MediaFiles          []string
	PlayOnlyOne         bool
	LoopFirstVideo      bool
	HaveTransitionVideo bool
	TransitionVideo     string
}

var outputHtmlName = "obs-random-videos.html"
var audioFileExts = []string{".mp3", ".ogg", ".aac"}
var videoFileExts = []string{".mp4", ".webm", ".mpeg4", ".m4v"}
var mediaFileExts = append(audioFileExts, videoFileExts...)
var promptDelay = 100 * time.Millisecond // helps with race conditionsin promptui

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory path: %v", err)
	}
	separator := string(os.PathSeparator)
	currentDir += separator
	if runtime.GOOS == "windows" {
		separatorEscaped := strings.Repeat(separator, 2)
		currentDir = strings.Replace(currentDir, separator, separatorEscaped, -1)
	}
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatalf("Failed to read current directory: %v", err)
	}
	mediaFiles := filterFiles(files, mediaFileExts)
	if len(mediaFiles) < 1 {
		log.Fatal("No media files found!")
		return
	}
	outputHtml, err := os.Create(outputHtmlName)
	if err != nil {
		log.Fatalf("Failed create output file: %v", err)
	}

	answers, err := askQuestions(currentDir, mediaFiles)
	if err != nil {
		outputHtml.Close()
		os.Remove(outputHtmlName)
		log.Fatalf("Something went wrong getting user input: %v", err)
	}
	if answers.TransitionVideo != "" {
		answers.MediaFiles = removeTransitionVideo(answers.TransitionVideo, answers.MediaFiles)
	}

	templateHtml = "<!--\nAUTO GENERATED FILE\nDON'T TOUCH\n-->\n" + templateHtml
	t := template.Must(template.New("html").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(templateHtml))
	err = t.Execute(outputHtml, answers)
	if err != nil {
		outputHtml.Close()
		os.Remove(outputHtmlName)
		log.Fatalf("Failed compiling template: %v", err)
	}
	outputHtml.Close()
}

func askQuestions(currentDir string, mediaFiles []string) (UserAnswers, error) {
	answers := UserAnswers{
		VideoFolder:         currentDir,
		MediaFiles:          mediaFiles,
		PlayOnlyOne:         false,
		LoopFirstVideo:      false,
		HaveTransitionVideo: false,
		TransitionVideo:     "",
	}

	answers.PlayOnlyOne = ShowQuestion("Do you only want to play one video? (The first random video will play once and then stop)", false)
	if !answers.PlayOnlyOne {
		answers.LoopFirstVideo = ShowQuestion("Do you want to loop the first video?", false)
		answers.HaveTransitionVideo = ShowQuestion("Do have a transition video? (This video plays after every other video)", false)
		if answers.HaveTransitionVideo {
			err := errors.New("")
			answers.TransitionVideo, err = GetTransitionVideo(mediaFiles)
			if err != nil {
				fmt.Printf("Prompt failed getting transition video: %v", err)
				return answers, err
			}
		}
	}
	return answers, nil
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

func filterFiles(files []fs.FileInfo, fileExts []string) []string {
	filteredFiles := []string{}
	for _, f := range files {
		for _, ext := range fileExts {
			if strings.HasSuffix(f.Name(), ext) {
				filteredFiles = append(filteredFiles, f.Name())
			}
		}
	}
	return filteredFiles
}

func GetTransitionVideo(mediaFiles []string) (string, error) {
	items := []string{}
	for _, f := range mediaFiles {
		items = append(items, f)
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
		return result, nil
	}
	return "", nil
}

func ShowQuestion(message string, defaultValueFlag bool) bool {
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
		return errors.New(fmt.Sprintf("Number should be one of the values %v", allowedValues))
	}

	prompt := promptui.Prompt{
		Label:    message,
		Validate: validate,
		Default:  defaultValue,
	}

	result, err := prompt.Run()
	if err != nil {
		time.Sleep(promptDelay)
		return ShowQuestion(message, defaultValueFlag)
	}

	result = strings.ToLower(result)
	time.Sleep(promptDelay)
	return result == "y" || result == "yes"
}
