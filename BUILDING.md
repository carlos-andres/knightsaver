# Building KnightSaver

This document explains how to build KnightSaver for all platforms.

## Quick Start

### Option 1: Automated Build Script (Recommended)
```bash
# Make sure Docker Desktop is running first!
./build-release.sh 1.0.0
```

### Option 2: Manual Commands
```bash
# macOS (no Docker needed)
fyne package --target darwin --icon assets/icon.png --name KnightSaver --app-id com.calo.knightsaver --src ./cmd/knightsaver

# Windows (requires Docker)  
fyne-cross windows --icon assets/icon.png --app-id com.calo.knightsaver

# Linux (requires Docker)
fyne-cross linux --icon assets/icon.png --app-id com.calo.knightsaver
```

## Prerequisites

### Required for All Builds
- Go 1.22 or later
- Fyne CLI tools: `go install fyne.io/tools/cmd/fyne@latest`

### Required for Windows/Linux Builds
- Docker Desktop (running)
- fyne-cross: `go install github.com/fyne-io/fyne-cross@latest`

## Platform-Specific Notes

### macOS
- Builds natively on macOS without Docker
- Creates `.app` bundle in current directory
- App is unsigned - users need to right-click → Open first time

### Windows
- Requires Docker for cross-compilation
- Creates `.exe` files and `.tar.xz` archives
- Supports both x64 and ARM64 architectures

### Linux
- Requires Docker for cross-compilation
- Creates binary executables and `.tar.xz` archives
- Supports both x64 and ARM64 architectures

## GitHub Actions

The repository includes a GitHub Actions workflow (`.github/workflows/release.yml`) that automatically builds for all platforms when you:

1. Push a tag like `v1.0.0`
2. Or manually trigger the workflow with a version number

## Distribution Structure

After building, you'll have:

```
dist/
├── README.md                    # Main distribution README
├── macos/
│   ├── KnightSaver.app/        # macOS application bundle
│   ├── KnightSaver-macOS.zip   # Compressed app for easy download
│   └── README-macOS.txt        # macOS-specific instructions
├── windows/
│   ├── knightsaver.exe         # Windows executable
│   ├── *.tar.xz               # Additional Windows archives
│   └── README-Windows.txt      # Windows-specific instructions
└── linux/
    ├── knightsaver             # Linux executable
    ├── *.tar.xz               # Additional Linux archives
    └── README-Linux.txt        # Linux-specific instructions
```

## Troubleshooting

### Docker Issues
- **Error**: "Cannot connect to the Docker daemon"
- **Solution**: Start Docker Desktop and wait for it to fully initialize

### Build Failures
- **Error**: "Missing application icon"
- **Solution**: Use absolute path to icon: `--icon /full/path/to/assets/icon.png`

### macOS Security
- **Issue**: "App is damaged and can't be opened"
- **Solution**: App is unsigned. Users should right-click → Open first time

### Linux Dependencies
- **Issue**: Missing GUI libraries
- **Solution**: Install: `sudo apt install libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev`

## Release Checklist

When preparing a GitHub release:

1. ✅ Test the app on macOS
2. ✅ Run `./build-release.sh` with Docker running  
3. ✅ Verify all platform builds in `dist/` folder
4. ✅ Create GitHub release with tag (e.g., `v1.0.0`)
5. ✅ Upload files from `dist/` folder to release assets
6. ✅ Include distribution README.md in release description

## File Sizes

Typical build sizes:
- **macOS**: ~30MB (.app bundle)
- **Windows**: ~25MB (.exe)
- **Linux**: ~25MB (binary)

The apps are statically linked and include all dependencies.