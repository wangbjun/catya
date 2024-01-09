package app

import (
	"catya/api"
	"catya/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"math/rand"
	"net/url"
	"os/exec"
	"strings"
)

type App struct {
	fyne         fyne.App
	api          api.LiveApi
	history      History
	window       fyne.Window
	inputRoom    *widget.Entry
	inputName    *widget.Entry
	submitButton *widget.Button
	historyList  *fyne.Container
}

func New(api api.LiveApi) *App {
	catya := app.NewWithID("catya")
	catya.Settings().SetTheme(&theme.MyTheme{})
	application := &App{
		api:         api,
		fyne:        catya,
		window:      catya.NewWindow("Catya"),
		historyList: container.New(NewHistoryLayout()),
	}
	application.history = History{app: application}
	return application
}

func (app *App) Run() {
	app.history.LoadConfig()
	app.setUp()
	app.window.Resize(fyne.NewSize(1025, 725))
	app.window.CenterOnScreen()
	app.window.ShowAndRun()
}

func (app *App) setUp() {
	app.inputName = widget.NewEntry()
	app.inputName.PlaceHolder = "请输入直播间备注名称，可选"
	app.inputRoom = widget.NewEntry()
	app.inputRoom.PlaceHolder = "请输入直播间地址或ID，比如：https://www.huya.com/991111、991111"
	app.inputRoom.OnSubmitted = func(roomId string) {
		app.submit(roomId)
		app.inputRoom.SetText("")
	}
	app.submitButton = widget.NewButton("查询&打开", func() {
		app.submit("")
		app.inputRoom.SetText("")
		app.inputName.SetText("")
	})
	app.window.SetContent(
		container.NewBorder(
			container.NewVBox(
				app.inputRoom,
				app.inputName,
				container.NewGridWithColumns(3, layout.NewSpacer(), app.submitButton, layout.NewSpacer()),
				widget.NewSeparator(),
			),
			nil,
			nil,
			nil,
			app.historyList,
		))
}

func (app *App) submit(roomId string) {
	if roomId == "" {
		roomId = strings.TrimSpace(app.inputRoom.Text)
	}
	if roomId == "" {
		app.alert("请输入直播间地址或ID")
		return
	}
	parse, err := url.Parse(roomId)
	if err == nil && parse.Path != "" {
		roomId = strings.Trim(parse.Path, "/")
	}
	app.submitButton.Text = "查询中......"
	app.submitButton.Disable()
	defer func() {
		app.submitButton.Text = "查询&打开"
		app.submitButton.Enable()
	}()
	roomInfo := app.history.Get(roomId)
	if roomInfo == nil {
		roomInfo, err = app.api.GetRealUrl(roomId)
		if err != nil {
			app.alert("查询失败")
			return
		}
	}
	roomInfo.Id = roomId
	if app.inputName.Text != "" {
		roomInfo.Name = app.inputName.Text
	}
	if roomInfo.Name == "" {
		app.alert("直播间不存在")
		return
	}
	app.history.Add(roomInfo)
	if len(roomInfo.Urls) == 0 {
		app.alert("主播暂未开播")
		return
	}
	//随机取一个地址
	randUrl := roomInfo.Urls[0]
	if len(roomInfo.Urls) > 1 {
		randUrl = roomInfo.Urls[rand.Intn(len(roomInfo.Urls)-1)]
	}

	exec.Command("killall", "mpv").Run()
	err = exec.Command("mpv", "--title="+roomInfo.Name, randUrl).Start()
	if err != nil {
		err = exec.Command("smplayer", randUrl).Start()
	}
	if err != nil {
		app.window.Clipboard().SetContent(randUrl)
		app.alert("直播地址已复制到粘贴板，可以手动打开播放器播放！")
		app.alert("播放失败，请确认是否安装smplayer或mpv，并确保在终端可以调用！")
	}
}

func (app *App) remove(roomId string) {
	app.history.Delete(roomId)
	app.history.update()
}

func (app *App) alert(msg string) {
	info := dialog.NewInformation("提示", msg, app.window)
	info.Show()
}
