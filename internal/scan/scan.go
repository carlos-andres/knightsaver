package scan

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

type SaveFile struct {
	Name         string
	Path         string
	Size         int64
	ModTime      time.Time
	SizeKB       int64
	ModTimeStr   string
}

var saveFileRegex = regexp.MustCompile(`(?i)^user[0-9]+\.dat$`)

func ScanDirectory(dirPath string) ([]*SaveFile, error) {
	var saves []*SaveFile

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return saves, nil // Return empty slice, not an error
	}

	// Walk through the directory recursively to find save files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/directories we can't access
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a save file
		name := info.Name()
		if !saveFileRegex.MatchString(name) {
			return nil
		}

		sizeKB := info.Size() / 1024
		if sizeKB == 0 && info.Size() > 0 {
			sizeKB = 1 // Show at least 1 KB for non-empty files
		}

		save := &SaveFile{
			Name:       name,
			Path:       path,
			Size:       info.Size(),
			ModTime:    info.ModTime(),
			SizeKB:     sizeKB,
			ModTimeStr: info.ModTime().Format("2006-01-02 15:04:05"),
		}

		saves = append(saves, save)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by modification time, newest first
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].ModTime.After(saves[j].ModTime)
	})

	return saves, nil
}