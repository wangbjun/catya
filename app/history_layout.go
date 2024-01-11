package app

import (
	"fyne.io/fyne/v2"
)

var _ fyne.Layout = (*HistoryLayout)(nil)

type HistoryLayout struct {
	width          float32
	height         float32
	originalWidth  float32
	originalHeight float32
}

func NewHistoryLayout() *HistoryLayout {
	return &HistoryLayout{
		width:          220,
		height:         175,
		originalWidth:  220,
		originalHeight: 175,
	}
}

func (g *HistoryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	perRow := int(size.Width / g.originalWidth)
	remainingSpace := size.Width - (g.originalWidth * float32(perRow))

	g.width = g.originalWidth
	g.height = g.originalHeight
	//如果有剩余空间，则增加对象大小来填充空间
	if remainingSpace > 0 && remainingSpace < g.originalWidth {
		perSpace := remainingSpace / float32(perRow)
		g.width += perSpace
		g.height += g.height * (perSpace / g.width)
	}

	x, y := float32(0), float32(0)
	for _, child := range objects {
		child.Resize(fyne.NewSize(g.width, g.height))
		if int(x+g.width) > int(size.Width) {
			x = 0
			y += g.height
		}
		child.Move(fyne.NewPos(x, y))
		x += g.width
	}
}

func (g *HistoryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(g.width*2, g.height)
}
