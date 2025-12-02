// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// sheetmanagerUI.go provides UI components for sheet management

package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SheetManagerCallbacks defines callbacks for sheet operations
type SheetManagerCallbacks struct {
	GetSheets          func() []SheetInfo
	GetActiveIndex     func() int
	GetWorkbookInfo    func() WorkbookInfo
	AddSheet           func(name string) error
	RenameSheet        func(index int, name string) error
	DeleteSheet        func(index int) error
	DuplicateSheet     func(index int) error
	MoveSheet          func(fromIndex, toIndex int) error
	SwitchToSheet      func(index int) error
	UpdateTabBar       func()
	UpdateTableTitle   func()
	MarkAsModified     func()
	RenderActiveSheet  func()
}

// SheetInfo contains information about a single sheet
type SheetInfo struct {
	Name      string
	CellCount int
	IsActive  bool
}

// WorkbookInfo contains information about the workbook
type WorkbookInfo struct {
	TotalSheets int
	ActiveSheet string
	TotalCells  int
	FileName    string
	HasChanges  bool
}

// ShowSheetManager displays the comprehensive sheet management dialog
func ShowSheetManager(app *tview.Application, table *tview.Table, callbacks SheetManagerCallbacks) {
	// Sheet list with enhanced styling
	list := tview.NewList().
		SetSelectedBackgroundColor(tcell.ColorDarkCyan).
		SetSelectedTextColor(tcell.ColorWhite).
		SetMainTextColor(tcell.ColorWhite).
		SetSecondaryTextColor(tcell.ColorGray).
		ShowSecondaryText(true)
	list.SetBorder(true).
		SetTitle(" Sheets ").
		SetBorderColor(tcell.ColorLightBlue).
		SetTitleAlign(tview.AlignLeft)

	updateSheetList(list, callbacks)

	// Info panel with better formatting
	infoPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true).
		SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
	infoPanel.SetBorder(true).
		SetTitle(" Details ").
		SetBorderColor(tcell.ColorLightBlue).
		SetTitleAlign(tview.AlignLeft)

	// Layout assembly
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 3, true).
		AddItem(infoPanel, 12, 0, false)

	mainContent := tview.NewFlex()
	
	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainContent, 0, 1, true)

	mainLayout.SetBorder(true).
		SetTitle(" Sheet Manager ").
		SetBorderColor(tcell.ColorYellow).
		SetTitleAlign(tview.AlignCenter)

	actionPanel := createActionPanel(app, table, callbacks, list, infoPanel)

	mainContent.
		AddItem(leftPanel, 0, 2, true).
		AddItem(actionPanel, 45, 0, false)

	mainLayout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.SetRoot(table, true).SetFocus(table)
			return nil
		case tcell.KeyEnter:
			switchToSelectedSheet(app, table, callbacks, list)
			return nil
		}

		if event.Modifiers()&tcell.ModAlt != 0 {
			switch event.Rune() {
			case 'n', 'N':
				showAddSheetDialog(app, callbacks, mainLayout, list, infoPanel)
				return nil
			case 'r', 'R':
				showRenameSheetFromManager(app, callbacks, mainLayout, list, infoPanel)
				return nil
			case 'd', 'D':
				confirmDeleteSheetFromManager(app, callbacks, mainLayout, list, infoPanel)
				return nil
			case 'm', 'M':
				showMoveSheetDialog(app, callbacks, mainLayout, list, infoPanel)
				return nil
			case 'c', 'C':
				duplicateSheetFromManager(app, table, callbacks, list, infoPanel)
				return nil
			case 's', 'S':
				switchToSelectedSheet(app, table, callbacks, list)
				return nil
			}
		}

		return event
	})

	// Update info panel when selection changes
	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		sheets := callbacks.GetSheets()
		if index >= 0 && index < len(sheets) {
			infoPanel.SetText(getSheetInfoText(sheets[index], index, len(sheets)))
		}
	})

	app.SetRoot(mainLayout, true).SetFocus(list)
}

// createActionPanel creates an enhanced button panel with icons and descriptions
func createActionPanel(app *tview.Application, table *tview.Table,
	callbacks SheetManagerCallbacks, list *tview.List, infoPanel *tview.TextView) *tview.Flex {
	
	actionPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	actionPanel.SetBorder(true).
		SetTitle(" ⚡Actions ").
		SetBorderColor(tcell.ColorLightBlue).
		SetTitleAlign(tview.AlignLeft)

	createActionBtn := func(icon, label, shortcut, description string, color tcell.Color, action func()) *tview.Box {
		btn := tview.NewBox().
			SetBorder(true).
			SetBorderColor(color)
		
		btn.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			tview.Print(screen, fmt.Sprintf(" %s %s", icon, label), x+1, y+1, width-2, tview.AlignLeft, color)
			
			tview.Print(screen, shortcut, x+width-len(shortcut)-2, y+1, len(shortcut), tview.AlignRight, tcell.ColorYellow)
			
			if len(description) > 0 {
				tview.Print(screen, description, x+1, y+2, width-2, tview.AlignLeft, tcell.ColorGray)
			}
			
			return x + 1, y + 3, width - 2, height - 3
		})

		btn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEnter {
				action()
				return nil
			}
			return event
		})

		return btn
	}

	// Action buttons
	newSheetBtn := createActionBtn(
		"", "New Sheet", "Alt+N",
		"Create a blank sheet",
		tcell.ColorGreen,
		func() { showAddSheetDialog(app, callbacks, actionPanel, list, infoPanel) },
	)

	renameBtn := createActionBtn(
		"", "Rename", "Alt+R",
		"Change sheet name",
		tcell.ColorBlue,
		func() { showRenameSheetFromManager(app, callbacks, actionPanel, list, infoPanel) },
	)

	duplicateBtn := createActionBtn(
		"", "Duplicate", "Alt+C",
		"Copy entire sheet",
		tcell.ColorLightBlue,
		func() { duplicateSheetFromManager(app, table, callbacks, list, infoPanel) },
	)

	moveBtn := createActionBtn(
		"", "Move/Reorder", "Alt+M",
		"Change sheet position",
		tcell.ColorYellow,
		func() { showMoveSheetDialog(app, callbacks, actionPanel, list, infoPanel) },
	)

	deleteBtn := createActionBtn(
		"", "Delete", "Alt+D",
		"Remove sheet permanently",
		tcell.ColorRed,
		func() { confirmDeleteSheetFromManager(app, callbacks, actionPanel, list, infoPanel) },
	)

	switchBtn := createActionBtn(
		"", "Switch To", "Enter",
		"Open selected sheet",
		tcell.ColorLightBlue,
		func() { switchToSelectedSheet(app, table, callbacks, list) },
	)

	exitBtn := createActionBtn(
		"", "Exit", "Esc",
		"Exit sheet manager",
		tcell.ColorLightBlue,
		func() { app.SetRoot(table, true).SetFocus(table) },
	)

	actionPanel.
		AddItem(newSheetBtn, 4, 0, false).
		AddItem(renameBtn, 4, 0, false).
		AddItem(duplicateBtn, 4, 0, false).
		AddItem(moveBtn, 4, 0, false).
		AddItem(deleteBtn, 4, 0, false).
		AddItem(switchBtn, 4, 0, false).
		AddItem(exitBtn, 4, 0, false).
		AddItem(tview.NewBox(), 0, 1, false)

	return actionPanel
}

// updateSheetList refreshes the sheet list with enhanced display
func updateSheetList(list *tview.List, callbacks SheetManagerCallbacks) {
	list.Clear()
	sheets := callbacks.GetSheets()

	for i, sheet := range sheets {
		icon := ""
		badge := ""
		
		if sheet.IsActive {
			icon = ""
			badge = " [yellow::b]> ACTIVE[::-]"
		}

		mainText := fmt.Sprintf(" %s  %s%s", icon, sheet.Name, badge)
		secondaryText := fmt.Sprintf("   └─ %d cells with data", sheet.CellCount)

		list.AddItem(
			mainText,
			secondaryText,
			0,
			func(idx int) func() {
				return func() {
					callbacks.SwitchToSheet(idx)
					callbacks.UpdateTabBar()
					callbacks.UpdateTableTitle()
					callbacks.RenderActiveSheet()
				}
			}(i),
		)
	}
}

// getWorkbookInfoText returns enhanced workbook information
func getWorkbookInfoText(info WorkbookInfo) string {
	fileName := info.FileName
	if fileName == "" {
		fileName = "[gray]Untitled Workbook[-]"
	} else {
		fileName = "[white]" + fileName + "[-]"
	}

	statusIcon := ""
	statusColor := "green"
	statusText := "Saved"
	
	if info.HasChanges {
		statusIcon = "o"
		statusColor = "yellow"
		statusText = "Modified"
	}

	return fmt.Sprintf(
		"[::b]WORKBOOK OVERVIEW[::-]\n"+
			"[gray]━━━━━━━━━━━━━━━━━━━━[-]\n"+
			"[lightblue]File:[-]\n  %s\n"+
			"[lightblue]Status:[-]  [%s]%s %s[-]\n"+
			"[lightblue]Structure:[-]\n"+
			"  • Sheets: [white]%d[-]\n"+
			"  • Active: [white]%s[-]\n"+
			"  • Total Cells: [white]%d[-]",
		fileName,
		statusColor, statusIcon, statusText,
		info.TotalSheets,
		info.ActiveSheet,
		info.TotalCells,
	)
}

// getSheetInfoText returns enhanced info for a specific sheet
func getSheetInfoText(sheet SheetInfo, index, total int) string {
	statusIcon := "o"
	statusColor := "gray"
	statusText := "Inactive"
	
	if sheet.IsActive {
		statusIcon = ">"
		statusColor = "yellow"
		statusText = "Active Sheet"
	}

	return fmt.Sprintf(
		"[::b]SHEET DETAILS[::-]\n"+
			"[gray]━━━━━━━━━━━━━━━━━━━━[-]\n"+
			"[lightblue]Name:[-]  [white::b]%s[::-]\n"+
			"[lightblue]Status:[-]  [%s]%s %s[-]\n"+
			"[lightblue]Content:[-]\n"+
			"  • Cells: [white]%d[-]\n"+
			"  • Position: [white]%d[-] of [white]%d[-]",
		sheet.Name,
		statusColor, statusIcon, statusText,
		sheet.CellCount,
		index+1,
		total,
	)
}

// showAddSheetDialog shows enhanced dialog to add a new sheet
func showAddSheetDialog(app *tview.Application, 
	callbacks SheetManagerCallbacks, returnTo tview.Primitive, list *tview.List, infoPanel *tview.TextView) {
	
	form := tview.NewForm()
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorDarkGreen)
	form.SetButtonTextColor(tcell.ColorWhite)
	
	sheets := callbacks.GetSheets()
	defaultName := fmt.Sprintf("Sheet%d", len(sheets)+1)
	
	nameInput := tview.NewInputField().
		SetLabel("Sheet Name: ").
		SetText(defaultName).
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

	form.AddFormItem(nameInput).
		AddButton("Create", func() {
			name := strings.TrimSpace(nameInput.GetText())
			if name == "" {
				ShowWarningModal(app, form, "Sheet name cannot be empty!")
				return
			}

			if err := callbacks.AddSheet(name); err != nil {
				ShowWarningModal(app, form, err.Error())
				return
			}

			updateSheetList(list, callbacks)
			callbacks.UpdateTabBar()
			infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
			callbacks.MarkAsModified()

			app.SetRoot(returnTo, true).SetFocus(list)
		}).
		AddButton("Cancel", func() {
			app.SetRoot(returnTo, true).SetFocus(list)
		})

	form.SetBorder(true).
		SetTitle(" + Add New Sheet ").
		SetBorderColor(tcell.ColorGreen).
		SetTitleAlign(tview.AlignCenter)

	app.SetRoot(form, true).SetFocus(form)
}

// showRenameSheetFromManager shows enhanced rename dialog
func showRenameSheetFromManager(app *tview.Application,
	callbacks SheetManagerCallbacks, returnTo tview.Primitive, list *tview.List, infoPanel *tview.TextView) {
	
	selectedIndex := list.GetCurrentItem()
	sheets := callbacks.GetSheets()
	
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	sheet := sheets[selectedIndex]

	form := tview.NewForm()
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorDarkBlue)
	form.SetButtonTextColor(tcell.ColorWhite)
	
	nameInput := tview.NewInputField().
		SetLabel(" New Name: ").
		SetText(sheet.Name).
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

	form.AddFormItem(nameInput).
		AddButton("Rename", func() {
			newName := strings.TrimSpace(nameInput.GetText())
			if newName == "" {
				ShowWarningModal(app, form, "Sheet name cannot be empty!")
				return
			}

			if err := callbacks.RenameSheet(selectedIndex, newName); err != nil {
				ShowWarningModal(app, form, err.Error())
				return
			}

			updateSheetList(list, callbacks)
			callbacks.UpdateTabBar()
			callbacks.UpdateTableTitle()
			infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
			callbacks.MarkAsModified()

			app.SetRoot(returnTo, true).SetFocus(list)
		}).
		AddButton("✗ Cancel", func() {
			app.SetRoot(returnTo, true).SetFocus(list)
		})

	form.SetBorder(true).
		SetTitle( " Rename Sheet ").
		SetBorderColor(tcell.ColorBlue).
		SetTitleAlign(tview.AlignCenter)

	app.SetRoot(form, true).SetFocus(form)
}

// confirmDeleteSheetFromManager shows enhanced deletion confirmation
func confirmDeleteSheetFromManager(app *tview.Application,
	callbacks SheetManagerCallbacks, returnTo tview.Primitive, list *tview.List, infoPanel *tview.TextView) {
	
	sheets := callbacks.GetSheets()
	if len(sheets) <= 1 {
		ShowWarningModal(app, returnTo, "⚠️  Cannot delete the last sheet!\n\nA workbook must have at least one sheet.")
		return
	}

	selectedIndex := list.GetCurrentItem()
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	sheet := sheets[selectedIndex]

	modal := tview.NewModal().
		SetText(fmt.Sprintf(
			"[red::b]⚠️  DELETE SHEET[::-]\n\n"+
				"Are you sure you want to delete:\n"+
				"[yellow]'%s'[-]?\n\n"+
				"[white]This will permanently remove:[-]\n"+
				"  • [white]%d[-] cells with data\n"+
				"  • All formulas and formatting\n"+
				"  • All undo/redo history\n\n"+
				"[red::b]⚠️  This action cannot be undone![::-]",
			sheet.Name,
			sheet.CellCount,
		)).
		AddButtons([]string{"X Delete", "x Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if strings.Contains(buttonLabel, "Delete") {
				if err := callbacks.DeleteSheet(selectedIndex); err != nil {
					ShowWarningModal(app, returnTo, err.Error())
					app.SetRoot(returnTo, true).SetFocus(list)
					return
				}

				updateSheetList(list, callbacks)
				callbacks.UpdateTabBar()
				callbacks.RenderActiveSheet()
				callbacks.UpdateTableTitle()
				infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
				callbacks.MarkAsModified()
			}
			app.SetRoot(returnTo, true).SetFocus(list)
		})

	modal.SetBackgroundColor(tcell.ColorDarkRed).
		SetBorderColor(tcell.ColorRed)

	app.SetRoot(modal, true).SetFocus(modal)
}

// duplicateSheetFromManager duplicates with visual feedback
func duplicateSheetFromManager(app *tview.Application, table *tview.Table,
	callbacks SheetManagerCallbacks, list *tview.List, infoPanel *tview.TextView) {
	
	selectedIndex := list.GetCurrentItem()
	sheets := callbacks.GetSheets()
	
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	if err := callbacks.DuplicateSheet(selectedIndex); err != nil {
		ShowWarningModal(app, table, "X "+err.Error())
		return
	}

	updateSheetList(list, callbacks)
	callbacks.UpdateTabBar()
	infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
	callbacks.MarkAsModified()

	list.SetCurrentItem(len(sheets))
}

// showMoveSheetDialog shows enhanced reorder dialog
func showMoveSheetDialog(app *tview.Application,
	callbacks SheetManagerCallbacks, returnTo tview.Primitive, list *tview.List, infoPanel *tview.TextView) {
	
	selectedIndex := list.GetCurrentItem()
	sheets := callbacks.GetSheets()
	
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	form := tview.NewForm()
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorDarkGoldenrod)
	form.SetButtonTextColor(tcell.ColorWhite)
	
	positions := make([]string, len(sheets))
	for i := range sheets {
		if i == selectedIndex {
			positions[i] = fmt.Sprintf("Position %d (current)", i+1)
		} else {
			positions[i] = fmt.Sprintf("Position %d", i+1)
		}
	}

	form.AddDropDown("Move to:", positions, selectedIndex, nil).
		AddButton("Move", func() {
			newPos, _ := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
			
			if newPos == selectedIndex {
				app.SetRoot(returnTo, true).SetFocus(list)
				return
			}

			if err := callbacks.MoveSheet(selectedIndex, newPos); err != nil {
				ShowWarningModal(app, form, err.Error())
				return
			}

			updateSheetList(list, callbacks)
			callbacks.UpdateTabBar()
			infoPanel.SetText(getWorkbookInfoText(callbacks.GetWorkbookInfo()))
			callbacks.MarkAsModified()

			list.SetCurrentItem(newPos)
			app.SetRoot(returnTo, true).SetFocus(list)
		}).
		AddButton("Cancel", func() {
			app.SetRoot(returnTo, true).SetFocus(list)
		})

	form.SetBorder(true).
		SetTitle(" Move/Reorder Sheet ").
		SetBorderColor(tcell.ColorYellow).
		SetTitleAlign(tview.AlignCenter)

	app.SetRoot(form, true).SetFocus(form)
}

// switchToSelectedSheet switches with smooth feedback
func switchToSelectedSheet(app *tview.Application, table *tview.Table, 
	callbacks SheetManagerCallbacks, list *tview.List) {
	
	selectedIndex := list.GetCurrentItem()
	sheets := callbacks.GetSheets()
	
	if selectedIndex < 0 || selectedIndex >= len(sheets) {
		return
	}

	if err := callbacks.SwitchToSheet(selectedIndex); err != nil {
		return
	}

	callbacks.UpdateTabBar()
	callbacks.UpdateTableTitle()
	callbacks.RenderActiveSheet()
	
	app.SetRoot(table, true).SetFocus(table)
}
