package file

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/services/fileop"

	"github.com/rivo/tview"
)

// FileFormatUI represents UI information for a file format
type FileFormatUI struct {
	Format      fileop.FileFormat
	Extension   string
	Description string
	SaveFunc    func(*tview.Table, string, map[[2]int]*cell.Cell) error
}

// GetFileFormats returns all available file formats for UI
func GetFileFormats() []FileFormatUI {
	formats := fileop.GetWritableFormats()
	result := make([]FileFormatUI, len(formats))

	for i, format := range formats {
		result[i] = FileFormatUI{
			Format:      format,
			Extension:   format.String(),
			Description: format.Description(),
			SaveFunc:    getSaveFunc(format),
		}
	}

	return result
}

// getSaveFunc returns the save function for a format (calls modern SaveWorkbookAs directly via hook)
func getSaveFunc(format fileop.FileFormat) func(*tview.Table, string, map[[2]int]*cell.Cell) error {
	return func(_ *tview.Table, filename string, _ map[[2]int]*cell.Cell) error {
		sheets, activeSheet, hasWorkbook := fileop.GetWorkbookForSave()
		if !hasWorkbook {
			return fileop.SaveWorkbookAs([]fileop.SheetInfo{}, 0, filename, format)
		}
		return fileop.SaveWorkbookAs(sheets, activeSheet, filename, format)
	}
}

// FileFormats array for UI code
var FileFormats = GetFileFormats()
