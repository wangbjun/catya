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
	"math/rand"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type App struct {
	api     api.Huya
	recents []Room

	window fyne.Window
	roomId *widget.Entry
	remark *widget.Entry
	button *widget.Button
	list   *fyne.Container
}

type Room struct {
	RoomId string
	Remark string
}

func New() *App {
	a := app.NewWithID("catya")
	a.Settings().SetTheme(&theme.MyTheme{})
	return &App{
		api:     api.New(),
		window:  a.NewWindow("Catya"),
		recents: []Room{{"lpl", "LPL"}, {"s4k", "LPL 4K"}, {"991111", "TheShy"}},
	}
}

func (app *App) Run() {
	app.init()
	app.window.SetContent(
		container.NewBorder(
			container.NewVBox(
				app.roomId, app.remark,
				container.NewGridWithColumns(3, layout.NewSpacer(), app.button, layout.NewSpacer()),
				widget.NewSeparator(),
			),
			nil,
			nil,
			nil,
			widget.NewCard("", "最近访问记录，双击可以快速打开：", container.NewGridWithRows(9, app.list)),
		))
	app.window.Resize(fyne.NewSize(640, 480))
	app.window.CenterOnScreen()
	app.window.ShowAndRun()
}

func (app *App) init() {
	rand.Seed(time.Now().UnixNano())

	app.roomId = widget.NewEntry()
	app.roomId.PlaceHolder = "请输入直播间地址或ID，比如：https://www.huya.com/991111、991111"
	app.roomId.OnSubmitted = func(roomId string) {
		app.submit(roomId)
	}
	app.remark = widget.NewEntry()
	app.remark.PlaceHolder = "可选备注，比如：TheShy"

	app.button = widget.NewButton("查询&打开", func() {
		app.submit("")
	})
	app.list = container.NewHBox()
	recent, err := app.loadRecent()
	if err == nil && len(recent) > 0 {
		app.recents = recent
	}
	for _, v := range app.recents {
		vv := v
		remark := vv.Remark
		if remark == "" {
			remark = vv.RoomId
		}
		app.list.Add(widget.NewButton(remark, func() {
			app.submit(vv.RoomId)
		}))
	}
}

func (app *App) submit(roomId string) {
	if roomId == "" {
		roomId = strings.TrimSpace(app.roomId.Text)
	}
	if roomId == "" {
		app.alert("请输入直播间地址或ID")
		return
	}
	parse, err := url.Parse(roomId)
	if err == nil && parse.Path != "" {
		roomId = strings.Trim(parse.Path, "/")
	}
	remark := app.remark.Text
	if remark == "" {
		remark = roomId
	}
	app.button.Text = "查询中......"
	app.button.Disable()
	defer func() {
		app.button.Text = "查询&打开"
		app.button.Enable()
	}()
	urls, err := app.api.GetRealUrl(roomId)
	if err != nil {
		app.alert(err.Error())
		return
	}
	randUrl := urls[rand.Intn(len(urls)-1)]
	var isExisted = false
	for _, recent := range app.recents {
		if recent.RoomId == roomId {
			isExisted = true
			break
		}
	}
	if !isExisted {
		num := len(app.list.Objects)
		// 最多保存最近100个记录
		if num == 100 {
			app.list.Remove(app.list.Objects[0])
			app.recents = app.recents[num-99:]
		}
		app.list.Add(widget.NewButton(remark, func() {
			app.roomId.SetText(remark)
			app.button.Tapped(nil)
		}))
		app.recents = append(app.recents, Room{RoomId: roomId, Remark: app.remark.Text})
		app.saveRecent()
	}
	app.window.Clipboard().SetContent(randUrl.Url)
	err = exec.Command("smplayer", randUrl.Url).Start()
	if err != nil {
		err = exec.Command("mpv", randUrl.Url).Start()
	}
	if err != nil {
		app.alert("打开播放器失败，请确认是否安装smplayer、mpv，并确保在终端里面可以成功调用！")
		app.alert("直播地址已复制到粘贴板，也可以手动打开播放器播放！")
	}
	time.Sleep(time.Second)
}

func (app *App) alert(msg string) {
	info := dialog.NewInformation("提示", msg, app.window)
	info.Show()
}

func (app *App) saveRecent() {
	text, err := json.Marshal(&app.recents)
	if err != nil {
		return
	}
	fyne.CurrentApp().Preferences().SetString("recents", string(text))
}

func (app *App) loadRecent() ([]Room, error) {
	recents := fyne.CurrentApp().Preferences().String("recents")
	var content []Room
	err := json.Unmarshal([]byte(recents), &content)
	if err != nil {
		return nil, err
	}
	return content, nil
}
