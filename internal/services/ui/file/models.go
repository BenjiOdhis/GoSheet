package file

import (
	"gosheet/internal/services/cell"
	"gosheet/internal/services/fileop"

	"github.com/rivo/tview"
)

// FileFormat represents supported export formats
type FileFormat struct {
	Extension   string
	Description string
	SaveFunc    func(*tview.Table, string, map[[2]int]*cell.Cell) error
}

// Available file formats
var FileFormats = []FileFormat{
	{Extension: ".gsheet", Description: "GSheet (Native)", SaveFunc: fileop.SaveTable},      // JSON zipped using gzip, so very efficient
	{Extension: ".json", Description: "JSON", SaveFunc: fileop.SaveTableAsJSON},         	     // works perfectly fine, but takes up more disk space 
	{Extension: ".csv", Description: "CSV", SaveFunc: fileop.SaveTableAsCSV},                  // works only for text, as CSV can only save text
	{Extension: ".html", Description: "HTML Table", SaveFunc: fileop.SaveTableAsHTML},		 // works; only for exporting
	//{Extension: ".xlsx", Description: "Excel Spreadsheet", SaveFunc: file.SaveTableAsExcel}, // not implemented
	{Extension: ".txt", Description: "Tab delimited text file", SaveFunc: fileop.SaveTableAsTXT}, // working, only for text
	//{Extension: ".dbf", Description: "dBASE file", SaveFunc: file.SaveTableAsDBF}, // not implemented
	//{Extension: ".ods", Description: "OpenDocument Spreadsheet", SaveFunc: file.SaveTableAsODS} // not implemented
	//{Extension: ".pdf", Description: "Portable Document Format", SaveFUnc: file.SaveTableAsPDF} // not implemented
}


