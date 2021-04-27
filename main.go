package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/sparkdemcisin81/promptui" // using this instead of manifoldco because of weird race conditions
)

//go:embed template.html
var templateHtml string

type UserAnswers struct {
	VideoFolder         string
	MediaFiles          []string
	PlayOnlyOne         bool
	LoopFirstVideo      bool
	HaveTransitionVideo bool
	TransitionVideo     string
}

func main() {
	outputHtmlName := "obs-random-videos.html"
	audioFileExts := []string{".mp3", ".ogg", ".aac"}
	videoFileExts := []string{".mp4", ".webm", ".mpeg4"}
	mediaFileExts := append(audioFileExts, videoFileExts...)

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}
	mediaFiles := filterFiles(files, mediaFileExts)
	if len(mediaFiles) < 1 {
		fmt.Printf("No media files found!")
		return
	}

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
			answers.TransitionVideo, err = GetTransitionVideo(mediaFiles)

			if err != nil {
				return
			}
		}
	}

	fields := reflect.TypeOf(answers)
	values := reflect.ValueOf(answers)

	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		value := values.Field(i)
		finalValue := ""
		switch value.Kind() {
		case reflect.String:
			finalValue = value.String()
			if field.Name == "VideoFolder" {
				finalValue = strings.Replace(finalValue, "\\", "\\\\", -1) + "\\\\"
			}
		case reflect.Bool:
			finalValue = strconv.FormatBool(value.Bool())
		case reflect.Slice:
			j, _ := json.Marshal(value.Interface())
			finalValue = string(j)
		default:
			finalValue = ""
		}

		templateHtml = strings.Replace(templateHtml, "$"+field.Name+"$", finalValue, 1)
	}
	templateHtml = "<!--\nAUTO GENERATED FILE\nDON'T TOUCH\n-->\n" + templateHtml
	os.WriteFile(outputHtmlName, []byte(templateHtml), 0644)
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

	if err != nil {
		fmt.Printf("Prompt failed getting transition video: %v\n", err)
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
		return ShowQuestion(message, defaultValueFlag)
	}

	result = strings.ToLower(result)

	return result == "y" || result == "yes"
}
