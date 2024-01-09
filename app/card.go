package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TappedCard 是一个自定义的包含点击事件的Card组件
type TappedCard struct {
	*widget.Card
	onTap  func()
	offTap func()
}

func NewTappedCard(title, subtitle string, content fyne.CanvasObject, onTap func(), offTap func()) *TappedCard {
	card := &TappedCard{
		Card:   widget.NewCard(title, subtitle, content),
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
