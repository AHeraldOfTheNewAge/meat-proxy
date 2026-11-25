package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const dataFile = "notes.json"

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

func getTitle(s string) string {
	t := strings.SplitN(s, "\n", 2)[0]

	if t == "" {
		return "<empty>"
	}

	return t //TODO -> Maybe add a star in front of it?
}

func main() {
	load()

	app := tview.NewApplication()

	tview.Styles.PrimitiveBackgroundColor = tcell.GetColor("#1e1e2e")
	tview.Styles.TitleColor = tcell.GetColor("#fff")
	tview.Styles.BorderColor = tcell.GetColor("#fff")

	header := tview.NewTextView().SetTextColor(tcell.GetColor("#ff004c")).SetTextAlign(tview.AlignCenter).SetText("MEAT-PROXY")

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(" Notes ")

	editor := tview.NewTextArea()
	editor.SetBorder(true).SetTitle(" Edit ")

	add := tview.NewButton("Add")
	del := tview.NewButton("Delete")
	help := tview.NewTextView().SetText("n: new | Tab: switch") //TODO

	refresh := func() {
		list.Clear()

		for _, n := range notes {
			list.AddItem(getTitle(n), "", 0, nil)
		}

		if len(notes) > 0 {
			list.SetCurrentItem(cur)
			editor.SetText(notes[cur], true)

			return
		}

		editor.SetText("", true)
	}

	list.SetChangedFunc(func(i int, _, _ string, _ rune) {
		if i < len(notes) {
			cur = i
			editor.SetText(notes[i], true)
		}
	})

	editor.SetChangedFunc(func() {
		if cur < len(notes) {
			t := editor.GetText()
			notes[cur] = t
			list.SetItemText(cur, getTitle(t), "")

			save()
		}
	})

	add.SetSelectedFunc(func() {
		notes = append(notes, "")
		cur = len(notes) - 1

		refresh()
		save()

		app.SetFocus(editor)
	})

	del.SetSelectedFunc(func() {
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
