package media

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// mockDirEntry is a simple mock for fs.DirEntry
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() fs.FileMode          { return 0 }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestDefaultFileTypes(t *testing.T) {
	fileTypes := DefaultFileTypes()

	expectedAudio := []string{".mp3", ".ogg", ".aac"}
	expectedVideo := []string{".mp4", ".webm", ".mpeg4", ".m4v", ".mov"}

	if len(fileTypes.Audio) != len(expectedAudio) {
		t.Errorf("Audio file types length = %v, want %v", len(fileTypes.Audio), len(expectedAudio))
	}

	if len(fileTypes.Video) != len(expectedVideo) {
		t.Errorf("Video file types length = %v, want %v", len(fileTypes.Video), len(expectedVideo))
	}
}

func TestIsValidFileType(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		exts     []string
		want     bool
	}{
		{"mp4 file", "video.mp4", []string{".mp4", ".webm"}, true},
		{"MP4 uppercase", "video.MP4", []string{".mp4", ".webm"}, true},
		{"webm file", "video.webm", []string{".mp4", ".webm"}, true},
		{"mp3 file", "audio.mp3", []string{".mp3", ".ogg"}, true},
		{"invalid file", "document.txt", []string{".mp4", ".mp3"}, false},
		{"no extension", "file", []string{".mp4"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &mockDirEntry{name: tt.fileName, isDir: false}
			got := IsValidFileType(entry, tt.exts)
			if got != tt.want {
				t.Errorf("IsValidFileType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasProblematicChars(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{"clean filename", "video.mp4", false},
		{"clean with spaces", "my video.mp4", false},
		{"clean with dash", "my-video.mp4", false},
		{"clean with underscore", "my_video.mp4", false},
		{"has hash", "video#1.mp4", true},
		{"has semicolon", "vid;eo.mp4", true},
		{"has question mark", "video?.mp4", true},
		{"has colon", "video:test.mp4", true},
		{"has at sign", "video@test.mp4", true},
		{"has ampersand", "video&test.mp4", true},
		{"has equals", "video=test.mp4", true},
		{"has plus", "video+test.mp4", true},
		{"has dollar", "video$test.mp4", true},
		{"has comma", "video,test.mp4", true},
		{"multiple problems", "video#1;test.mp4", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasProblematicChars(tt.fileName)
			if got != tt.want {
				t.Errorf("HasProblematicChars(%q) = %v, want %v", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     string
	}{
		{"clean filename", "video.mp4", "video.mp4"},
		{"has hash", "video#1.mp4", "video-1.mp4"},
		{"has semicolon", "vid;eo.mp4", "vid-eo.mp4"},
		{"has question mark", "video?.mp4", "video.mp4"},
		{"has colon", "video:test.mp4", "video-test.mp4"},
		{"has ampersand", "video&test.mp4", "videoandtest.mp4"},
		{"has plus", "video+test.mp4", "videoplustest.mp4"},
		{"has multiple problems", "video#1;test?.mp4", "video-1-test.mp4"},
		{"complex case", "earthbig # 3.mp4", "earthbig - 3.mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFileName(tt.fileName)
			if got != tt.want {
				t.Errorf("SanitizeFileName(%q) = %q, want %q", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestFixFilePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"unix path", "/home/user/video.mp4", "/home/user/video.mp4"},
		{"relative path", "videos/test.mp4", "videos/test.mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FixFilePath(tt.path)
			// On Unix systems, path should remain unchanged
			// On Windows, backslashes would be converted to forward slashes
			if os.PathSeparator == '/' && got != tt.want {
				t.Errorf("FixFilePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		basePath string
		filePath string
		wantErr  bool
	}{
		{
			name:     "valid path within base",
			basePath: tempDir,
			filePath: filepath.Join(tempDir, "video.mp4"),
			wantErr:  false,
		},
		{
			name:     "valid subdirectory path",
			basePath: tempDir,
			filePath: filepath.Join(tempDir, "subdir", "video.mp4"),
			wantErr:  false,
		},
		{
			name:     "path traversal attempt",
			basePath: tempDir,
			filePath: filepath.Join(tempDir, "..", "..", "etc", "passwd"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.basePath, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveTransitionVideo(t *testing.T) {
	tests := []struct {
		name            string
		transitionVideo string
		mediaFiles      []string
		want            []string
	}{
		{
			name:            "remove existing video",
			transitionVideo: "transition.mp4",
			mediaFiles:      []string{"video1.mp4", "transition.mp4", "video2.mp4"},
			want:            []string{"video1.mp4", "video2.mp4"},
		},
		{
			name:            "remove non-existent video",
			transitionVideo: "missing.mp4",
			mediaFiles:      []string{"video1.mp4", "video2.mp4"},
			want:            []string{"video1.mp4", "video2.mp4"},
		},
		{
			name:            "empty list",
			transitionVideo: "transition.mp4",
			mediaFiles:      []string{},
			want:            []string{},
		},
		{
			name:            "multiple occurrences",
			transitionVideo: "dup.mp4",
			mediaFiles:      []string{"dup.mp4", "video1.mp4", "dup.mp4", "video2.mp4"},
			want:            []string{"video1.mp4", "video2.mp4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveTransitionVideo(tt.transitionVideo, tt.mediaFiles)
			if len(got) != len(tt.want) {
				t.Errorf("RemoveTransitionVideo() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("RemoveTransitionVideo() got[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
