package navigation

import "github.com/rivo/tview"

// Warning Modal
func ShowWarningModal(app *tview.Application, returnTo tview.Primitive, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(returnTo, true).SetFocus(returnTo)
		})
	modal.SetBorder(true).SetTitle(" Info ").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)
}
