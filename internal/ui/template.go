package ui

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	htmltemplate "html/template"
	"net/url"
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

	// First pass: inject scripts using text/template to avoid JavaScript parsing issues
	var outputHTML bytes.Buffer
	t := template.Must(template.New("HTML").Parse(templateHTML))
	err := t.Execute(&outputHTML, scripts)
	if err != nil {
		return "", fmt.Errorf("failed compiling template: %w", err)
	}

	// Second pass: inject user answers using html/template for proper escaping
	funcMap := htmltemplate.FuncMap{
		"StringsJoin": func(items []string, sep string) htmltemplate.JS {
			// Properly escape each string for JavaScript
			escapedItems := make([]string, len(items))
			for i, item := range items {
				// Split path and encode each part properly for URIs
				parts := strings.Split(item, "/")
				encodedParts := make([]string, len(parts))
				for j, part := range parts {
					encodedParts[j] = url.PathEscape(part)
				}
				encodedPath := strings.Join(encodedParts, "/")
				escapedItems[i] = "\"" + htmltemplate.JSEscapeString(encodedPath) + "\""
			}
			return htmltemplate.JS(strings.Join(escapedItems, sep))
		},
		"JSEscape": func(s string) htmltemplate.JS {
			// Pre-encode path for URI compatibility, then escape for JavaScript
			parts := strings.Split(s, "/")
			encodedParts := make([]string, len(parts))
			for i, part := range parts {
				encodedParts[i] = url.PathEscape(part)
			}
			encodedPath := strings.Join(encodedParts, "/")
			return htmltemplate.JS("\"" + htmltemplate.JSEscapeString(encodedPath) + "\"")
		},
	}
	htmlt := htmltemplate.Must(htmltemplate.New("HTML").Funcs(funcMap).Parse(outputHTML.String()))
	outputHTML.Reset()
	err = htmlt.Execute(&outputHTML, answers)
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
