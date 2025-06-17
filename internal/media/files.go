package media

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileTypes contains the supported file extensions
type FileTypes struct {
	Audio []string
	Video []string
}

// DefaultFileTypes returns the default supported file types
func DefaultFileTypes() FileTypes {
	return FileTypes{
		Audio: []string{".mp3", ".ogg", ".aac"},
		Video: []string{".mp4", ".webm", ".mpeg4", ".m4v", ".mov"},
	}
}

// GetMediaFiles returns a list of media files in the given directory
func GetMediaFiles(currentDir string, fileTypes FileTypes) []string {
	mediaFiles := []string{}
	problematicFiles := []string{}
	allTypes := append(fileTypes.Audio, fileTypes.Video...)

	filepath.WalkDir(currentDir, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !file.IsDir() && IsValidFileType(file, allTypes) {
			fixedFilePath := FixFilePath(path)
			
			// Check for problematic characters that cause issues with file:// URLs
			if HasProblematicChars(file.Name()) {
				problematicFiles = append(problematicFiles, file.Name())
			}
			
			mediaFiles = append(mediaFiles, fixedFilePath)
		}
		return nil
	})
	
	// Warn about problematic files
	if len(problematicFiles) > 0 {
		fmt.Printf("\n⚠️  WARNING: Found files with characters that may cause playback issues:\n")
		for _, fileName := range problematicFiles {
			fmt.Printf("   • %s\n", fileName)
		}
		fmt.Printf("\nProblematic characters: # ; ? : @ & = + $ ,\n")
		fmt.Printf("Consider renaming these files to avoid potential issues.\n\n")
	}
	
	return mediaFiles
}

// IsValidFileType checks if a file has a supported extension
func IsValidFileType(file fs.DirEntry, fileExts []string) bool {
	for _, ext := range fileExts {
		if strings.HasSuffix(file.Name(), ext) {
			return true
		}
	}
	return false
}

// FixFilePath converts file path separators to forward slashes for HTML
func FixFilePath(filePath string) string {
	// Only do replacement if needed (Windows paths)
	if os.PathSeparator != '/' {
		return strings.ReplaceAll(filePath, string(os.PathSeparator), "/")
	}
	return filePath
}

// HasProblematicChars checks if a filename contains characters that cause issues with file:// URLs
func HasProblematicChars(fileName string) bool {
	problematicChars := []string{"#", ";", "?", ":", "@", "&", "=", "+", "$", ","}
	for _, char := range problematicChars {
		if strings.Contains(fileName, char) {
			return true
		}
	}
	return false
}

// RemoveTransitionVideo removes the transition video from the list of media files
func RemoveTransitionVideo(transitionVideo string, mediaFiles []string) []string {
	var result []string
	for _, file := range mediaFiles {
		if file != transitionVideo {
			result = append(result, file)
		}
	}
	return result
}
