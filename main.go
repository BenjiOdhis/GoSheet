// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// A simple terminal-based spreadsheet application using tview.
// main.go is the entry point of the application.

package main

import (
	"fmt"
	"gosheet/internal/services/table"
	"gosheet/internal/services/ui"
	"gosheet/internal/utils"
	"os"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"flag"
)

// main is the entry point of the application, where the tview application is initialized, it is checked for command-line arguments to open a file or create a new table
func main() {
	//runtime.MemProfileRate = 1

	utils.UpdateNrCellsOnScrn()

	app := tview.NewApplication()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			fmt.Println("\nCtrl+C detected. Exiting gracefully...\nUnsaved edits may be lost.")
			return nil
		}
		return event
	})

	defer func() {
	    if r := recover(); r != nil {
	        fmt.Fprintf(os.Stderr, "Application crashed: %v\n", r)
	        fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
	        os.Exit(1)
	    }
	}()
	
	var filename string

	flag.StringVar(&filename, "file", "", "Path to .gsheet or .json file to open")
	flag.Parse()

	var t *tview.Table
	if filename != "" {
		ui.AddToRecentFiles(filename)
		t = table.OpenTable(app, filename)
	} else {
		filename = ui.StartMenuUI(app)

		if filename == "QUIT" {
        	return
    	}

		if filename == "THERE_IS_NO_FILE_SELECTED" {
			t = table.NewTable(app, 1000, 702)
		} else {
			ui.AddToRecentFiles(filename)
			t = table.OpenTable(app, filename)
		}
    }

	t.Select(1, 1)

	app.SetRoot(t, true).SetFocus(t)

	if err := app.Run(); err != nil {
		panic(err)
	}

	//fmt.Println(utils.TOBEPRINTED)

	/*
    // Print memory stats
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("\nMemory Stats:\n")
    fmt.Printf("Alloc = %v MB\n", m.Alloc/1024/1024)
    fmt.Printf("TotalAlloc = %v MB\n", m.TotalAlloc/1024/1024)
    fmt.Printf("Sys = %v MB\n", m.Sys/1024/1024)
    fmt.Printf("NumGC = %v\n", m.NumGC)
	*/
}
