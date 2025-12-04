// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// save.go handles saving the current state of the table to a file.

package fileop

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"compress/gzip"

	"gosheet/internal/services/cell"
	"gosheet/internal/utils"

	"github.com/rivo/tview"
)

// SaveWorkbook saves multiple sheets in native .gsheet format
func SaveWorkbook(sheets []SheetInfo, activeSheet int, filename string) error {
	if !strings.HasSuffix(filename, ".gsheet") {
		if idx := strings.Index(filename, "."); idx != -1 {
			filename = filename[:idx]
		}
		filename += ".gsheet"
	}

	wbData := WorkbookData{
		Version:     utils.FILEVER,
		ActiveSheet: activeSheet,
		Sheets:      make([]SheetData, 0, len(sheets)),
	}

	for _, sheet := range sheets {
		sheetData := SheetData{
			Name:  sheet.Name,
			Rows:  sheet.Rows,
			Cols:  sheet.Cols,
			Cells: make(map[string]*CellData),
		}

		for _, c := range sheet.GlobalData {
			cName := fmt.Sprintf("%s%d", utils.ColumnName(int32(c.Column)), c.Row)
			cleanRawValue := cell.StripTviewTags(strings.TrimSpace(*c.RawValue))

			sheetData.Cells[cName] = &CellData{
				Cell:     c,
				RawValue: cleanRawValue,
			}
		}

		wbData.Sheets = append(wbData.Sheets, sheetData)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	jsonBytes, err := json.MarshalIndent(wbData, "", "  ")
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(f)
	defer gz.Close()

	_, err = gz.Write(jsonBytes)
	return err
}

// SaveWorkbookAsJSON saves multiple sheets in .json format
func SaveWorkbookAsJSON(sheets []SheetInfo, activeSheet int, filename string) error {
	if !strings.HasSuffix(filename, ".json") {
		if idx := strings.Index(filename, "."); idx != -1 {
			filename = filename[:idx]
		}
		filename += ".json"
	}

	wbData := WorkbookData{
		Version:     utils.FILEVER,
		ActiveSheet: activeSheet,
		Sheets:      make([]SheetData, 0, len(sheets)),
	}

	for _, sheet := range sheets {
		sheetData := SheetData{
			Name:  sheet.Name,
			Rows:  sheet.Rows,
			Cols:  sheet.Cols,
			Cells: make(map[string]*CellData),
		}

		for _, c := range sheet.GlobalData {
			cName := fmt.Sprintf("%s%d", utils.ColumnName(int32(c.Column)), c.Row)
			cleanRawValue := cell.StripTviewTags(strings.TrimSpace(*c.RawValue))

			sheetData.Cells[cName] = &CellData{
				Cell:     c,
				RawValue: cleanRawValue,
			}
		}

		wbData.Sheets = append(wbData.Sheets, sheetData)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(wbData)
}

// SaveTable saves in native .gsheet format (legacy single-sheet wrapper)
func SaveTable(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	sheets := []SheetInfo{
		{
			Name:       "Sheet1",
			Rows:       int32(table.GetRowCount()),
			Cols:       int32(table.GetColumnCount()),
			GlobalData: globalData,
		},
	}
	return SaveWorkbook(sheets, 0, filename)
}

// SaveTableAsJSON saves the table in .json format (legacy single-sheet wrapper)
func SaveTableAsJSON(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	sheets := []SheetInfo{
		{
			Name:       "Sheet1",
			Rows:       int32(table.GetRowCount()),
			Cols:       int32(table.GetColumnCount()),
			GlobalData: globalData,
		},
	}
	return SaveWorkbookAsJSON(sheets, 0, filename)
}

// SaveTableAsCSV exports table as CSV
func SaveTableAsCSV(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var maxRow, maxCol int32
	for key := range globalData {
		r, c := int32(key[0]), int32(key[1])
		if r > maxRow {
			maxRow = r
		}
		if c > maxCol {
			maxCol = c
		}
	}

	for row := int32(1); row <= maxRow; row++ {
		record := make([]string, maxCol)

		for col := int32(1); col <= maxCol; col++ {
			key := [2]int{int(row), int(col)}
			if cellData, exists := globalData[key]; exists && cellData.RawValue != nil {
				record[col-1] = *cellData.RawValue
			} else {
				record[col-1] = ""
			}
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to get alignment string
func getAlignmentStyle(align int8) string {
	switch align {
	case tview.AlignLeft:
		return "left"
	case tview.AlignCenter:
		return "center"
	case tview.AlignRight:
		return "right"
	default:
		return "left"
	}
}

// Helper function to build cell style
func buildCellStyle(cellData *cell.Cell) string {
	var styles []string

	if cellData.Color != utils.ColorOptions["White"] {
		styles = append(styles, "color: "+cellData.Color.Hex())
	}

	if cellData.BgColor != utils.ColorOptions["Black"] {
		styles = append(styles, "background-color: "+cellData.BgColor.Hex())
	}

	styles = append(styles, fmt.Sprintf("text-align: %s", getAlignmentStyle(cellData.Align)))

	if cellData.HasFlag(cell.FlagBold) {
		styles = append(styles, "font-weight: bold")
	}

	if cellData.HasFlag(cell.FlagUnderline) {
		styles = append(styles, "text-decoration: underline")
	}

	if cellData.HasFlag(cell.FlagStrikethrough) {
		if cellData.HasFlag(cell.FlagUnderline) {
			styles = append(styles, "text-decoration: underline line-through")
		} else {
			styles = append(styles, "text-decoration: line-through")
		}
	}

	styles = append(styles, "padding: 8px")

	return strings.Join(styles, "; ")
}

// Helper function for escape codes in html
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// SaveTableAsHTML exports table as a HTML webpage
func SaveTableAsHTML(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	var maxRow, maxCol int32
	for key := range globalData {
		r, c := int32(key[0]), int32(key[1])
		if r > maxRow {
			maxRow = r
		}
		if c > maxCol {
			maxCol = c
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>GoSheet Export</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			table { border-collapse: collapse; width: 100%; }
			th, td { border: 1px solid #ddd; padding: 8px; }
			th { background-color: #4CAF50; color: white; font-weight: bold; }
			tr:hover { background-color: #f5f5f5; }
			.formula-cell { border-left: 3px solid #2196F3; }
		</style>
	</head>
	<body>
		<table>
			<thead>
				<tr>
					<th>#</th>
`)

	for col := int32(1); col <= maxCol; col++ {
		html.WriteString(fmt.Sprintf("<th>%s</th>\n", utils.ColumnName(col)))
	}

	html.WriteString("</tr>\n</thead>\n<tbody>\n")

	for row := int32(1); row <= maxRow; row++ {
		html.WriteString("<tr>\n")
		html.WriteString(fmt.Sprintf("<td style=\"background-color: #4CAF50; color: white; font-weight: bold;\"><b>%d</b></td>\n", row))

		for col := int32(1); col <= maxCol; col++ {
			key := [2]int{int(row), int(col)}
			cellData, exists := globalData[key]

			if !exists || cellData == nil || cellData.Display == nil {
				html.WriteString("<td></td>\n")
				continue
			}

			content := *cellData.Display

			if cellData.HasFlag(cell.FlagAllCaps) {
				content = strings.ToUpper(content)
			}

			style := buildCellStyle(cellData)

			class := ""
			tooltip := ""
			if cellData.HasFlag(cell.FlagFormula) {
				class = " class=\"formula-cell\""
				tooltip = fmt.Sprintf(" title=\"Formula: %s\"", htmlEscape(*cellData.RawValue))
			}

			if style != "" {
				html.WriteString(
					fmt.Sprintf("<td%s style=\"%s\"%s>%s</td>\n",
						class, style, tooltip, htmlEscape(content)),
				)
			} else {
				html.WriteString(
					fmt.Sprintf("<td%s%s>%s</td>\n",
						class, tooltip, htmlEscape(content)),
				)
			}
		}

		html.WriteString("</tr>\n")
	}

	html.WriteString("</tbody>\n</table>\n</body>\n</html>")

	_, err = file.WriteString(html.String())
	return err
}

// SaveTableAsTXT exports table as tab-delimited text file
func SaveTableAsTXT(table *tview.Table, filename string, globalData map[[2]int]*cell.Cell) error {
	if !strings.HasSuffix(filename, ".txt") {
		if idx := strings.Index(filename, "."); idx != -1 {
			filename = filename[:idx]
		}
		filename += ".txt"
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var maxRow, maxCol int32
	for key := range globalData {
		r, c := int32(key[0]), int32(key[1])
		if r > maxRow {
			maxRow = r
		}
		if c > maxCol {
			maxCol = c
		}
	}

	for row := int32(1); row <= maxRow; row++ {
		var values []string

		for col := int32(1); col <= maxCol; col++ {
			key := [2]int{int(row), int(col)}
			if cellData, exists := globalData[key]; exists && cellData.RawValue != nil {
				values = append(values, *cellData.RawValue)
			} else {
				values = append(values, "")
			}
		}

		line := strings.Join(values, "\t") + "\n"
		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

// SaveTableAsExcel will save the table in the excel format. Currently not implemented.
func SaveTableAsExcel(table *tview.Table, filename string) error {
	// github.com/xuri/excelize/v2
	return fmt.Errorf("Excel export not yet implemented. Use CSV or HTML format instead.")
}

// SaveTableAsPDF will save the table as a PDF
func SaveTableAsPDF(table *tview.Table, filename string) error {
	// gopdf
	return fmt.Errorf("PDF export not yet implemented. Use CSV or HTML format instead.")
}
