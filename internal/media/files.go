// Package media provides functions for handling media files.
// It includes file discovery, validation, and path manipulation utilities.
package media

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileTypes contains the supported file extensions for audio and video files.
type FileTypes struct {
	// Audio contains supported audio file extensions
	Audio []string
	// Video contains supported video file extensions
	Video []string
}

// DefaultFileTypes returns the default supported file types.
// These are the file types that OBS and modern browsers can typically play.
//
// Returns:
//   - FileTypes: A struct containing default audio and video file extensions
func DefaultFileTypes() FileTypes {
	return FileTypes{
		Audio: []string{".mp3", ".ogg", ".aac"},
		Video: []string{".mp4", ".webm", ".mpeg4", ".m4v", ".mov"},
	}
}

// GetMediaFiles returns a list of media files in the given directory.
// It scans the directory recursively and warns about files with problematic characters.
//
// Parameters:
//   - currentDir: The directory to scan for media files
//   - fileTypes: The file types to look for (audio and video)
//   - verbose: Whether to output verbose logging
//
// Returns:
//   - A slice of file paths (with forward slashes for cross-platform compatibility)
func GetMediaFiles(currentDir string, fileTypes FileTypes, verbose bool) []string {
	if verbose {
		fmt.Println("Scanning for media files...")
	}

	mediaFiles := []string{}
	problematicFiles := []string{}
	allTypes := append(fileTypes.Audio, fileTypes.Video...)

	filepath.WalkDir(currentDir, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			// Log the error but continue scanning
			log.Printf("Warning: Cannot access %s: %v\n", path, err)
			// Skip the directory if we can't access it
			if file != nil && file.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !file.IsDir() && IsValidFileType(file, allTypes) {
			fixedFilePath := FixFilePath(path)

			// Check for problematic characters that cause issues with file:// URLs
			if HasProblematicChars(file.Name()) {
				problematicFiles = append(problematicFiles, file.Name())
			}

			mediaFiles = append(mediaFiles, fixedFilePath)

			if verbose {
				fmt.Printf("  Found: %s\n", file.Name())
			}
		}
		return nil
	})

	if verbose {
		fmt.Printf("Found %d media files\n", len(mediaFiles))
	}

	// Warn about problematic files
	if len(problematicFiles) > 0 {
		fmt.Printf("\n⚠️  WARNING: Found files with characters that may cause playback issues:\n")
		for _, fileName := range problematicFiles {
			fmt.Printf("   • %s\n", fileName)
		}
		fmt.Printf("\nProblematic characters: # ; ? : @ & = + $ ,\n")
		fmt.Printf("Consider renaming these files or use --sanitize flag.\n\n")
	}

	return mediaFiles
}

// GetMediaFilesConcurrent returns a list of media files in the given directory using concurrent scanning.
// This is faster for large media libraries.
//
// Parameters:
//   - currentDir: The directory to scan for media files
//   - fileTypes: The file types to look for (audio and video)
//   - verbose: Whether to output verbose logging
//   - workers: Number of concurrent workers (recommended: 4-8)
//
// Returns:
//   - A slice of file paths (with forward slashes for cross-platform compatibility)
func GetMediaFilesConcurrent(currentDir string, fileTypes FileTypes, verbose bool, workers int) []string {
	if verbose {
		fmt.Printf("Scanning for media files (using %d workers)...\n", workers)
	}

	type fileResult struct {
		path          string
		isProblematic bool
		fileName      string
	}

	fileChan := make(chan fileResult, 100)
	pathChan := make(chan string, 100)
	var wg sync.WaitGroup
	var mediaFiles []string
	var problematicFiles []string
	var mu sync.Mutex

	allTypes := append(fileTypes.Audio, fileTypes.Video...)

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range pathChan {
				info, err := os.Stat(path)
				if err != nil {
					log.Printf("Warning: Cannot access %s: %v\n", path, err)
					continue
				}

				if !info.IsDir() {
					// Create a DirEntry-like struct for IsValidFileType
					fileName := filepath.Base(path)
					if IsValidFileType(&fileInfo{name: fileName}, allTypes) {
						fixedPath := FixFilePath(path)
						result := fileResult{
							path:          fixedPath,
							isProblematic: HasProblematicChars(fileName),
							fileName:      fileName,
						}
						fileChan <- result
					}
				}
			}
		}()
	}

	// Collector goroutine
	go func() {
		for result := range fileChan {
			mu.Lock()
			mediaFiles = append(mediaFiles, result.path)
			if result.isProblematic {
				problematicFiles = append(problematicFiles, result.fileName)
			}
			if verbose {
				fmt.Printf("  Found: %s\n", result.fileName)
			}
			mu.Unlock()
		}
	}()

	// Walk directory and send paths to workers
	filepath.WalkDir(currentDir, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Warning: Cannot access %s: %v\n", path, err)
			if file != nil && file.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !file.IsDir() {
			pathChan <- path
		}
		return nil
	})

	close(pathChan)
	wg.Wait()
	close(fileChan)

	// Wait a bit for collector to finish
	mu.Lock()
	fileCount := len(mediaFiles)
	mu.Unlock()

	if verbose {
		fmt.Printf("Found %d media files\n", fileCount)
	}

	// Warn about problematic files
	if len(problematicFiles) > 0 {
		fmt.Printf("\n⚠️  WARNING: Found files with characters that may cause playback issues:\n")
		for _, fileName := range problematicFiles {
			fmt.Printf("   • %s\n", fileName)
		}
		fmt.Printf("\nProblematic characters: # ; ? : @ & = + $ ,\n")
		fmt.Printf("Consider renaming these files or use --sanitize flag.\n\n")
	}

	return mediaFiles
}

// fileInfo is a simple implementation of fs.DirEntry for concurrent processing
type fileInfo struct {
	name string
}

func (f *fileInfo) Name() string               { return f.name }
func (f *fileInfo) IsDir() bool                { return false }
func (f *fileInfo) Type() fs.FileMode          { return 0 }
func (f *fileInfo) Info() (fs.FileInfo, error) { return nil, nil }

// IsValidFileType checks if a file has a supported extension.
//
// Parameters:
//   - file: The directory entry to check
//   - fileExts: List of valid file extensions (e.g., [".mp4", ".mp3"])
//
// Returns:
//   - true if the file has one of the supported extensions
func IsValidFileType(file fs.DirEntry, fileExts []string) bool {
	fileName := file.Name()
	for _, ext := range fileExts {
		if strings.HasSuffix(strings.ToLower(fileName), strings.ToLower(ext)) {
			return true
		}
	}
	return false
}

// FixFilePath converts file path separators to forward slashes for HTML compatibility.
// This ensures that file paths work correctly in HTML regardless of the OS.
//
// Parameters:
//   - filePath: The file path to normalize
//
// Returns:
//   - A file path with forward slashes (/) as separators
func FixFilePath(filePath string) string {
	// Only do replacement if needed (Windows paths use backslashes)
	if os.PathSeparator != '/' {
		return strings.ReplaceAll(filePath, string(os.PathSeparator), "/")
	}
	return filePath
}

// HasProblematicChars checks if a filename contains characters that cause issues with file:// URLs.
// These characters can interfere with URL parsing or file access in browsers.
//
// Parameters:
//   - fileName: The filename to check
//
// Returns:
//   - true if the filename contains any problematic characters
func HasProblematicChars(fileName string) bool {
	problematicChars := []string{"#", ";", "?", ":", "@", "&", "=", "+", "$", ","}
	for _, char := range problematicChars {
		if strings.Contains(fileName, char) {
			return true
		}
	}
	return false
}

// SanitizeFileName replaces problematic characters in a filename with safe alternatives.
// This helps ensure files work properly with file:// URLs.
//
// Parameters:
//   - fileName: The filename to sanitize
//
// Returns:
//   - A sanitized version of the filename
func SanitizeFileName(fileName string) string {
	problematicChars := map[string]string{
		"#": "-",
		";": "-",
		"?": "",
		":": "-",
		"@": "-",
		"&": "and",
		"=": "-",
		"+": "plus",
		"$": "",
		",": "-",
	}

	result := fileName
	for char, replacement := range problematicChars {
		result = strings.ReplaceAll(result, char, replacement)
	}
	return result
}

// SanitizeMediaFiles renames media files with problematic characters.
// Returns the number of files renamed and any errors encountered.
//
// Parameters:
//   - mediaFiles: List of file paths to check and sanitize
//
// Returns:
//   - renamedCount: Number of files successfully renamed
//   - errors: Slice of errors encountered during renaming
func SanitizeMediaFiles(mediaFiles []string) (int, []error) {
	var renamedCount int
	var errors []error

	for _, filePath := range mediaFiles {
		dir := filepath.Dir(filePath)
		fileName := filepath.Base(filePath)

		if HasProblematicChars(fileName) {
			sanitized := SanitizeFileName(fileName)
			newPath := filepath.Join(dir, sanitized)

			err := os.Rename(filePath, newPath)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to rename %s: %w", fileName, err))
			} else {
				fmt.Printf("✓ Renamed: %s → %s\n", fileName, sanitized)
				renamedCount++
			}
		}
	}

	return renamedCount, errors
}

// ValidatePath ensures that a file path is within the expected base directory.
// This prevents directory traversal attacks.
//
// Parameters:
//   - basePath: The base directory that files should be within
//   - filePath: The file path to validate
//
// Returns:
//   - error if the path is outside the base directory or if there's an error resolving paths
func ValidatePath(basePath, filePath string) error {
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Ensure the file path starts with the base path
	if !strings.HasPrefix(absFile, absBase) {
		return fmt.Errorf("path traversal detected: %s is outside %s", filePath, basePath)
	}

	return nil
}

// RemoveTransitionVideo removes the transition video from the list of media files.
//
// Parameters:
//   - transitionVideo: The path to the transition video to remove
//   - mediaFiles: The list of media file paths
//
// Returns:
//   - A new slice with the transition video removed
func RemoveTransitionVideo(transitionVideo string, mediaFiles []string) []string {
	var result []string
	for _, file := range mediaFiles {
		if file != transitionVideo {
			result = append(result, file)
		}
	}
	return result
}
