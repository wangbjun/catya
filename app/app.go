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
	"log"
	"math/rand"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type App struct {
	fyne         fyne.App
	api          api.LiveApi
	history      History
	window       fyne.Window
	inputRoom    *widget.Entry
	inputRemark  *widget.Entry
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (app *App) Run() {
	app.setUp()
	app.window.Resize(fyne.NewSize(640, 420))
	app.window.CenterOnScreen()
	app.window.ShowAndRun()
}

func (app *App) setUp() {
	app.inputRoom = widget.NewEntry()
	app.inputRoom.PlaceHolder = "请输入直播间地址或ID，比如：https://www.huya.com/991111、991111"
	app.inputRoom.OnSubmitted = func(roomId string) {
		app.submit(roomId)
	}
	app.inputRemark = widget.NewEntry()
	app.inputRemark.PlaceHolder = "可选备注，比如：TheShy"

	app.submitButton = widget.NewButton("查询&打开", func() {
		app.submit("")
	})
	app.history.Load()
	app.window.SetContent(
		container.NewBorder(
			container.NewVBox(
				app.inputRoom, app.inputRemark,
				container.NewGridWithColumns(3, layout.NewSpacer(), app.submitButton, layout.NewSpacer()),
				widget.NewSeparator(),
			),
			nil,
			nil,
			nil,
			widget.NewCard("", "最近访问记录，双击可以快速打开：", app.historyList),
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
	remark := app.inputRemark.Text
	if remark == "" {
		remark = roomId
	}
	app.submitButton.Text = "查询中......"
	app.submitButton.Disable()
	defer func() {
		app.submitButton.Text = "查询&打开"
		app.submitButton.Enable()
	}()
	app.open(Room{Id: roomId, Remark: remark})
}

func (app *App) submitHistory(room Room) {
	app.submitButton.Text = "查询中......"
	app.submitButton.Disable()
	defer func() {
		app.submitButton.Text = "查询&打开"
		app.submitButton.Enable()
	}()
	app.open(room)
}

func (app *App) open(room Room) {
	log.Printf("open room: %s", room.Remark)
	// 获取直播地址
	urls := app.history.Get(room.Id)
	if urls == nil {
		urls, _ = app.api.GetLiveUrl(room.Id)
	}
	if urls == nil || len(urls) == 0 {
		app.alert("未开播或不存在")
		return
	}
	app.history.Add(room)
	randUrl := urls[rand.Intn(len(urls)-1)]
	app.window.Clipboard().SetContent(randUrl)
	err := exec.Command("smplayer", randUrl).Start()
	if err != nil {
		err = exec.Command("mpv", randUrl).Start()
	}
	if err != nil {
		app.alert("直播地址已复制到粘贴板，可以手动打开播放器播放！")
		app.alert("播放失败，请确认是否安装smplayer或mpv，并确保在终端可以调用！")
	}
}

func (app *App) alert(msg string) {
	info := dialog.NewInformation("提示", msg, app.window)
	info.Show()
}
