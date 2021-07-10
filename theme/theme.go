package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

type MyTheme struct{}

var _ fyne.Theme = (*MyTheme)(nil)

// Font return bundled font resource
// ResourceSourceHanSansTtf 即是 bundle.go 文件中 var 的变量名
func (m MyTheme) Font(s fyne.TextStyle) fyne.Resource {
	return resourceWqyMicroheiTtc
}
func (*MyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.LightTheme().Color(n, v)
}

func (*MyTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.LightTheme().Icon(n)
}

func (*MyTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.LightTheme().Size(n)
}
