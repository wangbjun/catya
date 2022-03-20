package app

import (
	"fyne.io/fyne/v2"
)

var _ fyne.Layout = (*HistoryLayout)(nil)

type HistoryLayout struct {
	height float32
}

func NewHistoryLayout() *HistoryLayout {
	return &HistoryLayout{}
}

func (g *HistoryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x, y := float32(0), float32(0)
	for i, child := range objects {
		if !child.Visible() {
			continue
		}
		if i == 0 {
			g.height = 0
		}
		child.Move(fyne.NewPos(x, y))
		x += child.MinSize().Width
		if x >= size.Width {
			x = 0
			y += child.MinSize().Height
			child.Move(fyne.NewPos(x, y))
			x += child.MinSize().Width
			g.height += child.MinSize().Height
		}
		child.Resize(fyne.NewSize(child.MinSize().Width, 36))
	}
}

func (g *HistoryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, g.height)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		minSize.Width = fyne.Max(child.MinSize().Width, minSize.Width)
	}
	return minSize
}
