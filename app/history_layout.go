package app

import (
	"fyne.io/fyne/v2"
)

var _ fyne.Layout = (*HistoryLayout)(nil)

type HistoryLayout struct{}

func NewHistoryLayout() *HistoryLayout {
	return &HistoryLayout{}
}

func (g *HistoryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x, y := float32(0), float32(0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		if x+child.Size().Width >= size.Width {
			x = 0
			y += child.Size().Height
		}
		child.Move(fyne.NewPos(x, y))
		x += child.Size().Width
	}
}

func (g *HistoryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(10, float32(len(objects)*10))
}
