package main

import (
	"catya/app"
	"fyne.io/fyne/v2"
)

func main() {
	huya := app.New()
	huya.SetUp()
	huya.Window.Resize(fyne.NewSize(640, 480))
	huya.Window.CenterOnScreen()
	huya.Window.ShowAndRun()
}
