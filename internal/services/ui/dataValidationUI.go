// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// dataValidationUI.go implements cell validation rules with Excel-like presets

package ui

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

// GetValidationPresets returns all available validation presets
func GetValidationPresets() []ValidationPreset {
	return []ValidationPreset{
		{
			Name:        "Custom",
			Description: "Write your own validation expression using 'THIS' to refer to the cell value",
			Fields:      []ValidationField{},
			BuildRule:   func(params map[string]string) string { return params["custom"] },
		},
		{
			Name:        "Whole Number - Between",
			Description: "Value must be a whole number between two values",
			Fields: []ValidationField{
				{Name: "min", Label: "Minimum:", Type: "number", Placeholder: "0"},
				{Name: "max", Label: "Maximum:", Type: "number", Placeholder: "100"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS >= %s && THIS <= %s && THIS == FLOOR(THIS)", 
					params["min"], params["max"])
			},
		},
		{
			Name:        "Whole Number - Greater Than",
			Description: "Value must be a whole number greater than a value",
			Fields: []ValidationField{
				{Name: "value", Label: "Greater than:", Type: "number", Placeholder: "0"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS > %s && THIS == FLOOR(THIS)", params["value"])
			},
		},
		{
			Name:        "Whole Number - Less Than",
			Description: "Value must be a whole number less than a value",
			Fields: []ValidationField{
				{Name: "value", Label: "Less than:", Type: "number", Placeholder: "100"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS < %s && THIS == FLOOR(THIS)", params["value"])
			},
		},
		{
			Name:        "Decimal - Between",
			Description: "Value must be a decimal number between two values",
			Fields: []ValidationField{
				{Name: "min", Label: "Minimum:", Type: "number", Placeholder: "0.0"},
				{Name: "max", Label: "Maximum:", Type: "number", Placeholder: "1.0"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS >= %s && THIS <= %s", params["min"], params["max"])
			},
		},
		{
			Name:        "Decimal - Greater Than",
			Description: "Value must be greater than a decimal value",
			Fields: []ValidationField{
				{Name: "value", Label: "Greater than:", Type: "number", Placeholder: "0.0"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS > %s", params["value"])
			},
		},
		{
			Name:        "Decimal - Less Than",
			Description: "Value must be less than a decimal value",
			Fields: []ValidationField{
				{Name: "value", Label: "Less than:", Type: "number", Placeholder: "100.0"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("THIS < %s", params["value"])
			},
		},
		{
			Name:        "Text Length - Between",
			Description: "Text length must be between two values",
			Fields: []ValidationField{
				{Name: "min", Label: "Minimum length:", Type: "number", Placeholder: "1"},
				{Name: "max", Label: "Maximum length:", Type: "number", Placeholder: "50"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("LEN(THIS) >= %s && LEN(THIS) <= %s", 
					params["min"], params["max"])
			},
		},
		{
			Name:        "Text Length - Maximum",
			Description: "Text cannot exceed a certain length",
			Fields: []ValidationField{
				{Name: "max", Label: "Maximum length:", Type: "number", Placeholder: "255"},
			},
			BuildRule: func(params map[string]string) string {
				return fmt.Sprintf("LEN(THIS) <= %s", params["max"])
			},
		},
		{
			Name:        "Text - Not Empty",
			Description: "Cell must contain text (cannot be empty)",
			Fields:      []ValidationField{},
			BuildRule: func(params map[string]string) string {
				return "LEN(THIS) > 0"
			},
		},
		{
			Name:        "List - Allowed Values",
			Description: "Value must be one of the specified options",
			Fields: []ValidationField{
				{Name: "list", Label: "Allowed values (comma-separated):", Type: "text", 
					Placeholder: "Yes,No,Maybe"},
			},
			BuildRule: func(params map[string]string) string {
				values := strings.Split(params["list"], ",")
				conditions := make([]string, len(values))
				for i, val := range values {
					val = strings.TrimSpace(val)
					conditions[i] = fmt.Sprintf("THIS == \"%s\"", val)
				}
				return strings.Join(conditions, " || ")
			},
		},
		{
			Name:        "Email Format",
			Description: "Value must be a valid email format",
			Fields:      []ValidationField{},
			BuildRule: func(params map[string]string) string {
				return "CONTAINS(THIS, \"@\") && CONTAINS(SUBSTR(THIS, INDEX(THIS, \"@\")), \".\")"
			},
		},
		{
			Name:        "Positive Numbers Only",
			Description: "Value must be positive (greater than 0)",
			Fields:      []ValidationField{},
			BuildRule: func(params map[string]string) string {
				return "THIS > 0"
			},
		},
		{
			Name:        "Percentage (0-100)",
			Description: "Value must be between 0 and 100",
			Fields:      []ValidationField{},
			BuildRule: func(params map[string]string) string {
				return "THIS >= 0 && THIS <= 100"
			},
		},
	}
}

// ValidateValidationRule checks if a validation rule is syntactically correct
func ValidateValidationRule(ruleText string, cellData *cell.Cell) bool {
	if strings.TrimSpace(ruleText) == "" {
		return true
	}

	upperRule := strings.ToUpper(ruleText)
	matches := validationCellRefRegex.FindAllString(upperRule, -1)

	for _, match := range matches {
		if match != "THIS" {
			return false
		}
	}

	testRule := strings.ReplaceAll(upperRule, "THIS", "5")

	_, err := govaluate.NewEvaluableExpressionWithFunctions(testRule, utils.GovalFuncs())
	if err != nil {
		return false
	}

	return true
}

// EnforceValidationOnEdit checks validation before saving a cell edit
func EnforceValidationOnEdit(app *tview.Application, returnTo tview.Primitive, cellData *cell.Cell, newValue string) bool {
	if strings.TrimSpace(newValue) == "" {
		return true
	}

	isValid, errMsg := CheckValidationRule(cellData, newValue)

	if !isValid {
		displayMsg := errMsg
		if cellData.Valrulemsg != nil && strings.TrimSpace(*cellData.Valrulemsg) != "" {
			displayMsg = *cellData.Valrulemsg
		}

		modal := tview.NewModal().
			SetText(fmt.Sprintf("Validation Failed!\n\n%s\n\nValidation Rule:\n%s", displayMsg, *cellData.Valrule)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				app.SetRoot(returnTo, true).SetFocus(returnTo)
			})

		modal.SetBackgroundColor(tcell.ColorDarkRed).
			SetBorderColor(tcell.ColorRed)
		modal.SetButtonBackgroundColor(tcell.ColorDarkRed).
			SetButtonTextColor(tcell.ColorWhite)

		app.SetRoot(modal, true).SetFocus(modal)
		return false
	}

	return true
}

func showValidationErrorModal(app *tview.Application, container *tview.Flex, returnTo tview.Primitive, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(container, true).SetFocus(returnTo)
		})

	modal.SetBackgroundColor(tcell.ColorDarkRed).
		SetBorderColor(tcell.ColorRed)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed).
		SetButtonTextColor(tcell.ColorWhite)

	app.SetRoot(modal, true).SetFocus(modal)
}

func CheckValidationRule(cellData *cell.Cell, newValue string) (bool, string) {
	if cellData.Valrule == nil || strings.TrimSpace(*cellData.Valrule) == "" {
		return true, ""
	}

	if strings.TrimSpace(newValue) == "" {
		return true, ""
	}

	var testValue any

	cellDataTypeAux := *cellData.Type
	switch strings.ToLower(cellDataTypeAux) {
	case "number", "financial":
		normalized := strings.ReplaceAll(newValue, string(cellData.ThousandsSeparator), "")
		normalized = strings.TrimPrefix(normalized, string(cellData.FinancialSign))

		if num, err := strconv.ParseFloat(normalized, 64); err == nil {
			testValue = num
		} else {
			return false, "Value must be a number"
		}

	case "string":
		testValue = newValue

	default:
		testValue = newValue
	}

	rule := strings.TrimSpace(*cellData.Valrule)
	upperRule := strings.ToUpper(rule)

	var replacementValue string
	if _, ok := testValue.(float64); ok {
		replacementValue = fmt.Sprintf("%v", testValue)
	} else {
		strValue := fmt.Sprintf("%v", testValue)
		strValue = strings.ReplaceAll(strValue, `"`, `\"`)
		replacementValue = fmt.Sprintf(`"%s"`, strValue)
	}

	evaluableRule := strings.ReplaceAll(upperRule, "THIS", replacementValue)

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(evaluableRule, utils.GovalFuncs())
	if err != nil {
		return false, fmt.Sprintf("Invalid validation rule: %s", err.Error())
	}

	result, err := expr.Evaluate(nil)
	if err != nil {
		return false, fmt.Sprintf("Validation error: %s", err.Error())
	}

	isValid, ok := result.(bool)
	if !ok {
		return false, "Validation rule must return true/false"
	}

	if !isValid {
		if cellData.Valrulemsg != nil && strings.TrimSpace(*cellData.Valrulemsg) != "" {
			return false, *cellData.Valrulemsg
		}
		return false, fmt.Sprintf("Value does not meet validation rule: %s", *cellData.Valrule)
	}

	return true, ""
}

// detectPresetFromRule tries to detect which preset was used to create a rule
func detectPresetFromRule(rule string) (int, map[string]string) {
	if strings.TrimSpace(rule) == "" {
		return 0, nil
	}

	presets := GetValidationPresets()
	
	for i, preset := range presets {
		if preset.Name == "Custom" {
			continue
		}

		switch preset.Name {
		case "Whole Number - Between":
			re := regexp.MustCompile(`THIS >= ([\d.]+) && THIS <= ([\d.]+) && THIS == FLOOR\(THIS\)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"min": matches[1], "max": matches[2]}
			}
		case "Whole Number - Greater Than":
			re := regexp.MustCompile(`THIS > ([\d.]+) && THIS == FLOOR\(THIS\)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"value": matches[1]}
			}
		case "Whole Number - Less Than":
			re := regexp.MustCompile(`THIS < ([\d.]+) && THIS == FLOOR\(THIS\)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"value": matches[1]}
			}
		case "Decimal - Between":
			re := regexp.MustCompile(`THIS >= ([\d.]+) && THIS <= ([\d.]+)$`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"min": matches[1], "max": matches[2]}
			}
		case "Decimal - Greater Than":
			re := regexp.MustCompile(`^THIS > ([\d.]+)$`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"value": matches[1]}
			}
		case "Decimal - Less Than":
			re := regexp.MustCompile(`^THIS < ([\d.]+)$`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"value": matches[1]}
			}
		case "Text Length - Between":
			re := regexp.MustCompile(`LEN\(THIS\) >= ([\d]+) && LEN\(THIS\) <= ([\d]+)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"min": matches[1], "max": matches[2]}
			}
		case "Text Length - Maximum":
			re := regexp.MustCompile(`LEN\(THIS\) <= ([\d]+)`)
			if matches := re.FindStringSubmatch(rule); matches != nil {
				return i, map[string]string{"max": matches[1]}
			}
		case "Text - Not Empty":
			if rule == "LEN(THIS) > 0" {
				return i, nil
			}
		case "List - Allowed Values":
			if strings.Contains(rule, "THIS == \"") && strings.Contains(rule, "||") {
				parts := strings.Split(rule, " || ")
				values := make([]string, 0, len(parts))
				for _, part := range parts {
					re := regexp.MustCompile(`THIS == "([^"]+)"`)
					if matches := re.FindStringSubmatch(part); matches != nil {
						values = append(values, matches[1])
					}
				}
				if len(values) > 0 {
					return i, map[string]string{"list": strings.Join(values, ",")}
				}
			}
		case "Positive Numbers Only":
			if rule == "THIS > 0" {
				return i, nil
			}
		case "Percentage (0-100)":
			if rule == "THIS >= 0 && THIS <= 100" {
				return i, nil
			}
		}
	}

	return 0, nil 
}

// ShowValidationRuleDialog displays the enhanced validation rule editor with presets
func ShowValidationRuleDialog(app *tview.Application, table *tview.Table, returnTo tview.Primitive, focus tview.Primitive, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	visualRow, visualCol := utils.ConvertToInt32(table.GetSelection())
	row, col := globalViewport.ToAbsolute(visualRow, visualCol)

	if row == 0 || col == 0 {
		return
	}

	key := [2]int{int(row), int(col)}
	cellData, exists := globalData[key]
	if !exists {
		cellData = cell.NewCell(row, col, "")
		globalData[key] = cellData
	}

	if cellData.Valrule == nil {
		emptyStr := ""
		cellData.Valrule = &emptyStr
	}
	if cellData.Valrulemsg == nil {
		emptyStr := ""
		cellData.Valrulemsg = &emptyStr
	}

	presets := GetValidationPresets()
	presetNames := make([]string, len(presets))
	for i, preset := range presets {
		presetNames[i] = preset.Name
	}

	detectedPresetIdx, detectedParams := detectPresetFromRule(*cellData.Valrule)

	container := tview.NewFlex().SetDirection(tview.FlexRow)

	presetDropdown := tview.NewDropDown().
		SetLabel("Validation Type: ").
		SetOptions(presetNames, nil).
		SetCurrentOption(detectedPresetIdx)
	presetDropdown.SetBorder(true).
		SetTitle(" 1. Select Type ").
		SetBorderColor(tcell.ColorLightBlue)

	dynamicForm := tview.NewForm()
	dynamicForm.SetBorder(true).
		SetTitle(" 2. Configure ").
		SetBorderColor(tcell.ColorLightBlue)

	customRuleArea := tview.NewTextArea().
		SetPlaceholder("Enter custom validation rule using 'THIS'...\nExample: THIS > 0 && THIS < 100")
	customRuleArea.SetText(*cellData.Valrule, true)
	customRuleArea.SetBorder(true).
		SetTitle(" Custom Rule (Advanced) ").
		SetBorderColor(tcell.ColorYellow)

	customMessageInput := tview.NewInputField().
		SetLabel("Custom Error Message (optional): ").
		SetText(*cellData.Valrulemsg).
		SetFieldWidth(60).
		SetPlaceholder("Leave empty for default message")
	customMessageInput.SetBorder(true).
		SetTitle(" 3. Error Message ").
		SetBorderColor(tcell.ColorPurple)

	previewText := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	previewText.SetBorder(true).
		SetTitle(" Preview ").
		SetBorderColor(tcell.ColorGreen)

	buttonForm := tview.NewForm()
	buttonForm.AddButton("Apply", func() {
		currentPresetIdx, _ := presetDropdown.GetCurrentOption()
		preset := presets[currentPresetIdx]

		var finalRule string
		if preset.Name == "Custom" {
			finalRule = strings.TrimSpace(customRuleArea.GetText())
		} else {
			params := make(map[string]string)
			for i := 0; i < dynamicForm.GetFormItemCount(); i++ {
				item := dynamicForm.GetFormItem(i)
				if inputField, ok := item.(*tview.InputField); ok {
					if i < len(preset.Fields) {
						fieldName := preset.Fields[i].Name
						params[fieldName] = inputField.GetText()
					}
				}
			}
			
			allFilled := true
			for _, field := range preset.Fields {
				if params[field.Name] == "" {
					allFilled = false
					break
				}
			}
			
			if !allFilled {
				showValidationErrorModal(app, container, container, "Please fill in all required fields before applying.")
				return
			}
			
			finalRule = preset.BuildRule(params)
		}

		customMsg := strings.TrimSpace(customMessageInput.GetText())
		cellData.Valrulemsg = &customMsg

		saveRule(app, table, cellData, finalRule, row, col, buttonForm, returnTo, focus, container, globalData, globalViewport)
	})

	buttonForm.AddButton("Delete", func() {
		deleteRule(cellData, table, row, col, globalData, globalViewport)
		app.SetRoot(returnTo, true).SetFocus(focus)
	})

	buttonForm.AddButton("Cancel", func() {
		app.SetRoot(returnTo, true).SetFocus(focus)
	})

	var currentFormItemsInDynamic []tview.FormItem

	updateDynamicForm := func(presetIdx int) {
		dynamicForm.Clear(true)
		currentFormItemsInDynamic = nil
		preset := presets[presetIdx]

		previewText.SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s\n\n[gray]Empty cells are always allowed, validation only applies when entering a value.[-]", 
			preset.Name, preset.Description))

		if preset.Name == "Custom" {
			//if container.GetItemCount() == 5 {
			//	container.RemoveItem(dynamicForm)
			//}
			//if container.GetItemCount() == 4 {
			//	items := make([]tview.Primitive, 0)
			//	for i := 0; i < container.GetItemCount(); i++ {
			//		item := container.GetItem(i)
			//		if item == customMessageInput {
			//			items = append(items, customRuleArea)
			//		}
			//		items = append(items, item)
			//	}
			container.Clear()
				container.AddItem(presetDropdown, 0, 1, false).
				AddItem(previewText, 0, 2, false).
				AddItem(customRuleArea, 0, 5, true).
				AddItem(customMessageInput, 0, 1, false).
				AddItem(buttonForm, 0, 1, false)
			//}
		} else {
			if container.GetItemCount() == 5 {
				container.Clear()
				container.AddItem(presetDropdown, 0, 1, false).
					AddItem(previewText, 0, 2, false).
					AddItem(dynamicForm, 0, 3, false).
					AddItem(customMessageInput, 3, 1, false).
					AddItem(buttonForm, 0, 1, false)
			}

			for i, field := range preset.Fields {
				inputField := tview.NewInputField().
					SetLabel(field.Label).
					SetPlaceholder(field.Placeholder).
					SetFieldWidth(30)
				
				if detectedParams != nil && detectedPresetIdx == presetIdx {
					if val, ok := detectedParams[field.Name]; ok {
						inputField.SetText(val)
					}
				}
				
				dynamicForm.AddFormItem(inputField)
				currentFormItemsInDynamic = append(currentFormItemsInDynamic, inputField)
				
				idx := i
				inputField.SetChangedFunc(func(text string) {
					params := make(map[string]string)
					for j := 0; j < len(preset.Fields); j++ {
						fieldItem := dynamicForm.GetFormItem(j)
						if fi, ok := fieldItem.(*tview.InputField); ok {
							params[preset.Fields[j].Name] = fi.GetText()
						}
					}
					
					allFilled := true
					for _, f := range preset.Fields {
						if params[f.Name] == "" {
							allFilled = false
							break
						}
					}
					
					if allFilled {
						rule := preset.BuildRule(params)
						previewText.SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s\n\n[green]Generated Rule:[white]\n%s\n\n[gray]Empty cells are always allowed.[-]", 
							preset.Name, preset.Description, rule))
					} else {
						previewText.SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s\n\n[gray]Fill in all fields to see the generated rule.\nEmpty cells are always allowed.[-]", 
							preset.Name, preset.Description))
					}
					_ = idx 
				})
			}
			
			if detectedParams != nil && detectedPresetIdx == presetIdx && len(preset.Fields) > 0 {
				rule := preset.BuildRule(detectedParams)
				previewText.SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s\n\n[green]Current Rule:[white]\n%s\n\n[gray]Empty cells are always allowed.[-]", 
					preset.Name, preset.Description, rule))
			}
		}
	}

	presetDropdown.SetSelectedFunc(func(text string, index int) {
		updateDynamicForm(index)
	})

	container.
		AddItem(presetDropdown, 0, 1, false).
		AddItem(previewText, 0, 2, false).
		AddItem(dynamicForm, 0, 3, false).
		AddItem(customMessageInput, 0, 1, false).
		AddItem(buttonForm, 0, 1, false)

	updateDynamicForm(detectedPresetIdx)


	container.SetBorder(true).
		SetTitle(fmt.Sprintf(" Data Validation - %s%d  •  Ctrl+←/→ to navigate  •  Esc to cancel ", utils.ColumnName(col), row)).
		SetBorderColor(tcell.ColorYellow)

	getFocusablePrimitives := func() []tview.Primitive {
		focusable := []tview.Primitive{}
		
		focusable = append(focusable, presetDropdown)
		
		for i := 0; i < container.GetItemCount(); i++ {
			item := container.GetItem(i)
			if item == customRuleArea {
				focusable = append(focusable, customRuleArea)
				break
			} else if item == dynamicForm && dynamicForm.GetFormItemCount() > 0 {
				focusable = append(focusable, dynamicForm)
				break
			}
		}
		
		focusable = append(focusable, customMessageInput)
		focusable = append(focusable, buttonForm)

		return focusable
	}

	currentPrim := 0

	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetRoot(returnTo, true).SetFocus(focus)
			return nil
		}
		if event.Modifiers()&tcell.ModCtrl != 0 {
			focusables := getFocusablePrimitives()
			
			if event.Key() == tcell.KeyRight {
				currentPrim++
				currentPrim %= len(focusables)
				app.SetFocus(focusables[currentPrim])
				return nil
			} else if event.Key() == tcell.KeyLeft {
				currentPrim--
				if currentPrim < 0 {
					currentPrim = len(focusables) - 1
				}
				app.SetFocus(focusables[currentPrim])
				return nil
			}
		}
		return event
	})

	app.SetRoot(container, true).SetFocus(presetDropdown)
}

func deleteRule(cellData *cell.Cell, table *tview.Table, row, col int32, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	emptyStr := ""
	cellData.Valrule = &emptyStr
	cellData.Valrulemsg = &emptyStr

	key := [2]int{int(row), int(col)}
	globalData[key] = cellData

	if globalViewport.IsVisible(row, col) {
		visualR, visualC := globalViewport.ToRelative(row, col)
		table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
	}
}

func saveRule(app *tview.Application, table *tview.Table, cellData *cell.Cell, ruleText string, row, col int32, form tview.Primitive, returnTo tview.Primitive, focus tview.Primitive, container *tview.Flex, globalData map[[2]int]*cell.Cell, globalViewport *utils.Viewport) {
	ruleText = strings.TrimSpace(ruleText)

	if ruleText == "" || ValidateValidationRule(ruleText, cellData) {
		cellData.Valrule = &ruleText

		key := [2]int{int(row), int(col)}
		globalData[key] = cellData

		if globalViewport.IsVisible(row, col) {
			visualR, visualC := globalViewport.ToRelative(row, col)
			table.SetCell(int(visualR), int(visualC), cellData.ToTViewCell())
		}

		app.SetRoot(returnTo, true).SetFocus(focus)
	} else {
		showValidationErrorModal(app, container, form, "Invalid validation rule!\n\nMake sure:\n- You use 'THIS' instead of cell references (e.g., A1)\n- The syntax is correct\n- The rule returns true/false")
	}
}
