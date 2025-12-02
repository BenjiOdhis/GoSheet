// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// indel.go contains functions for inserting and deleting rows/columns.

package table

import (
	"fmt"
	"gosheet/internal/services/ui"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"

	"github.com/rivo/tview"
)

var totalRows, totalCols int32

// DELETE FUNCTIONS
func eliminateRowCol(app *tview.Application, table *tview.Table) {
	activeViewport := GetActiveViewport()
	
	if activeViewport == nil {
		return
	}

	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := activeViewport.ToAbsolute(visualRow, visualCol)

	if row == 0 && col > 0 {
		eliminateCol(app, table, col)
	} else if col == 0 && row > 0 {
		eliminateRow(app, table, row)
	} else {
		ui.ShowWarningModal(app, table, "Select a column or a row before deleting it.")
	}
}

func eliminateCol(app *tview.Application, table *tview.Table, col int32) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you wish to delete col %s?", utils.ColumnName(col))).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				activeData := GetActiveSheetData()
				activeViewport := GetActiveViewport()
				
				if activeData == nil || activeViewport == nil {
					app.SetRoot(table, true).SetFocus(table)
					return
				}
				
				newData := make(map[[2]int]*cell.Cell)
				for key, cellData := range activeData {
					r, c := key[0], key[1]
					if int32(c) < col {
						newData[key] = cellData
					} else if int32(c) > col {
						newKey := [2]int{r, c - 1}
						cellData.Column--
						newData[newKey] = cellData
					}
				}
				
				// CRITICAL: Update the actual sheet's data reference
				sheet := globalWorkbook.GetActiveSheet()
				if sheet != nil {
					sheet.Data = newData
				}
				
				RenderVisible(table, activeViewport, newData)
				app.SetRoot(table, true).SetFocus(table)
			} else {
				app.SetRoot(table, true).SetFocus(table)
			}
		})
	modal.SetBorder(true).SetTitle(" Confirmation ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)	
}

func eliminateRow(app *tview.Application, table *tview.Table, row int32) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you wish to delete row %d?", row)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				activeData := GetActiveSheetData()
				activeViewport := GetActiveViewport()
				
				if activeData == nil || activeViewport == nil {
					app.SetRoot(table, true).SetFocus(table)
					return
				}
				
				newData := make(map[[2]int]*cell.Cell)
				for key, cellData := range activeData {
					r, c := key[0], key[1]
					if int32(r) < row {
						newData[key] = cellData
					} else if int32(r) > row {
						newKey := [2]int{r - 1, c}
						cellData.Row--
						newData[newKey] = cellData
					}
				}
				
				// CRITICAL: Update the actual sheet's data reference
				sheet := globalWorkbook.GetActiveSheet()
				if sheet != nil {
					sheet.Data = newData
				}
				
				RenderVisible(table, activeViewport, newData)
				app.SetRoot(table, true).SetFocus(table)
			} else {
				app.SetRoot(table, true).SetFocus(table)
			}
		})
	modal.SetBorder(true).SetTitle(" Confirmation ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)	
}

// INSERT FUNCTIONS
func insertRowCol(app *tview.Application, table *tview.Table) {
	activeData := GetActiveSheetData()
	activeViewport := GetActiveViewport()
	
	if activeData == nil || activeViewport == nil {
		return
	}

	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := activeViewport.ToAbsolute(visualRow, visualCol)

	if row == 0 && col > 0 {
		insertCol(app, table, col)
	} else if col == 0 && row > 0 {
		insertRow(app, table, row)
	} else {
		ui.ShowWarningModal(app, table, "Select a column or row header to insert before it.")
	}
}

func insertCol(app *tview.Application, table *tview.Table, col int32) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Insert a new column before %s?", utils.ColumnName(col))).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				activeData := GetActiveSheetData()
				activeViewport := GetActiveViewport()
				
				if activeData == nil || activeViewport == nil {
					app.SetRoot(table, true).SetFocus(table)
					return
				}
				
				newData := make(map[[2]int]*cell.Cell)
				for key, cellData := range activeData {
					r, c := key[0], key[1]
					if int32(c) < col {
						newData[key] = cellData
					} else {
						newKey := [2]int{r, c + 1}
						cellData.Column++
						newData[newKey] = cellData
					}
				}
				
				// CRITICAL: Update the actual sheet's data reference
				sheet := globalWorkbook.GetActiveSheet()
				if sheet != nil {
					sheet.Data = newData
				}
				
				RenderVisible(table, activeViewport, newData)
				app.SetRoot(table, true).SetFocus(table)
			} else {
				app.SetRoot(table, true).SetFocus(table)
			}
		})
	modal.SetBorder(true).SetTitle(" Insert Column ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)	
}

func insertRow(app *tview.Application, table *tview.Table, row int32) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Insert a new row before row %d?", row)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				activeData := GetActiveSheetData()
				activeViewport := GetActiveViewport()
				
				if activeData == nil || activeViewport == nil {
					app.SetRoot(table, true).SetFocus(table)
					return
				}
				
				newData := make(map[[2]int]*cell.Cell)
				for key, cellData := range activeData {
					r, c := key[0], key[1]
					if int32(r) < row {
						newData[key] = cellData
					} else {
						newKey := [2]int{r + 1, c}
						cellData.Row++
						newData[newKey] = cellData
					}
				}
				
				// CRITICAL: Update the actual sheet's data reference
				sheet := globalWorkbook.GetActiveSheet()
				if sheet != nil {
					sheet.Data = newData
				}
				
				RenderVisible(table, activeViewport, newData)
				app.SetRoot(table, true).SetFocus(table)
			} else {
				app.SetRoot(table, true).SetFocus(table)
			}
		})
	modal.SetBorder(true).SetTitle(" Insert Row ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)	
}
