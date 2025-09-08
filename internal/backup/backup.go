package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"knightsaver/internal/scan"
)

type GameType int

const (
	HollowKnight GameType = iota
	Silksong
)

type ExportResult struct {
	SourceFile      string
	TargetFile      string
	Success         bool
	Error           string
}

type ExportSummary struct {
	Results       []ExportResult
	TotalFiles    int
	SuccessCount  int
	FailureCount  int
	DestFolder    string
}

func ExportBackups(saves []*scan.SaveFile, gameType GameType, destFolder string) *ExportSummary {
	summary := &ExportSummary{
		Results:      make([]ExportResult, 0),
		TotalFiles:   len(saves),
		DestFolder:   destFolder,
	}

	today := time.Now().Format("20060102")
	var prefix string
	
	switch gameType {
	case HollowKnight:
		prefix = "hk"
	case Silksong:
		prefix = "ss"
	}

	for _, save := range saves {
		result := ExportResult{
			SourceFile: save.Name,
		}

		// Extract user number from filename (e.g., "user1.dat" -> "1")
		userNum := strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(save.Name), "user"), ".dat")
		
		// Generate target filename with collision handling
		targetName := generateBackupFilename(prefix, userNum, today, destFolder)
		targetPath := filepath.Join(destFolder, targetName)
		
		result.TargetFile = targetName

		// Copy the file
		err := copyFile(save.Path, targetPath)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			summary.FailureCount++
		} else {
			result.Success = true
			summary.SuccessCount++
			
			// Try to preserve modification time
			if stat, err := os.Stat(save.Path); err == nil {
				os.Chtimes(targetPath, stat.ModTime(), stat.ModTime())
			}
		}

		summary.Results = append(summary.Results, result)
	}

	return summary
}

func generateBackupFilename(prefix, userNum, date, destFolder string) string {
	baseFilename := fmt.Sprintf("%s-user%s-%s.dat", prefix, userNum, date)
	targetPath := filepath.Join(destFolder, baseFilename)
	
	// Check if file already exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return baseFilename
	}

	// File exists, generate suffix
	suffix := 2
	for {
		filename := fmt.Sprintf("%s-user%s-%s-%d.dat", prefix, userNum, date, suffix)
		targetPath = filepath.Join(destFolder, filename)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return filename
		}
		suffix++
	}
}

func copyFile(src, dest string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
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