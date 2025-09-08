package main

import (
	"fmt"
	"log"
	"path/filepath"

	"knightsaver/internal/restore"
	"knightsaver/internal/paths"
)

func main() {
	fmt.Println("=== Testing Restore Functionality ===\n")

	// Get game paths
	gamePaths, err := paths.GetGamePaths()
	if err != nil {
		log.Fatalf("Error getting paths: %v", err)
	}

	// Test files - using relative paths
	testFiles := []string{
		filepath.Join("test-backups", "hk-user1-20250908.dat"),
		filepath.Join("test-backups", "ss-user2-20250908.dat"),
		"invalid-file.dat",  // Should be rejected
	}

	fmt.Printf("Hollow Knight Path: %s\n", gamePaths.HollowKnight)
	fmt.Printf("Silksong Path: %s\n\n", gamePaths.Silksong)

	// Parse backup files
	fmt.Println("=== Parsing Backup Files ===")
	restoreFiles, invalidFiles := restore.ParseBackupFiles(testFiles, gamePaths.HollowKnight, gamePaths.Silksong)

	if len(invalidFiles) > 0 {
		fmt.Printf("❌ Invalid files: %v\n", invalidFiles)
	}

	fmt.Printf("✅ Valid files to restore: %d\n", len(restoreFiles))
	for _, rf := range restoreFiles {
		fmt.Printf("  🎮 %s -> %s (%s)\n", rf.BackupName, rf.TargetName, rf.GameType.String())
		fmt.Printf("      Target: %s\n", rf.TargetPath)
	}

	// Test overwrite callback
	fmt.Println("\n=== Testing Restore Process ===")
	overwriteCallback := func(targetFile string) restore.OverwriteDecision {
		fmt.Printf("❓ Would ask about overwriting: %s (choosing Skip for safety)\n", targetFile)
		return restore.Skip // Skip for safety during testing
	}

	if len(restoreFiles) > 0 {
		fmt.Println("🔄 Running restore simulation...")
		summary := restore.RestoreBackups(restoreFiles, overwriteCallback)
		
		fmt.Printf("📊 Results: %d total, %d success, %d skipped, %d failed\n", 
			summary.TotalFiles, summary.SuccessCount, summary.SkippedCount, summary.FailureCount)
			
		for _, result := range summary.Results {
			if result.Success {
				fmt.Printf("  ✅ %s\n", result.BackupFile)
			} else if result.Skipped {
				fmt.Printf("  ⏭️  %s (skipped)\n", result.BackupFile)
			} else {
				fmt.Printf("  ❌ %s: %s\n", result.BackupFile, result.Error)
			}
		}
	}
}