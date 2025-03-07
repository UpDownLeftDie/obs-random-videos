package ui

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"text/template"
)

// UserAnswers holds config answers from command line prompt ui
type UserAnswers struct {
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

// GenerateHTML generates the HTML file from the template and user answers
func GenerateHTML(templateHTML string, scripts Scripts, answers UserAnswers) (string, error) {
	// Add version header
	templateHTML = "<!--\nOBS Random Videos\nAUTO GENERATED FILE\nDON'T TOUCH\n-->\n" + templateHTML

	// First pass: inject scripts
	var outputHTML bytes.Buffer
	t := template.Must(template.New("HTML").Parse(templateHTML))
	err := t.Execute(&outputHTML, scripts)
	if err != nil {
		return "", fmt.Errorf("failed compiling template: %w", err)
	}

	// Second pass: inject user answers
	t = template.Must(template.New("HTML").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(outputHTML.String()))
	outputHTML.Reset()
	err = t.Execute(&outputHTML, answers)
	if err != nil {
		return "", fmt.Errorf("failed compiling template final: %w", err)
	}

	return outputHTML.String(), nil
}

// CreateHashFromUserAnswers creates a unique hash based on user settings
func CreateHashFromUserAnswers(answers UserAnswers) string {
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
