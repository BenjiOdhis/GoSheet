// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// open.go handles opening and reading .gsheet, .json, or .txt files with workbook support.

package fileop

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"io"
	"os"
	"strings"
)

// WorkbookResult contains all sheets loaded from a file
type WorkbookResult struct {
	Sheets      []SheetResult
	ActiveSheet int
	Version     string
}

// SheetResult contains the data for a single sheet
type SheetResult struct {
	Name  string
	Cells []*cell.Cell
	Rows  int32
	Cols  int32
}

// OpenWorkbook loads a workbook from a .gsheet, .json, or .txt file
func OpenWorkbook(filename string) (*WorkbookResult, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist")
	}

	if strings.HasSuffix(filename, ".txt") {
		cells, rows, cols, err := openTxtFile(filename)
		if err != nil {
			return nil, err
		}
		return &WorkbookResult{
			Sheets: []SheetResult{
				{
					Name:  "Workbook1",
					Cells: cells,
					Rows:  rows,
					Cols:  cols,
				},
			},
			ActiveSheet: 0,
			Version:     utils.FILEVER,
		}, nil
	}

	if !(strings.HasSuffix(filename, ".gsheet") || strings.HasSuffix(filename, ".json")) {
		return nil, fmt.Errorf("invalid file format (expected .gsheet, .json, or .txt)")
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(filename, ".gsheet") {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to open gzip: %v", err)
		}
		defer gz.Close()
		reader = gz
	}

	var wbData WorkbookData
	if err := json.NewDecoder(reader).Decode(&wbData); err != nil {
		return nil, fmt.Errorf("failed to decode workbook: %v", err)
	}

	// Check if it's a legacy single-sheet file
	if wbData.Version == "" && len(wbData.Sheets) == 0 {
		file.Close()
		return openLegacyFormat(filename)
	}

	result := &WorkbookResult{
		Sheets:      make([]SheetResult, 0, len(wbData.Sheets)),
		ActiveSheet: wbData.ActiveSheet,
		Version:     wbData.Version,
	}

	for _, sheetData := range wbData.Sheets {
		cells := processCellData(sheetData.Cells)
		result.Sheets = append(result.Sheets, SheetResult{
			Name:  sheetData.Name,
			Cells: cells,
			Rows:  sheetData.Rows,
			Cols:  sheetData.Cols,
		})
	}

	return result, nil
}

// openLegacyFormat handles old single-sheet format
func openLegacyFormat(filename string) (*WorkbookResult, error) {
	cells, err := OpenTable(filename)
	if err != nil {
		return nil, err
	}

	return &WorkbookResult{
		Sheets: []SheetResult{
			{
				Name:  "Workbook1",
				Cells: cells,
			},
		},
		ActiveSheet: 0,
		Version:     "1.0",
	}, nil
}

// processCellData converts CellData map to cell slice
func processCellData(cellDataMap map[string]*CellData) []*cell.Cell {
	var cells []*cell.Cell

	for _, c := range cellDataMap {
		c.Cell.RawValue = &c.RawValue

		if c.Cell.Display == nil {
			displayValue := c.RawValue
			c.Cell.Display = &displayValue
		}

		if c.Cell.Type == nil {
			typeValue := "string"
			c.Cell.Type = &typeValue
		}

		if c.Cell.Notes == nil {
			emptyStr := ""
			c.Cell.Notes = &emptyStr
		}

		if c.Cell.Valrule == nil {
			emptyStr := ""
			c.Cell.Valrule = &emptyStr
		}

		if c.Cell.Valrulemsg == nil {
			emptyStr := ""
			c.Cell.Valrulemsg = &emptyStr
		}

		if c.Cell.DependsOn == nil {
			c.Cell.DependsOn = []*string{}
		}

		if c.Cell.Dependents == nil {
			c.Cell.Dependents = []*string{}
		}

		cells = append(cells, c.Cell)
	}

	return cells
}

// Legacy OpenTable function for backward compatibility - returns only first sheet
func OpenTable(filename string) ([]*cell.Cell, error) {
	result, err := OpenWorkbook(filename)
	if err != nil {
		return nil, err
	}

	if len(result.Sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in workbook")
	}

	firstSheet := result.Sheets[0]
	return firstSheet.Cells, nil
}

// openTxtFile opens a tab-delimited text file and converts it to cells
func openTxtFile(filename string) ([]*cell.Cell, int32, int32, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, err
	}
	defer file.Close()

	var cells []*cell.Cell
	scanner := bufio.NewScanner(file)
	var maxCol int32
	row := int32(1)

	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Split(line, "\t")

		if int32(len(values)) > maxCol {
			maxCol = int32(len(values))
		}

		for col, value := range values {
			if value == "" {
				continue
			}

			cellValue := value
			displayValue := value
			typeValue := "string"
			emptyStr := ""
			autotype := "auto"

			c := &cell.Cell{
				Row:      row,
				Column:   int32(col + 1),
				MaxWidth: utils.DEFAULT_CELL_MAX_WIDTH,
				MinWidth: utils.DEFAULT_CELL_MIN_WIDTH,
				RawValue: &cellValue,
				Display:  &displayValue,
				Type:     &typeValue,

				Notes:   &emptyStr,
				Valrule: &emptyStr,

				Color:   utils.ColorOptions["White"],
				BgColor: utils.ColorOptions["Black"],

				DecimalPoints:      utils.DEFAULT_CELL_DECIMAL_POINTS,
				ThousandsSeparator: utils.DEFAULT_CELL_THOUSANDS_SEPARATOR,
				DecimalSeparator:   utils.DEFAULT_CELL_DECIMAL_SEPARATOR,
				FinancialSign:      utils.DEFAULT_CELL_FINANCIAL_SIGN,
				DateTimeFormat:     &autotype,

				Align: 0,
				Flags: 0,

				DependsOn:  []*string{},
				Dependents: []*string{},
			}
			cells = append(cells, c)
		}
		row++
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("error reading file: %v", err)
	}

	rows := row - 1
	return cells, rows, maxCol, nil
}
