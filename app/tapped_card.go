package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TappedCard 是一个自定义的包含点击事件的Card组件
type TappedCard struct {
	widget.BaseWidget
	card   *widget.Card
	onTap  func()
	offTap func()
}

func NewTappedCard(title, subtitle string, content fyne.CanvasObject, onTap func(), offTap func()) *TappedCard {
	card := &TappedCard{
		card:   widget.NewCard(title, subtitle, content),
		onTap:  onTap,
		offTap: offTap,
	}
	card.ExtendBaseWidget(card)
	return card
}

// Tapped 在组件被点击时调用
func (m *TappedCard) Tapped(*fyne.PointEvent) {
	if m.onTap != nil {
		m.onTap() // 调用点击事件处理函数
	}
}

// TappedSecondary 是右击事件
func (m *TappedCard) TappedSecondary(*fyne.PointEvent) {
	if m.offTap != nil {
		m.offTap() // 调用点击事件处理函数
	}
}

func (m *TappedCard) CreateRenderer() fyne.WidgetRenderer {
	return m.card.CreateRenderer()
}

func (m *TappedCard) MinSize() fyne.Size {
	return m.card.MinSize()
}

func (m *TappedCard) ReSize(size fyne.Size) {
	m.card.Resize(size)
	m.BaseWidget.Resize(size)
}

func (m *TappedCard) Position() fyne.Position {
	return m.card.Position()
}

func (m *TappedCard) Move(pos fyne.Position) {
	m.card.Move(pos)
	m.BaseWidget.Move(pos)
}

func (m *TappedCard) Show() {
	m.card.Show()
	m.BaseWidget.Show()
}

func (m *TappedCard) Hide() {
	m.card.Hide()
	m.BaseWidget.Hide()
}

func (m *TappedCard) Refresh() {
	m.card.Refresh()
	m.BaseWidget.Refresh()
}

func (m *TappedCard) Visible() bool {
	return m.card.Visible()
}
