package fynekit

import (
	"fyne.io/fyne/v2"
)

func WindowForObject(obj fyne.CanvasObject) fyne.Window {
	a := fyne.CurrentApp()
	if a == nil {
		return nil
	}

	d := a.Driver()
	if d == nil {
		return nil
	}

	c := d.CanvasForObject(obj)
	if c == nil {
		return nil
	}

	for _, w := range d.AllWindows() {
		if w.Canvas() == c {
			return w
		}
	}

	return nil
}
