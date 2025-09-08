#!/bin/bash

# KnightSaver Release Build Script
# Builds distributables for all platforms using fyne-cross (requires Docker)

set -e

VERSION=${1:-"1.0.0"}
ICON_PATH="$(pwd)/assets/icon.png"
APP_ID="com.calo.knightsaver"

echo "🚀 KnightSaver Release Build v${VERSION}"
echo "================================================="

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        echo ""
        echo "❌ Docker is not running!"
        echo ""
        echo "To build Windows and Linux versions:"
        echo "1. Start Docker Desktop from Applications"
        echo "2. Wait for Docker to fully start"
        echo "3. Run this script again"
        echo ""
        echo "For now, building macOS version only..."
        return 1
    fi
    echo "✅ Docker is running"
    return 0
}

# Clean previous builds
echo ""
echo "🧹 Cleaning previous builds..."
rm -rf dist/ fyne-cross/
mkdir -p dist/{windows,linux,macos}

# Build macOS (native - no Docker needed)
echo ""
echo "🍎 Building macOS version..."
fyne package --target darwin --icon "$ICON_PATH" --name KnightSaver --app-id "$APP_ID" --src ./cmd/knightsaver

if [ -d "KnightSaver.app" ]; then
    mv KnightSaver.app dist/macos/
    echo "✅ macOS: dist/macos/KnightSaver.app"
else
    echo "❌ macOS build failed"
fi

# Try Docker builds for Windows and Linux
DOCKER_AVAILABLE=false
if check_docker; then
    DOCKER_AVAILABLE=true
    
    echo ""
    echo "🪟 Building Windows version..."
    fyne-cross windows --icon "$ICON_PATH" --app-id "$APP_ID"
    
    echo ""
    echo "🐧 Building Linux version..."
    fyne-cross linux --icon "$ICON_PATH" --app-id "$APP_ID"
    
    # Move fyne-cross outputs to proper dist structure
    echo ""
    echo "📦 Organizing cross-platform builds..."
    
    # Windows files
    if [ -d "fyne-cross/dist" ]; then
        find fyne-cross/dist -name "*.exe" -exec cp {} dist/windows/ \;
        find fyne-cross/dist -name "*.tar.xz" -path "*/windows-*" -exec cp {} dist/windows/ \;
        
        # Linux files  
        find fyne-cross/dist -name "*.tar.xz" -path "*/linux-*" -exec cp {} dist/linux/ \;
        
        # Copy any loose binaries
        find fyne-cross/dist -name "knightsaver" -exec cp {} dist/linux/ \; 2>/dev/null || true
        
        echo "✅ Cross-platform builds copied to dist/"
    fi
fi

# Create release documentation
echo ""
echo "📝 Creating release documentation..."

# Main README for the dist folder
cat > dist/README.md << EOF
# KnightSaver v${VERSION} - Distribution Files

Cross-platform GUI application for backing up and restoring Hollow Knight and Silksong save files.

## 📁 Platform Downloads

### 🍎 macOS
- **File**: \`macos/KnightSaver.app\`
- **Installation**: Download and drag to Applications folder
- **First Run**: Right-click → Open (app is unsigned)
- **Supports**: Intel and Apple Silicon Macs

### 🪟 Windows  
- **Files**: \`windows/\` folder
- **Installation**: Download and run the .exe file
- **Supports**: Windows 10/11, both x64 and ARM64

### 🐧 Linux
- **Files**: \`linux/\` folder  
- **Installation**: Download, make executable (\`chmod +x\`), and run
- **Supports**: Most Linux distributions, x64 and ARM64

## 🎮 Features

- **Dual Game Support**: Hollow Knight and Silksong
- **Smart Backups**: Date-stamped with collision handling
- **Cross-Platform**: Native GUI on all platforms
- **Auto-Detection**: Finds game save directories automatically
- **Restore Validation**: Prevents mixing save files between games

## 📖 Quick Start

1. **Launch** the application
2. **Click "SCAN DIRS"** to find your save files
3. **Select saves** using checkboxes  
4. **Export backups** to a safe location
5. **Restore backups** when needed

## 🔧 Game Save Locations

### Hollow Knight
- **Windows**: \`%AppData%\..\LocalLow\Team Cherry\Hollow Knight\`
- **macOS**: \`~/Library/Application Support/unity.Team Cherry.Hollow Knight\`  
- **Linux**: \`~/.config/unity3d/Team Cherry/Hollow Knight\`

### Silksong
- **Windows**: \`%USERPROFILE%/AppData/LocalLow/Team Cherry/Hollow Knight Silksong/\`
- **macOS**: \`~/Library/Application Support/unity.Team-Cherry.Silksong/\`
- **Linux**: \`~/.config/unity3d/Team Cherry/Hollow Knight Silksong/\`

## 🐛 Support

For issues, bug reports, or feature requests, please visit the GitHub repository.

## 📄 License

MIT License - See repository for full license text.

---
*Built with ❤️ using Go and Fyne*
EOF

# Create platform-specific instructions
cat > dist/windows/README-Windows.txt << EOF
KnightSaver for Windows

INSTALLATION:
1. Download KnightSaver.exe (or the specific architecture version)
2. Double-click to run
3. Windows may show a security warning - click "More info" then "Run anyway"

USAGE:
- The app will create an icon in your system tray
- Click "SCAN DIRS" to find your save files
- Your save files are typically in: %AppData%\..\LocalLow\Team Cherry\

TROUBLESHOOTING:
- If the app won't start, make sure you have the latest Visual C++ Redistributable
- Windows Defender may flag the app - add it to exclusions if needed
EOF

cat > dist/linux/README-Linux.txt << EOF
KnightSaver for Linux

INSTALLATION:
1. Download the appropriate file for your architecture
2. Make it executable: chmod +x knightsaver
3. Run: ./knightsaver

DEPENDENCIES:
Most modern Linux distributions should work out of the box. If you get errors:
- Install: libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev
- For Ubuntu/Debian: sudo apt install libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev

USAGE:
- Your save files are typically in: ~/.config/unity3d/Team Cherry/
- The app uses native GTK file dialogs
EOF

cat > dist/macos/README-macOS.txt << EOF
KnightSaver for macOS

INSTALLATION:
1. Download KnightSaver.app
2. Drag to Applications folder (optional)
3. First time: Right-click → "Open" (bypasses security check)
4. Subsequently: Double-click to run

COMPATIBILITY:
- macOS 10.13 or later
- Works on both Intel and Apple Silicon Macs

USAGE:
- Your save files are in: ~/Library/Application Support/unity.Team Cherry.*/
- The app uses native macOS file dialogs

TROUBLESHOOTING:
- If macOS blocks the app: System Preferences → Security & Privacy → Allow anyway
EOF

echo ""
echo "🎉 Build Complete!"
echo "=================="
echo ""

if [ "$DOCKER_AVAILABLE" = true ]; then
    echo "✅ All platforms built successfully!"
else
    echo "⚠️  Only macOS version built (Docker not available)"
    echo "   To build Windows/Linux: Start Docker Desktop and run again"
fi

echo ""
echo "📁 Distribution structure:"
find dist/ -type f | sort

echo ""
echo "📊 File sizes:"
find dist/ -type f -exec ls -lh {} \; | awk '{print $9 ": " $5}' | sort

echo ""
echo "🚀 Ready for GitHub Release!"
echo "   Upload the contents of the 'dist/' folder"
echo "   Suggested release files:"
echo "   - dist/macos/KnightSaver.app (compress to .zip)"
echo "   - dist/windows/* files" 
echo "   - dist/linux/* files"
echo "   - dist/README.md"

echo ""
echo "✨ Done! Check the 'dist/' folder for all release files."