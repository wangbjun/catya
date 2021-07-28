package app

import (
	"catya/api"
	"catya/theme"
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"net/url"
	"os/exec"
	"strings"
)

type App struct {
	window   fyne.Window
	entry    *widget.Entry
	button   *widget.Button
	list     *widget.List
	top      *fyne.Container
	api      api.Huya
	dataList []string
	recents  []string
}

func New() *App {
	a := app.NewWithID("catya")
	a.Settings().SetTheme(&theme.MyTheme{})
	return &App{
		api:      api.New(),
		dataList: []string{},
		window:   a.NewWindow("Catya"),
		recents:  []string{"lpl", "s4k"},
	}
}

func (app *App) Run() {
	app.init()
	app.window.SetContent(
		container.NewBorder(
			container.NewVBox(
				container.NewHBox(widget.NewLabel("最近访问直播间："), app.top),
				app.entry,
				container.NewGridWithColumns(3, layout.NewSpacer(), app.button, layout.NewSpacer()),
				widget.NewSeparator(),
			),
			nil,
			nil,
			nil,
			app.list,
		))
	app.window.Resize(fyne.NewSize(640, 480))
	app.window.CenterOnScreen()
	app.window.ShowAndRun()
}

func (app *App) init() {
	entry := widget.NewEntry()
	entry.PlaceHolder = "请输入直播间地址或房间号，比如：https://www.huya.com/lpl、lpl"
	app.entry = entry

	app.button = widget.NewButton("提交", app.submit)
	list := widget.NewList(
		func() int {
			return len(app.dataList)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			data := app.dataList[i]
			textCount := int(o.Size().Width) / 8
			if textCount >= len(data) {
				o.(*widget.Label).SetText(data)
			} else {
				o.(*widget.Label).SetText(data[:textCount] + "...")
			}
		})
	list.OnSelected = func(i widget.ListItemID) {
		// 选中复制地址到粘贴板，自动打开播放器
		app.window.Clipboard().SetContent(app.dataList[i])
		err := exec.Command("smplayer", app.dataList[i]).Start()
		if err != nil {
			exec.Command("mpv", app.dataList[i]).Start()
		}
	}
	app.list = list
	app.top = container.NewHBox()
	recent, err := app.loadRecent()
	if err == nil && len(recent) > 0 {
		app.recents = recent
	}
	for _, v := range app.recents {
		vv := v
		app.top.Add(widget.NewButton(vv, func() {
			app.entry.SetText(vv)
			app.button.Tapped(nil)
		}))
	}
}

func (app *App) submit() {
	roomId := strings.TrimSpace(app.entry.Text)
	if roomId == "" {
		app.alert("请输入直播房间号")
		return
	}
	parse, err := url.Parse(roomId)
	if err == nil && parse.Path != "" {
		roomId = strings.Trim(parse.Path, "/")
	}
	app.button.Text = "查询中......"
	app.button.Disable()
	defer func() {
		app.button.Text = "提交"
		app.button.Enable()
	}()
	app.dataList = []string{}
	urls, err := app.api.GetRealUrl(roomId)
	if err != nil {
		app.alert(err.Error())
		return
	}
	for _, v := range urls {
		app.dataList = append(app.dataList, v.Url)
	}
	app.list.Unselect(0)

	var isExisted = false
	for _, recent := range app.recents {
		if recent == roomId {
			isExisted = true
			break
		}
	}
	if !isExisted {
		num := len(app.top.Objects)
		// 最多保存最近8个记录
		if num == 8 {
			app.top.Remove(app.top.Objects[0])
			app.recents = app.recents[num-7:]
		}
		app.top.Add(widget.NewButton(roomId, func() {
			app.entry.SetText(roomId)
			app.button.Tapped(nil)
		}))
		app.recents = append(app.recents, roomId)
		app.saveRecent()
	}
}

func (app *App) alert(msg string) {
	dialog.ShowInformation("提示", msg, app.window)
}

func (app *App) saveRecent() {
	text, err := json.Marshal(&app.recents)
	if err != nil {
		return
	}
	fyne.CurrentApp().Preferences().SetString("recents", string(text))
}

func (app *App) loadRecent() ([]string, error) {
	recents := fyne.CurrentApp().Preferences().String("recents")
	var content []string
	err := json.Unmarshal([]byte(recents), &content)
	if err != nil {
		return nil, err
	}
	return content, nil
}
