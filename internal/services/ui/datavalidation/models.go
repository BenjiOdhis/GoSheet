package datavalidation

import "regexp"

var validationCellRefRegex = regexp.MustCompile(`\b([A-Z]+)(\d+)\b`)

// ValidationPreset represents a predefined validation type
type ValidationPreset struct {
	Name        string
	Description string
	BuildRule   func(params map[string]string) string
	Fields      []ValidationField
}

type ValidationField struct {
	Name        string
	Label       string
	Type        string
	Placeholder string
}
