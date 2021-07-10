package main

import (
	"catya/huya"
	"catya/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"net/url"
	"os/exec"
	"strings"
)

var hy *HuYa

type HuYa struct {
	hy     huya.Huya
	data   []string
	window fyne.Window
	entry  *widget.Entry
	button *widget.Button
	list   *widget.List
}

func init() {
	a := app.New()
	a.Settings().SetTheme(&theme.MyTheme{})

	entry := widget.NewEntry()
	entry.PlaceHolder = "请输入直播间地址或房间号，比如：https://www.huya.com/lpl、lpl"
	hy = &HuYa{
		hy:     huya.New(),
		data:   []string{},
		entry:  entry,
		window: a.NewWindow("Catya Live"),
	}
	hy.button = widget.NewButton("提交", hy.Run)
	list := widget.NewList(
		func() int {
			return len(hy.data)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(hy.data[i])
		})
	list.OnSelected = func(i widget.ListItemID) {
		hy.window.Clipboard().SetContent(hy.data[i])
		exec.Command("smplayer", hy.data[i]).Run()
	}
	hy.list = list
}

func (r *HuYa) Run() {
	roomId := strings.TrimSpace(r.entry.Text)
	if roomId == "" {
		r.Alert("请输入直播房间号")
		return
	}
	parse, err := url.Parse(roomId)
	if err == nil && parse.Path != "" {
		roomId = strings.Trim(parse.Path, "/")
	}
	r.button.Text = "查询中......"
	r.button.Disable()
	defer func() {
		r.button.Text = "提交"
		r.button.Enable()
	}()
	r.data = []string{}
	urls, err := r.hy.GetRealUrl(roomId)
	if err != nil {
		r.Alert(err.Error())
		return
	}
	for _, v := range urls {
		r.data = append(r.data, v.Url)
	}
}

func (r *HuYa) Alert(msg string) {
	dialog.ShowInformation("提示", msg, r.window)
}

func main() {
	hy.window.SetContent(container.NewBorder(container.NewVBox(widget.NewLabel("直播间"),
		widget.NewSeparator(), hy.entry, hy.button), nil, nil, nil, hy.list))
	hy.window.Resize(fyne.NewSize(640, 480))
	hy.window.ShowAndRun()
}
