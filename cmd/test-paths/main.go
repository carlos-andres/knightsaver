package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"knightsaver/internal/paths"
	"knightsaver/internal/scan"
)

func main() {
	fmt.Println("=== KnightSaver Path & File Testing ===\n")

	// Get paths
	gamePaths, err := paths.GetGamePaths()
	if err != nil {
		fmt.Printf("Error getting paths: %v\n", err)
		return
	}

	fmt.Printf("Hollow Knight Path: %s\n", gamePaths.HollowKnight)
	fmt.Printf("Silksong Path: %s\n\n", gamePaths.Silksong)

	// Test Hollow Knight
	fmt.Println("=== HOLLOW KNIGHT ===")
	testDirectory("Hollow Knight", gamePaths.HollowKnight)

	fmt.Println("\n=== SILKSONG ===")
	testDirectory("Silksong", gamePaths.Silksong)
}

func testDirectory(gameName, dirPath string) {
	fmt.Printf("Testing: %s\n", dirPath)

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("❌ Directory does not exist: %s\n", dirPath)
		return
	}
	fmt.Printf("✅ Directory exists\n")

	// List all files
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return
	}

	fmt.Printf("📁 Found %d entries:\n", len(entries))
	saveFileRegex := regexp.MustCompile(`(?i)^user[0-9]+\.dat$`)
	
	saveFiles := 0
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("  📂 %s (directory)\n", entry.Name())
		} else {
			if saveFileRegex.MatchString(entry.Name()) {
				fmt.Printf("  🎮 %s (SAVE FILE)\n", entry.Name())
				saveFiles++
				
				// Get file info
				fullPath := filepath.Join(dirPath, entry.Name())
				if info, err := os.Stat(fullPath); err == nil {
					sizeKB := info.Size() / 1024
					if sizeKB == 0 && info.Size() > 0 {
						sizeKB = 1
					}
					fmt.Printf("      Size: %d KB, Modified: %s\n", sizeKB, info.ModTime().Format("2006-01-02 15:04:05"))
				}
			} else {
				fmt.Printf("  📄 %s (other file)\n", entry.Name())
			}
		}
	}

	// Use scan package
	fmt.Printf("\n🔍 Using scan package:\n")
	saves, err := scan.ScanDirectory(dirPath)
	if err != nil {
		fmt.Printf("❌ Scan error: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d save files using scan package\n", len(saves))
		for _, save := range saves {
			fmt.Printf("  🎮 %s (%d KB, %s)\n", save.Name, save.SizeKB, save.ModTimeStr)
		}
	}
}