// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// model.go provides data models for saving and loading spreadsheet data with workbook support.

package fileop

import (
	"gosheet/internal/services/cell"
)

// CellData represents the data of a single cell in the saved spreadsheet
type CellData struct {
	Cell     *cell.Cell
	RawValue string
}

// SheetData represents the structure of a single sheet
type SheetData struct {
	Name  string                  `json:"name"`
	Rows  int32                   `json:"rows"`
	Cols  int32                   `json:"cols"`
	Cells map[string]*CellData    `json:"cells"`
}

// SheetInfo is amost the same as SheetData
type SheetInfo struct {
	Name       string
	Rows       int32
	Cols       int32
	GlobalData map[[2]int]*cell.Cell
}

// WorkbookData represents the complete workbook structure for saving/loading
type WorkbookData struct {
	Version     string       `json:"version"`
	ActiveSheet int          `json:"active_sheet"`
	Sheets      []SheetData  `json:"sheets"`
}
