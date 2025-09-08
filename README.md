# KnightSaver

A cross-platform GUI application for backing up and restoring Hollow Knight and Hollow Knight: Silksong save files.

## Features

- **Cross-platform**: Works on Windows, macOS, and Linux
- **Dual-game support**: Manages saves for both Hollow Knight and Silksong
- **Smart backup naming**: Uses prefixed names with date stamps and collision handling
- **Restore validation**: Prevents mixing save files between games
- **Native file dialogs**: Uses OS-native folder and file pickers
- **Automatic path detection**: Finds game save directories automatically per OS

## Usage

1. **Launch the application**
2. **Click "Scan Dirs"** to find save files in both game directories
3. **Select files** using checkboxes or "Select All" options
4. **Export Backup**: Choose a destination folder and backup selected saves
5. **Restore Backup**: Select backup files to restore to the appropriate game directories

## Game Save Locations

The application automatically detects save directories for each operating system:

### Hollow Knight
- **Windows**: `%AppData%..\LocalLow\Team Cherry\Hollow Knight`
- **macOS**: `~/Library/Application Support/unity.Team Cherry.Hollow Knight`
- **Linux**: `~/.config/unity3d/Team Cherry/Hollow Knight`

### Silksong  
- **Windows**: `%USERPROFILE%/AppData/LocalLow/Team Cherry/Hollow Knight Silksong/`
- **macOS**: `~/Library/Application Support/unity.Team-Cherry.Silksong/`
- **Linux**: `~/.config/unity3d/Team Cherry/Hollow Knight Silksong/`

## Backup File Format

Backup files use a specific naming convention:
- **Hollow Knight**: `hk-userN-YYYYMMDD.dat`
- **Silksong**: `ss-userN-YYYYMMDD.dat`

If multiple backups are created on the same day, a numeric suffix is added (e.g., `-2`, `-3`).

## Restore Behavior

### Hollow Knight
Restores directly to the main save directory.

### Silksong  
Silksong stores saves in user-specific subdirectories (e.g., `84412171/`). KnightSaver automatically:
- **Finds existing user directory** if you have played Silksong
- **Creates a new user directory** if restoring to a fresh Silksong installation
- **Handles multiple user directories** by using the first one found

## Installation Notes

### macOS
The app is unsigned, so you may need to right-click → Open the first time you run it.

### Windows/Linux
No special installation requirements. Simply run the executable.

## License

MIT License - see LICENSE file for details.

## Project Structure

```
├── cmd/knightsaver/main.go     # Application entry point
├── internal/
│   ├── paths/paths.go          # OS path resolution
│   ├── scan/scan.go            # Save file discovery
│   ├── backup/backup.go        # Export functionality  
│   ├── restore/restore.go      # Restore functionality
│   └── ui/app.go              # Fyne GUI application
├── assets/
│   └── icon.png               # Application icon
└── README.md
```