package app

import (
	"fyne.io/fyne/v2"
)

var _ fyne.Layout = (*ProportionLayout)(nil)

type ProportionLayout struct {
	ratio []float64
}

func NewProportionLayout(ratio []float64) *ProportionLayout {
	return &ProportionLayout{ratio: ratio}
}

func (m *ProportionLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	pos := fyne.NewPos(0, 0)
	totalRatio := 0.0
	for _, r := range m.ratio {
		totalRatio += r
	}
	for i, obj := range objects {
		objRatio := m.ratio[i]
		objWidth := size.Width * float32(objRatio/totalRatio)
		objHeight := size.Height
		obj.Resize(fyne.NewSize(objWidth, objHeight))
		obj.Move(pos)

		pos = pos.Add(fyne.NewPos(objWidth, 0))
	}
}

func (m *ProportionLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minWidth := float32(0)
	minHeight := float32(0)
	for _, obj := range objects {
		minSize := obj.MinSize()
		minWidth += minSize.Width
		minHeight = fyne.Max(minHeight, minSize.Height)
	}
	return fyne.NewSize(minWidth, minHeight)
}
