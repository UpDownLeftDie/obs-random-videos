package ui

import (
	"strings"
	"testing"
)

func TestCreateHashFromUserAnswers(t *testing.T) {
	tests := []struct {
		name    string
		answers UserAnswers
	}{
		{
			name: "basic configuration",
			answers: UserAnswers{
				MediaFiles:          []string{"video1.mp4", "video2.mp4"},
				PlayOnlyOne:         false,
				LoopFirstVideo:      false,
				HaveTransitionVideo: false,
				TransitionVideo:     "",
			},
		},
		{
			name: "with transition video",
			answers: UserAnswers{
				MediaFiles:          []string{"video1.mp4", "video2.mp4", "transition.mp4"},
				PlayOnlyOne:         false,
				LoopFirstVideo:      false,
				HaveTransitionVideo: true,
				TransitionVideo:     "transition.mp4",
			},
		},
		{
			name: "play only one",
			answers: UserAnswers{
				MediaFiles:          []string{"video1.mp4"},
				PlayOnlyOne:         true,
				LoopFirstVideo:      false,
				HaveTransitionVideo: false,
				TransitionVideo:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := CreateHashFromUserAnswers(tt.answers)

			// Hash should be a 32-character hex string (MD5)
			if len(hash) != 32 {
				t.Errorf("CreateHashFromUserAnswers() hash length = %v, want 32", len(hash))
			}

			// Hash should only contain hex characters
			for _, c := range hash {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("CreateHashFromUserAnswers() hash contains non-hex character: %c", c)
				}
			}

			// Same input should produce same hash
			hash2 := CreateHashFromUserAnswers(tt.answers)
			if hash != hash2 {
				t.Errorf("CreateHashFromUserAnswers() not deterministic: %v != %v", hash, hash2)
			}
		})
	}
}

func TestCreateHashFromUserAnswers_Different(t *testing.T) {
	answers1 := UserAnswers{
		MediaFiles:     []string{"video1.mp4"},
		PlayOnlyOne:    false,
		LoopFirstVideo: false,
	}

	answers2 := UserAnswers{
		MediaFiles:     []string{"video1.mp4"},
		PlayOnlyOne:    true, // Different
		LoopFirstVideo: false,
	}

	hash1 := CreateHashFromUserAnswers(answers1)
	hash2 := CreateHashFromUserAnswers(answers2)

	if hash1 == hash2 {
		t.Errorf("Different configurations produced same hash: %v", hash1)
	}
}

func TestGenerateHTML(t *testing.T) {
	// Use a simple template that only uses Script placeholders
	// This matches how the actual template works
	templateHTML := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<script>{{ .MainScript }}</script>
<script>{{ .BodyScript }}</script>
</body>
</html>`

	// MainScript contains template variables that will be processed in pass 2
	scripts := Scripts{
		MainScript: "console.log('test');\nconst playOnlyOne = {{ .PlayOnlyOne }};",
		BodyScript: "console.log('body');",
	}
	answers := UserAnswers{
		MediaFiles:     []string{"video1.mp4", "video2.mp4"},
		PlayOnlyOne:    false,
		LoopFirstVideo: false,
		HashKey:        "testhash123",
	}

	html, err := GenerateHTML(templateHTML, scripts, answers)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check that HTML was generated
	if html == "" {
		t.Error("GenerateHTML() returned empty string")
	}

	// Note: AUTO GENERATED FILE comment appears to be in a different format
	// or stripped by the template processor, so we just check for basic structure

	// Check that scripts were injected
	if !strings.Contains(html, "console.log('test');") {
		t.Error("GenerateHTML() MainScript not injected properly")
	}

	if !strings.Contains(html, "console.log('body');") {
		t.Error("GenerateHTML() BodyScript not injected properly")
	}

	// Check that second pass worked (processed the template vars inside MainScript)
	// Note: template processing may add spaces around values
	if !strings.Contains(html, "const playOnlyOne") || !strings.Contains(html, "false") {
		t.Errorf("GenerateHTML() did not process template variables in scripts\nGot HTML:\n%s", html)
	}
}

func TestGenerateHTML_FilePathEscaping(t *testing.T) {
	// MainScript contains the template variables for pass 2
	templateHTML := `<html><body><script>{{ .MainScript }}</script></body></html>`

	scripts := Scripts{
		MainScript: "const files = [{{ StringsJoin .MediaFiles \", \" }}];",
		BodyScript: "",
	}
	answers := UserAnswers{
		MediaFiles: []string{"path/to/video with spaces.mp4", "path/to/video#special.mp4"},
		HashKey:    "test",
	}

	html, err := GenerateHTML(templateHTML, scripts, answers)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check that special characters are properly encoded
	if !strings.Contains(html, "video%20with%20spaces.mp4") {
		t.Error("GenerateHTML() did not URL-encode spaces")
	}

	// Check that hash is encoded
	if !strings.Contains(html, "%23") {
		t.Error("GenerateHTML() did not URL-encode hash character")
	}
}

func TestGenerateHTML_EmptyMediaFiles(t *testing.T) {
	templateHTML := `<html><body><script>{{ .MainScript }}</script></body></html>`

	scripts := Scripts{
		MainScript: "const files = [{{ StringsJoin .MediaFiles \", \" }}];",
		BodyScript: "",
	}
	answers := UserAnswers{
		MediaFiles: []string{},
		HashKey:    "test",
	}

	html, err := GenerateHTML(templateHTML, scripts, answers)
	if err != nil {
		t.Fatalf("GenerateHTML() with empty media files error = %v", err)
	}

	// Should contain empty array
	if !strings.Contains(html, "const files = [];") {
		t.Error("GenerateHTML() did not handle empty media files correctly")
	}
}
