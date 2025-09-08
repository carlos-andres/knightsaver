package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

type GamePaths struct {
	HollowKnight string
	Silksong     string
}

func GetGamePaths() (*GamePaths, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsPaths()
	case "darwin":
		return getMacOSPaths()
	case "linux":
		return getLinuxPaths()
	default:
		return getLinuxPaths() // fallback to Linux paths
	}
}

func getWindowsPaths() (*GamePaths, error) {
	userProfile := os.Getenv("USERPROFILE")
	appData := os.Getenv("APPDATA")
	
	// For LocalLow, we need to go up one level from AppData and then into LocalLow
	var localLow string
	if appData != "" {
		localLow = filepath.Join(filepath.Dir(appData), "LocalLow")
	} else if userProfile != "" {
		localLow = filepath.Join(userProfile, "AppData", "LocalLow")
	}

	hkPath := filepath.Join(localLow, "Team Cherry", "Hollow Knight")
	ssPath := filepath.Join(userProfile, "AppData", "LocalLow", "Team Cherry", "Hollow Knight Silksong")

	return &GamePaths{
		HollowKnight: hkPath,
		Silksong:     ssPath,
	}, nil
}

func getMacOSPaths() (*GamePaths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	hkPath := filepath.Join(homeDir, "Library", "Application Support", "unity.Team Cherry.Hollow Knight")
	ssPath := filepath.Join(homeDir, "Library", "Application Support", "unity.Team-Cherry.Silksong")

	return &GamePaths{
		HollowKnight: hkPath,
		Silksong:     ssPath,
	}, nil
}

func getLinuxPaths() (*GamePaths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	hkPath := filepath.Join(homeDir, ".config", "unity3d", "Team Cherry", "Hollow Knight")
	ssPath := filepath.Join(homeDir, ".config", "unity3d", "Team Cherry", "Hollow Knight Silksong")

	return &GamePaths{
		HollowKnight: hkPath,
		Silksong:     ssPath,
	}, nil
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}