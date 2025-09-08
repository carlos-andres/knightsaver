package main

import (
	"knightsaver/internal/ui"
)

func main() {
	app := ui.NewApp()
	if app != nil {
		app.Run()
	}
}