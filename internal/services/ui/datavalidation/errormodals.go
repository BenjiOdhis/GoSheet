package datavalidation

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
