package ui

import (
	"fmt"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"knightsaver/internal/backup"
	"knightsaver/internal/paths"
	"knightsaver/internal/restore"
	"knightsaver/internal/scan"
)

type App struct {
	app       fyne.App
	window    fyne.Window
	gamePaths *paths.GamePaths

	// UI components
	hkList *widget.List
	ssList *widget.List

	hkData     []*scan.SaveFile
	ssData     []*scan.SaveFile
	hkSelected map[int]bool
	ssSelected map[int]bool

	hkSelectAll     *widget.Check
	ssSelectAll     *widget.Check
	hkStatusLabel   *widget.Label
	ssStatusLabel   *widget.Label
	mainStatusLabel *widget.Label
}

func NewApp() *App {
	a := app.NewWithID("com.calo.knightsaver")

	w := a.NewWindow("KnightSaver")
	w.Resize(fyne.NewSize(800, 600))

	gamePaths, err := paths.GetGamePaths()
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to determine game paths: %w", err), w)
		return nil
	}

	appInstance := &App{
		app:        a,
		window:     w,
		gamePaths:  gamePaths,
		hkSelected: make(map[int]bool),
		ssSelected: make(map[int]bool),
	}

	appInstance.createUI()
	return appInstance
}

func (a *App) createUI() {
	// Create toolbar buttons
	scanButton := widget.NewButton("SCAN DIRS", a.scanDirectories)
	exportButton := widget.NewButton("EXPORT BACKUP", a.exportBackup)
	restoreButton := widget.NewButton("RESTORE BACKUP", a.restoreBackup)

	// Make buttons larger and more prominent
	scanButton.Importance = widget.HighImportance
	exportButton.Importance = widget.MediumImportance
	restoreButton.Importance = widget.MediumImportance

	toolbar := container.NewHBox(
		scanButton,
		widget.NewSeparator(),
		exportButton,
		widget.NewSeparator(),
		restoreButton,
	)

	// Create main status label for footer
	a.mainStatusLabel = widget.NewLabel("Ready - Click 'SCAN DIRS' to begin")
	a.mainStatusLabel.Alignment = fyne.TextAlignCenter

	// Create panel status labels
	a.hkStatusLabel = widget.NewLabel("Click 'SCAN DIRS' to find save files")
	a.ssStatusLabel = widget.NewLabel("Click 'SCAN DIRS' to find save files")

	// Create select all checkboxes
	a.hkSelectAll = widget.NewCheck("Select All", func(checked bool) {
		a.toggleAllHK(checked)
	})
	a.ssSelectAll = widget.NewCheck("Select All", func(checked bool) {
		a.toggleAllSS(checked)
	})

	// Create Open Folder buttons
	hkOpenButton := widget.NewButton("Open Folder", func() {
		a.openFolder(a.gamePaths.HollowKnight, "Hollow Knight")
	})
	hkOpenButton.Importance = widget.LowImportance

	ssOpenButton := widget.NewButton("Open Folder", func() {
		a.openFolder(a.gamePaths.Silksong, "Silksong")
	})
	ssOpenButton.Importance = widget.LowImportance

	// Create lists with proper sizing
	a.hkList = widget.NewList(
		func() int { return len(a.hkData) },
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			label := widget.NewLabel("Template save file")
			return container.NewHBox(check, label)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			cont := item.(*fyne.Container)
			check := cont.Objects[0].(*widget.Check)
			label := cont.Objects[1].(*widget.Label)

			if id < len(a.hkData) {
				save := a.hkData[id]
				label.SetText(fmt.Sprintf("%s (%d KB, %s)", save.Name, save.SizeKB, save.ModTimeStr))

				check.OnChanged = nil // Prevent recursive calls
				check.SetChecked(a.hkSelected[id])
				check.OnChanged = func(checked bool) {
					a.hkSelected[id] = checked
					a.updateHKStatus()
					a.updateSelectAllHK()
				}
			}
		},
	)

	a.ssList = widget.NewList(
		func() int { return len(a.ssData) },
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			label := widget.NewLabel("Template save file")
			return container.NewHBox(check, label)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			cont := item.(*fyne.Container)
			check := cont.Objects[0].(*widget.Check)
			label := cont.Objects[1].(*widget.Label)

			if id < len(a.ssData) {
				save := a.ssData[id]
				label.SetText(fmt.Sprintf("%s (%d KB, %s)", save.Name, save.SizeKB, save.ModTimeStr))

				check.OnChanged = nil // Prevent recursive calls
				check.SetChecked(a.ssSelected[id])
				check.OnChanged = func(checked bool) {
					a.ssSelected[id] = checked
					a.updateSSStatus()
					a.updateSelectAllSS()
				}
			}
		},
	)

	// Create panel headers - just the game names
	hkTitle := widget.NewLabel("HOLLOW KNIGHT")
	hkTitle.Alignment = fyne.TextAlignCenter
	hkTitle.TextStyle = fyne.TextStyle{Bold: true}

	ssTitle := widget.NewLabel("SILKSONG")
	ssTitle.Alignment = fyne.TextAlignCenter
	ssTitle.TextStyle = fyne.TextStyle{Bold: true}

	// Create scrollable containers for the lists
	hkScroll := container.NewScroll(a.hkList)
	hkScroll.SetMinSize(fyne.NewSize(300, 200))

	ssScroll := container.NewScroll(a.ssList)
	ssScroll.SetMinSize(fyne.NewSize(300, 200))

	// Build left panel (Hollow Knight)
	hkPanel := container.NewBorder(
		container.NewVBox(
			hkTitle,
			widget.NewSeparator(),
			container.NewHBox(a.hkSelectAll, hkOpenButton),
		), // top
		a.hkStatusLabel, // bottom
		nil,             // left
		nil,             // right
		hkScroll,        // center
	)

	// Build right panel (Silksong)
	ssPanel := container.NewBorder(
		container.NewVBox(
			ssTitle,
			widget.NewSeparator(),
			container.NewHBox(a.ssSelectAll, ssOpenButton),
		), // top
		a.ssStatusLabel, // bottom
		nil,             // left
		nil,             // right
		ssScroll,        // center
	)

	// Create main split pane
	mainPanes := container.NewHSplit(hkPanel, ssPanel)
	mainPanes.SetOffset(0.5) // 50/50 split

	// Create footer with padding
	footer := container.NewBorder(
		nil, nil, nil, nil,
		container.NewVBox(
			widget.NewSeparator(),
			a.mainStatusLabel,
		),
	)

	// Main layout: Toolbar | Content | Footer
	mainContent := container.NewBorder(
		container.NewVBox(toolbar, widget.NewSeparator()), // top
		footer,    // bottom
		nil,       // left
		nil,       // right
		mainPanes, // center
	)

	a.window.SetContent(mainContent)
	a.window.Resize(fyne.NewSize(800, 600))
}

func (a *App) scanDirectories() {
	a.mainStatusLabel.SetText("Scanning directories...")

	// Scan Hollow Knight directory
	hkSaves, err := scan.ScanDirectory(a.gamePaths.HollowKnight)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error scanning Hollow Knight directory: %w", err), a.window)
		return
	}

	// Scan Silksong directory
	ssSaves, err := scan.ScanDirectory(a.gamePaths.Silksong)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error scanning Silksong directory: %w", err), a.window)
		return
	}

	// Update data
	a.hkData = hkSaves
	a.ssData = ssSaves

	// Reset selections
	a.hkSelected = make(map[int]bool)
	a.ssSelected = make(map[int]bool)

	// Refresh lists
	a.hkList.Refresh()
	a.ssList.Refresh()

	// Update status
	a.updateHKStatus()
	a.updateSSStatus()
	a.updateSelectAllHK()
	a.updateSelectAllSS()

	totalFiles := len(hkSaves) + len(ssSaves)
	a.mainStatusLabel.SetText(fmt.Sprintf("Scan complete - Found %d save files", totalFiles))
}

func (a *App) exportBackup() {
	// Check if any files are selected
	selectedHK := a.getSelectedHKFiles()
	selectedSS := a.getSelectedSSFiles()

	if len(selectedHK) == 0 && len(selectedSS) == 0 {
		dialog.ShowInformation("Nothing to Export", "No save files are selected. Please select some files first.", a.window)
		return
	}

	// Show folder picker
	folderDialog := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil || folder == nil {
			return // User cancelled or error
		}

		a.performExport(selectedHK, selectedSS, folder.Path())
	}, a.window)

	// Make dialog larger
	folderDialog.Resize(fyne.NewSize(800, 600))
	folderDialog.Show()
}

func (a *App) performExport(hkFiles, ssFiles []*scan.SaveFile, destFolder string) {
	totalFiles := len(hkFiles) + len(ssFiles)
	a.mainStatusLabel.SetText(fmt.Sprintf("Exporting %d files...", totalFiles))

	// Perform export in a goroutine to avoid blocking UI
	go func() {
		var allResults []backup.ExportResult
		successCount := 0

		// Export Hollow Knight files
		if len(hkFiles) > 0 {
			hkSummary := backup.ExportBackups(hkFiles, backup.HollowKnight, destFolder)
			allResults = append(allResults, hkSummary.Results...)
			successCount += hkSummary.SuccessCount
		}

		// Export Silksong files
		if len(ssFiles) > 0 {
			ssSummary := backup.ExportBackups(ssFiles, backup.Silksong, destFolder)
			allResults = append(allResults, ssSummary.Results...)
			successCount += ssSummary.SuccessCount
		}

		// Update UI on main thread
		go func() {
			a.showExportResults(allResults, successCount, totalFiles, destFolder)
			a.mainStatusLabel.SetText(fmt.Sprintf("Export complete - %d of %d files exported", successCount, totalFiles))
		}()
	}()
}

func (a *App) showExportResults(results []backup.ExportResult, successCount, totalFiles int, destFolder string) {
	message := fmt.Sprintf("Exported %d of %d files to:\n%s", successCount, totalFiles, destFolder)

	// Add failure details if any
	failures := 0
	for _, result := range results {
		if !result.Success {
			failures++
		}
	}

	if failures > 0 {
		message += fmt.Sprintf("\n\n%d files failed to export:", failures)
		for _, result := range results {
			if !result.Success {
				message += fmt.Sprintf("\n• %s: %s", result.SourceFile, result.Error)
			}
		}
	}

	if failures > 0 {
		dialog.ShowInformation("Export Complete (with errors)", message, a.window)
	} else {
		dialog.ShowInformation("Export Complete", message, a.window)
	}
}

func (a *App) restoreBackup() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return // User cancelled or error
		}
		defer reader.Close()

		// Single file selection - convert to slice for consistency
		filePaths := []string{reader.URI().Path()}
		a.performRestore(filePaths)
	}, a.window)

	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".dat"}))

	// Make dialog larger
	fileDialog.Resize(fyne.NewSize(800, 600))
	fileDialog.Show()
}

func (a *App) performRestore(filePaths []string) {
	a.mainStatusLabel.SetText("Processing restore files...")

	// Parse backup files
	restoreFiles, invalidFiles := restore.ParseBackupFiles(filePaths, a.gamePaths.HollowKnight, a.gamePaths.Silksong)

	if len(invalidFiles) > 0 {
		message := "The following files have invalid backup names and will be skipped:\n\n"
		for _, file := range invalidFiles {
			message += fmt.Sprintf("• %s\n", file)
		}
		message += "\nExpected format: hk-userN-YYYYMMDD.dat or ss-userN-YYYYMMDD.dat"

		dialog.ShowInformation("Invalid Backup Files", message, a.window)
	}

	if len(restoreFiles) == 0 {
		a.mainStatusLabel.SetText("No valid backup files to restore")
		return
	}

	a.mainStatusLabel.SetText(fmt.Sprintf("Restoring %d files...", len(restoreFiles)))

	// Perform restore in a goroutine to avoid blocking UI
	go func() {
		summary := restore.RestoreBackups(restoreFiles, func(targetFile string) restore.OverwriteDecision {
			// For now, default to overwrite to avoid dialog blocking issues
			// TODO: Implement a safer overwrite confirmation later
			return restore.Overwrite
		})

		// Update UI on main thread
		go func() {
			a.showRestoreResults(summary)
			statusText := fmt.Sprintf("Restore complete - %d successful", summary.SuccessCount)
			if summary.SkippedCount > 0 {
				statusText += fmt.Sprintf(", %d skipped", summary.SkippedCount)
			}
			if summary.FailureCount > 0 {
				statusText += fmt.Sprintf(", %d failed", summary.FailureCount)
			}
			a.mainStatusLabel.SetText(statusText)
		}()
	}()
}

func (a *App) showRestoreResults(summary *restore.RestoreSummary) {
	message := fmt.Sprintf("Restore Summary:\n• %d files processed\n• %d successful\n• %d skipped\n• %d failed",
		summary.TotalFiles, summary.SuccessCount, summary.SkippedCount, summary.FailureCount)

	// Add details for failures and skips
	hasDetails := false
	for _, result := range summary.Results {
		if !result.Success || result.Skipped {
			if !hasDetails {
				message += "\n\nDetails:"
				hasDetails = true
			}
			if result.Skipped {
				message += fmt.Sprintf("\n• Skipped %s -> %s (%s)", result.BackupFile, result.TargetFile, result.OverwriteInfo)
			} else if !result.Success {
				message += fmt.Sprintf("\n• Failed %s -> %s: %s", result.BackupFile, result.TargetFile, result.Error)
			}
		}
	}

	title := "Restore Complete"
	if summary.FailureCount > 0 {
		title = "Restore Complete (with errors)"
	}

	dialog.ShowInformation(title, message, a.window)
}

func (a *App) getSelectedHKFiles() []*scan.SaveFile {
	var selected []*scan.SaveFile
	for i, isSelected := range a.hkSelected {
		if isSelected && i < len(a.hkData) {
			selected = append(selected, a.hkData[i])
		}
	}
	return selected
}

func (a *App) getSelectedSSFiles() []*scan.SaveFile {
	var selected []*scan.SaveFile
	for i, isSelected := range a.ssSelected {
		if isSelected && i < len(a.ssData) {
			selected = append(selected, a.ssData[i])
		}
	}
	return selected
}

func (a *App) toggleAllHK(checked bool) {
	for i := range a.hkData {
		a.hkSelected[i] = checked
	}
	a.hkList.Refresh()
	a.updateHKStatus()
}

func (a *App) toggleAllSS(checked bool) {
	for i := range a.ssData {
		a.ssSelected[i] = checked
	}
	a.ssList.Refresh()
	a.updateSSStatus()
}

func (a *App) updateHKStatus() {
	selectedCount := 0
	for _, isSelected := range a.hkSelected {
		if isSelected {
			selectedCount++
		}
	}

	if len(a.hkData) == 0 {
		if paths.DirExists(a.gamePaths.HollowKnight) {
			a.hkStatusLabel.SetText("No save files found")
		} else {
			a.hkStatusLabel.SetText("Save directory not found")
		}
	} else {
		a.hkStatusLabel.SetText(fmt.Sprintf("%d files, %d selected", len(a.hkData), selectedCount))
	}
}

func (a *App) updateSSStatus() {
	selectedCount := 0
	for _, isSelected := range a.ssSelected {
		if isSelected {
			selectedCount++
		}
	}

	if len(a.ssData) == 0 {
		if paths.DirExists(a.gamePaths.Silksong) {
			a.ssStatusLabel.SetText("No save files found")
		} else {
			a.ssStatusLabel.SetText("Save directory not found")
		}
	} else {
		a.ssStatusLabel.SetText(fmt.Sprintf("%d files, %d selected", len(a.ssData), selectedCount))
	}
}

func (a *App) updateSelectAllHK() {
	if len(a.hkData) == 0 {
		a.hkSelectAll.OnChanged = nil
		a.hkSelectAll.SetChecked(false)
		a.hkSelectAll.OnChanged = func(checked bool) { a.toggleAllHK(checked) }
		return
	}

	selectedCount := 0
	for _, isSelected := range a.hkSelected {
		if isSelected {
			selectedCount++
		}
	}

	a.hkSelectAll.OnChanged = nil
	a.hkSelectAll.SetChecked(selectedCount == len(a.hkData))
	a.hkSelectAll.OnChanged = func(checked bool) { a.toggleAllHK(checked) }
}

func (a *App) updateSelectAllSS() {
	if len(a.ssData) == 0 {
		a.ssSelectAll.OnChanged = nil
		a.ssSelectAll.SetChecked(false)
		a.ssSelectAll.OnChanged = func(checked bool) { a.toggleAllSS(checked) }
		return
	}

	selectedCount := 0
	for _, isSelected := range a.ssSelected {
		if isSelected {
			selectedCount++
		}
	}

	a.ssSelectAll.OnChanged = nil
	a.ssSelectAll.SetChecked(selectedCount == len(a.ssData))
	a.ssSelectAll.OnChanged = func(checked bool) { a.toggleAllSS(checked) }
}

func (a *App) Run() {
	a.window.ShowAndRun()
}

func (a *App) openFolder(folderPath, gameName string) {
	// Check if folder exists
	if !paths.DirExists(folderPath) {
		dialog.ShowInformation("Folder Not Found",
			fmt.Sprintf("The %s save directory does not exist:\n\n%s\n\nThe game may not have been run yet or save files haven't been created.", gameName, folderPath),
			a.window)
		return
	}

	// Open folder using OS-specific command
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", folderPath)
	case "windows":
		cmd = exec.Command("explorer", folderPath)
	case "linux":
		// Try common Linux file managers
		cmd = exec.Command("xdg-open", folderPath)
	default:
		dialog.ShowInformation("Unsupported Platform",
			fmt.Sprintf("Cannot open folder on %s platform.\n\nManually navigate to:\n%s", runtime.GOOS, folderPath),
			a.window)
		return
	}

	// Execute command
	if err := cmd.Start(); err != nil {
		dialog.ShowError(fmt.Errorf("Failed to open folder: %v\n\nPath: %s", err, folderPath), a.window)
	} else {
		a.mainStatusLabel.SetText(fmt.Sprintf("Opened %s save folder", gameName))
	}
}

func (a *App) Close() {
	a.app.Quit()
}
