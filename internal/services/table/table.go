// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// table.go provides functions to create tview.Table instances for a spreadsheet application.

package table

import (
	"fmt"
	"path/filepath"

	"gosheet/internal/services/cell"
	"gosheet/internal/services/fileop"
	"gosheet/internal/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)


var globalWorkbook *Workbook

// Creates an empty tview table
func CreateTable(title string) *tview.Table {
    table := tview.NewTable().
        SetBorders(false).
        SetFixed(1, 1).
        SetSelectable(true, true)
	table.SetBorder(true)

	SetCurrentFilename(table, title)	
	updateTableTitle(table)

    return table
}

// OpenTable loads a table from a file and returns a tview.Table
func OpenTable(app *tview.Application, filename string) *tview.Table {
	cellSlice, err := fileop.OpenTable(filename)

	if err != nil {
		return nil
	}

	// Initialize workbook
	globalWorkbook = NewWorkbook()
	globalWorkbook.CurrentFile = filename
	globalWorkbook.HasChanges = false

	// Get the first sheet
	sheet := globalWorkbook.GetActiveSheet()
	
	// Populate sheet data
	for _, c := range cellSlice {
		key := [2]int{int(c.Row), int(c.Column)}
		sheet.Data[key] = c
	}

	table := CreateTable(filename)

	RenderVisible(table, sheet.Viewport, sheet.Data)
	table = SelectInTable(app, table, sheet.Viewport, sheet.Data)

	return table
}

// Makes a new table/Workbook
func NewTable(app *tview.Application) *tview.Table {
	globalWorkbook = NewWorkbook()
	globalWorkbook.CurrentFile = ""
	globalWorkbook.HasChanges = false
	
	table := CreateTable("Untitled")
	
	sheet := globalWorkbook.GetActiveSheet()
	RenderVisible(table, sheet.Viewport, sheet.Data)
	table = SelectInTable(app, table, sheet.Viewport, sheet.Data)
	
	updateTableTitle(table)
	
	return table
}

// Cleanup unused cells from memory
func CleanupDistantCells(data map[[2]int]*cell.Cell, vp *utils.Viewport, keepDistance int32) {
	minRow := max(1, vp.TopRow - keepDistance)
    maxRow := vp.TopRow + vp.ViewRows + keepDistance
    minCol := max(1, vp.LeftCol - keepDistance)
    maxCol := vp.LeftCol + vp.ViewCols + keepDistance

	for key, cellData := range data {
		row, col := int32(key[0]), int32(key[1])
		
		if row < minRow || row > maxRow || col < minCol || col > maxCol {
			if isEmptyCell(cellData) {
				delete(data, key)
			}
		}
	}
}

// Checks if cell is empty
func isEmptyCell(c *cell.Cell) bool {
	if c.RawValue != nil && *c.RawValue != "" {
		return false
	}
	if c.IsFormula() {
		return false
	}
	if c.Notes != nil && *c.Notes != "" {
		return false
	}
	if c.HasFlag(cell.FlagBold) || c.HasFlag(cell.FlagItalic) || c.HasFlag(cell.FlagUnderline) {
		return false
	}
	if c.Valrule != nil && *c.Valrule != "" {
		return false
	}
	if c.Color[0] != 255 || c.Color[1] != 255 || c.Color[2] != 255 {
        return false
    }
    if c.BgColor[0] != 0 || c.BgColor[1] != 0 || c.BgColor[2] != 0 {
        return false
    }	
	
	return true
}

// Render Table Viewport for optimised memory usage
func RenderVisible(table *tview.Table, vp *utils.Viewport, data map[[2]int]*cell.Cell) {
	table.Clear()

	table.SetCell(0, 0, tview.NewTableCell("").SetAlign(tview.AlignCenter))

	for c := vp.LeftCol; c < vp.LeftCol+vp.ViewCols; c++ {
		label := utils.ColumnName(int32(c))
		colCell := cell.NewCell(0, int32(c), label)
		table.SetCell(0, int(c-vp.LeftCol+1), colCell.ToTViewCell().SetAlign(tview.AlignCenter))
	}

	for r := vp.TopRow; r < vp.TopRow+vp.ViewRows; r++ {
		label := fmt.Sprintf("%d", r)
		rowCell := cell.NewCell(int32(r), 0, label)
		rowCell.MinWidth = 2
		rowCell.MaxWidth = int16(len(label))+2
		table.SetCell(int(r-vp.TopRow+1), 0, rowCell.ToTViewCell())
	}

	for r := vp.TopRow; r < vp.TopRow+vp.ViewRows; r++ {
		for c := vp.LeftCol; c < vp.LeftCol+vp.ViewCols; c++ {
			key := [2]int{int(r), int(c)}
			visualRow := r - vp.TopRow + 1
			visualCol := c - vp.LeftCol + 1
			
			var tvCell *tview.TableCell
			if cellData, exists := data[key]; exists {
				tvCell = cellData.ToTViewCell()
			} else {
				tvCell = tview.NewTableCell("").
					SetAlign(tview.AlignLeft).
					SetTextColor(tcell.NewRGBColor(255, 255, 255)).
					SetBackgroundColor(tcell.NewRGBColor(0, 0, 0))	
			}
			
			table.SetCell(int(visualRow), int(visualCol), tvCell)
		}
	}
	
	CleanupDistantCells(data, vp, 100)
}

// MarkAsModified marks the file as modified
func MarkAsModified(table *tview.Table) {
	if globalWorkbook != nil {
		globalWorkbook.HasChanges = true
		updateTableTitle(table)
	}
}

// MarkAsSaved marks the workbook as saved
func MarkAsSaved(table *tview.Table) {
	if globalWorkbook != nil {
		globalWorkbook.HasChanges = false
		updateTableTitle(table)
	}
}

// SetCurrentFilename sets the current filename and updates title
func SetCurrentFilename(table *tview.Table, filename string) {
	if globalWorkbook != nil {
		globalWorkbook.CurrentFile = filename
		updateTableTitle(table)
	}
}

// Updates the table title to include a • that signals that the file has been modified
func updateTableTitle(table *tview.Table) {
	if globalWorkbook == nil {
		table.SetTitle(" Untitled ")
		return
	}

	sheet := globalWorkbook.GetActiveSheet()
	if sheet == nil {
		table.SetTitle(" Untitled ")
		return
	}

	filename := globalWorkbook.CurrentFile

	var title string
	if filename == "" {
		title = fmt.Sprintf(" Untitled - %s ", sheet.Name)
	} else {
		title = fmt.Sprintf(" %s - %s ", filepath.Base(filename), sheet.Name)
	}

	if globalWorkbook.HasChanges {
		title += "● "
	}

	table.SetTitle(title)
}

// HasUnsavedChanges returns whether there are unsaved changes
func HasUnsavedChanges() bool {
	if globalWorkbook == nil {
		return false
	}
	return globalWorkbook.HasChanges
}

// GetCurrentFilename returns the current filename
func GetCurrentFilename() string {
	if globalWorkbook == nil {
		return ""
	}
	return globalWorkbook.CurrentFile
}

// GetWorkbook returns the global workbook (for file operations)
func GetWorkbook() *Workbook {
	return globalWorkbook
}
