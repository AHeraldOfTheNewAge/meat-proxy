package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	dataFile   = "notes.json"
	colorBg    = "#1e1e2e"
	colorWhite = "#ffffff"
	colorBlack = "#000000"
	colorRed   = "#ff004c"
)

var notes []string
var cur int

func load() {
	b, err := os.ReadFile(dataFile)

	if err != nil { //TODO Maybe trigger some errors?
		return
	}

	json.Unmarshal(b, &notes)
}

func save() {
	b, _ := json.Marshal(notes)
	os.WriteFile(dataFile, b, 0o644)
}

func getNoteTitle(s string) string {
	t := strings.SplitN(s, "\n", 2)[0]

	if t == "" {
		return "<empty>"
	}

	return t
}

func initialiseBorderColorsListOrEditor(element *tview.Box) {
	element.SetBorderColor(tcell.GetColor(colorWhite))
	element.SetFocusFunc(func() { element.SetBorderColor(tcell.GetColor(colorRed)) })
	element.SetBlurFunc(func() { element.SetBorderColor(tcell.GetColor(colorWhite)) })
}

func main() {
	load()

	app := tview.NewApplication()

	tview.Styles.PrimitiveBackgroundColor = tcell.GetColor(colorBg)
	tview.Styles.TitleColor = tcell.GetColor(colorWhite)
	tview.Styles.BorderColor = tcell.GetColor(colorWhite)

	header := tview.NewTextView().SetTextColor(tcell.GetColor(colorRed)).SetTextAlign(tview.AlignCenter).SetText("MEAT-PROXY")

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(" Notes ")
	initialiseBorderColorsListOrEditor(list.Box)

	editor := tview.NewTextArea()
	editor.SetBorder(true).SetTitle(" Edit ")
	initialiseBorderColorsListOrEditor(editor.Box)

	add := tview.NewButton("Add")

	add.SetStyle(tcell.StyleDefault.
		Background(tcell.GetColor(colorWhite)).
		Foreground(tcell.GetColor(colorBlack)))
	add.SetActivatedStyle(tcell.StyleDefault.
		Background(tcell.GetColor(colorBlack)).
		Foreground(tcell.GetColor(colorWhite))) // Focused!

	del := tview.NewButton("Delete")
	del.SetStyle(tcell.StyleDefault.
		Background(tcell.GetColor(colorRed)).
		Foreground(tcell.GetColor(colorBlack)))
	del.SetActivatedStyle(tcell.StyleDefault.
		Background(tcell.GetColor(colorBlack)).
		Foreground(tcell.GetColor(colorRed))) // Focused!

	help := tview.NewTextView().SetText("Tab: Switch element focus | If on list n: New note, Delete: Remove note")

	refresh := func() {
		curAux := cur // remember before Clear/AddItem overwrite it
		list.Clear()

		for _, n := range notes {
			list.AddItem(getNoteTitle(n), "", 0, nil)
		}

		cur = curAux // restore - so we can focus the newly added note!

		if len(notes) > 0 {
			list.SetCurrentItem(cur)
			editor.SetText(notes[cur], true)

			return
		}

		editor.SetText("", true)
	}

	list.SetChangedFunc(func(i int, _, _ string, _ rune) {
		if i >= len(notes) {
			return
		}

		cur = i
		editor.SetText(notes[i], true)
	})

	editor.SetChangedFunc(func() {
		if cur >= len(notes) {
			return
		}

		t := editor.GetText()
		notes[cur] = t
		list.SetItemText(cur, getNoteTitle(t), "")

		save()
	})

	doAdd := func() {
		notes = append(notes, "")
		cur = len(notes) - 1

		refresh()
		save()

		app.SetFocus(editor)
	}

	doDelete := func() {
		if len(notes) == 0 {
			return
		}

		notes = append(notes[:cur], notes[cur+1:]...)

		if cur >= len(notes) {
			cur = len(notes) - 1
		}

		if cur < 0 {
			cur = 0
		}

		refresh()
		save()

		app.SetFocus(list)
	}

	add.SetSelectedFunc(doAdd)
	del.SetSelectedFunc(doDelete)

	list.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyDelete { // On delete keyboard button press - remove a note
			doDelete()

			return nil
		}

		if e.Rune() == 'n' { // On n keyboard button press - add a new note
			doAdd()

			return nil
		}

		return e // Pass other keys through
	})

	top := tview.NewFlex().AddItem(list, 40, 0, true).AddItem(editor, 0, 1, false)

	bottom := tview.NewFlex().
		AddItem(add, 10, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(del, 10, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(help, 0, 1, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(top, 0, 1, true).
		AddItem(bottom, 1, 0, false)

	order := []tview.Primitive{list, editor, add, del} //TABCYCLE?

	app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() != tcell.KeyTab {
			return e
		}

		for i, p := range order {
			if !p.HasFocus() {
				continue
			}

			app.SetFocus(order[(i+1)%len(order)])

			break
		}

		return nil
	})

	refresh()

	if err := app.SetRoot(root, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
