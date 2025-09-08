package restore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RestoreFile struct {
	BackupPath   string
	BackupName   string
	GameType     GameType
	UserNumber   string
	TargetPath   string
	TargetName   string
}

type GameType int

const (
	HollowKnight GameType = iota
	Silksong
)

type RestoreResult struct {
	BackupFile    string
	TargetFile    string
	Success       bool
	Error         string
	Skipped       bool
	OverwriteInfo string
}

type RestoreSummary struct {
	Results       []RestoreResult
	TotalFiles    int
	SuccessCount  int
	SkippedCount  int
	FailureCount  int
}

var (
	hkBackupRegex = regexp.MustCompile(`^hk-user([0-9]+)-(\d{8})(?:-(\d+))?\.dat$`)
	ssBackupRegex = regexp.MustCompile(`^ss-user([0-9]+)-(\d{8})(?:-(\d+))?\.dat$`)
)

// findSilksongUserDir finds the user-specific directory in Silksong save path
// or creates one if none exists (for first-time restore)
func findSilksongUserDir(ssSaveDir string) (string, error) {
	// Check if the main directory exists
	if _, err := os.Stat(ssSaveDir); os.IsNotExist(err) {
		return "", fmt.Errorf("Silksong save directory does not exist: %s", ssSaveDir)
	}

	// Look for existing user directories (numeric folders)
	entries, err := os.ReadDir(ssSaveDir)
	if err != nil {
		return "", fmt.Errorf("failed to read Silksong directory: %w", err)
	}

	// Find the first numeric directory that exists
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it's a numeric directory name
			if matched, _ := regexp.MatchString(`^\d+$`, entry.Name()); matched {
				userDir := filepath.Join(ssSaveDir, entry.Name())
				return userDir, nil
			}
		}
	}

	// No user directory found - we'll create a default one when restoring
	// Use a generic Steam-like ID for new users
	defaultUserDir := filepath.Join(ssSaveDir, "76561198000000000")
	return defaultUserDir, nil
}

func ParseBackupFiles(backupPaths []string, hkSaveDir, ssSaveDir string) ([]*RestoreFile, []string) {
	var restoreFiles []*RestoreFile
	var invalidFiles []string

	for _, backupPath := range backupPaths {
		fileName := filepath.Base(backupPath)
		fileName = strings.ToLower(fileName) // Case insensitive matching
		
		var gameType GameType
		var userNum string
		var targetDir string
		
		if matches := hkBackupRegex.FindStringSubmatch(fileName); len(matches) >= 3 {
			gameType = HollowKnight
			userNum = matches[1]
			targetDir = hkSaveDir
		} else if matches := ssBackupRegex.FindStringSubmatch(fileName); len(matches) >= 3 {
			gameType = Silksong
			userNum = matches[1]
			// For Silksong, find the user-specific subdirectory
			silksongUserDir, err := findSilksongUserDir(ssSaveDir)
			if err != nil {
				invalidFiles = append(invalidFiles, fmt.Sprintf("%s (Silksong directory error: %v)", fileName, err))
				continue
			}
			targetDir = silksongUserDir
		} else {
			invalidFiles = append(invalidFiles, fileName)
			continue
		}

		targetName := fmt.Sprintf("user%s.dat", userNum)
		targetPath := filepath.Join(targetDir, targetName)

		restoreFile := &RestoreFile{
			BackupPath: backupPath,
			BackupName: filepath.Base(backupPath),
			GameType:   gameType,
			UserNumber: userNum,
			TargetPath: targetPath,
			TargetName: targetName,
		}

		restoreFiles = append(restoreFiles, restoreFile)
	}

	return restoreFiles, invalidFiles
}

type OverwriteDecision int

const (
	Overwrite OverwriteDecision = iota
	Skip
	Cancel
)

type OverwriteCallback func(targetFile string) OverwriteDecision

func RestoreBackups(restoreFiles []*RestoreFile, overwriteCallback OverwriteCallback) *RestoreSummary {
	summary := &RestoreSummary{
		Results:    make([]RestoreResult, 0),
		TotalFiles: len(restoreFiles),
	}

	for _, rf := range restoreFiles {
		result := RestoreResult{
			BackupFile: rf.BackupName,
			TargetFile: rf.TargetName,
		}

		// Check if target directory exists, create if needed (especially for Silksong)
		targetDir := filepath.Dir(rf.TargetPath)
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			// Try to create the directory (particularly important for Silksong user dirs)
			if rf.GameType == Silksong {
				if err := os.MkdirAll(targetDir, 0755); err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("Failed to create Silksong user directory: %s. Error: %v", targetDir, err)
					summary.FailureCount++
					summary.Results = append(summary.Results, result)
					continue
				}
				result.OverwriteInfo = "Created Silksong user directory"
			} else {
				result.Success = false
				result.Error = fmt.Sprintf("Target directory does not exist: %s. The game may not have created save files yet.", targetDir)
				summary.FailureCount++
				summary.Results = append(summary.Results, result)
				continue
			}
		}

		// Check if target file already exists
		if _, err := os.Stat(rf.TargetPath); err == nil {
			if overwriteCallback != nil {
				decision := overwriteCallback(rf.TargetName)
				switch decision {
				case Skip:
					result.Success = false
					result.Skipped = true
					result.OverwriteInfo = "User chose to skip"
					summary.SkippedCount++
					summary.Results = append(summary.Results, result)
					continue
				case Cancel:
					// Return early with current results
					return summary
				case Overwrite:
					result.OverwriteInfo = "User chose to overwrite"
				}
			}
		}

		// Perform the copy
		err := copyFile(rf.BackupPath, rf.TargetPath)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			summary.FailureCount++
		} else {
			result.Success = true
			summary.SuccessCount++
		}

		summary.Results = append(summary.Results, result)
	}

	return summary
}

func copyFile(src, dest string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil && err == nil {
			// If copy succeeded but close failed, remove the partial file
			os.Remove(dest)
		}
	}()

	// Copy the file contents
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		os.Remove(dest) // Remove partial file on copy failure
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Try to preserve modification time from source
	if stat, err := os.Stat(src); err == nil {
		os.Chtimes(dest, stat.ModTime(), stat.ModTime())
	}

	return nil
}

func (gt GameType) String() string {
	switch gt {
	case HollowKnight:
		return "Hollow Knight"
	case Silksong:
		return "Silksong"
	default:
		return "Unknown"
	}
}

func (gt GameType) Prefix() string {
	switch gt {
	case HollowKnight:
		return "hk"
	case Silksong:
		return "ss"
	default:
		return "unknown"
	}
}