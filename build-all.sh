#!/bin/bash

# KnightSaver Cross-Platform Build Script
# This script builds KnightSaver for Windows, Linux, and macOS

set -e

echo "🚀 Building KnightSaver for all platforms..."

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/
mkdir -p dist/{windows,linux,macos}

# Build version info
VERSION=${1:-"1.0.0"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "📦 Building version: $VERSION"
echo "🕒 Build time: $BUILD_TIME"

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        echo "❌ Docker is not running. Please start Docker Desktop first."
        echo "   You can start it from Applications or by running: open -a Docker"
        echo "   Wait for Docker to start, then run this script again."
        return 1
    fi
    echo "✅ Docker is running"
    return 0
}

# Build macOS (native)
echo ""
echo "🍎 Building for macOS..."
ICON_PATH="$(pwd)/assets/icon.png"
fyne package --target darwin --icon "$ICON_PATH" --name KnightSaver --app-id com.calo.knightsaver --src ./cmd/knightsaver
if [ -d "KnightSaver.app" ]; then
    mv KnightSaver.app dist/macos/
    echo "✅ macOS build complete: dist/macos/KnightSaver.app"
else
    echo "❌ macOS build failed"
fi

# Try Docker builds
if check_docker; then
    echo ""
    echo "🪟 Building for Windows..."
    fyne-cross windows --icon "$ICON_PATH" --app-id com.calo.knightsaver
    
    echo ""
    echo "🐧 Building for Linux..."
    fyne-cross linux --icon "$ICON_PATH" --app-id com.calo.knightsaver
    
    # Move fyne-cross outputs to dist folder
    if [ -d "fyne-cross/dist" ]; then
        echo "📦 Organizing distribution files..."
        
        # Windows
        if [ -d "fyne-cross/dist/windows-amd64" ]; then
            cp -r fyne-cross/dist/windows-amd64/* dist/windows/ 2>/dev/null || true
        fi
        if [ -d "fyne-cross/dist/windows-arm64" ]; then
            cp -r fyne-cross/dist/windows-arm64/* dist/windows/ 2>/dev/null || true
        fi
        
        # Linux
        if [ -d "fyne-cross/dist/linux-amd64" ]; then
            cp -r fyne-cross/dist/linux-amd64/* dist/linux/ 2>/dev/null || true
        fi
        if [ -d "fyne-cross/dist/linux-arm64" ]; then
            cp -r fyne-cross/dist/linux-arm64/* dist/linux/ 2>/dev/null || true
        fi
        
        echo "✅ Cross-platform builds complete"
    fi
else
    echo ""
    echo "⚠️  Skipping Docker-based builds (Windows/Linux)"
    echo "   To build for Windows/Linux:"
    echo "   1. Start Docker Desktop"
    echo "   2. Run this script again"
fi

# Create README for releases
echo ""
echo "📝 Creating distribution README..."

cat > dist/README.md << 'EOF'
# KnightSaver Distribution Files

KnightSaver is a cross-platform GUI application for backing up and restoring Hollow Knight and Hollow Knight: Silksong save files.

## Installation

### Windows
- Download files from the `windows/` folder
- Run `KnightSaver.exe`

### macOS  
- Download `KnightSaver.app` from the `macos/` folder
- First time: Right-click → Open (app is unsigned)
- Subsequently: Double-click to run

### Linux
- Download the executable from the `linux/` folder
- Make executable: `chmod +x KnightSaver`
- Run: `./KnightSaver`

## Features

- **Cross-platform**: Windows, macOS, Linux
- **Dual-game support**: Hollow Knight and Silksong
- **Smart backup naming**: Date-stamped with collision handling  
- **Restore validation**: Prevents mixing save files between games
- **Native file dialogs**: OS-native folder and file pickers
- **Automatic path detection**: Finds game directories automatically

## Usage

1. Click **SCAN DIRS** to find save files
2. Select files using checkboxes
3. **EXPORT BACKUP**: Choose destination and backup selected saves
4. **RESTORE BACKUP**: Select backup files to restore

## Game Save Locations

### Hollow Knight
- **Windows**: `%AppData%..\LocalLow\Team Cherry\Hollow Knight`
- **macOS**: `~/Library/Application Support/unity.Team Cherry.Hollow Knight`
- **Linux**: `~/.config/unity3d/Team Cherry/Hollow Knight`

### Silksong
- **Windows**: `%USERPROFILE%/AppData/LocalLow/Team Cherry/Hollow Knight Silksong/`
- **macOS**: `~/Library/Application Support/unity.Team-Cherry.Silksong/`
- **Linux**: `~/.config/unity3d/Team Cherry/Hollow Knight Silksong/`

## License

MIT License - See repository for full license text.
EOF

echo ""
echo "🎉 Build process complete!"
echo ""
echo "📁 Distribution files:"
find dist/ -type f -name "*" | sort

echo ""
echo "🚀 Ready for GitHub release!"
echo "   Upload the entire 'dist/' folder contents to your GitHub release"