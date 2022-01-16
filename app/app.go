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
	"sort"
	"strings"
	"time"
)

type App struct {
	api     api.Huya
	recents RoomList

	window  fyne.Window
	roomId  *widget.Entry
	remark  *widget.Entry
	button  *widget.Button
	history *fyne.Container
}

func New() *App {
	a := app.NewWithID("catya")
	a.Settings().SetTheme(&theme.MyTheme{})
	return &App{
		api:     api.New(),
		window:  a.NewWindow("Catya"),
		history: container.NewVBox(),
	}
}

func (app *App) Run() {
	app.setUp()
	app.window.Resize(fyne.NewSize(640, 480))
	app.window.CenterOnScreen()
	app.window.ShowAndRun()
}

func (app *App) setUp() {
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
	app.setUpRecents()
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
			widget.NewCard("", "最近访问记录，双击可以快速打开：", app.history),
		))
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
	app.openRoom(Room{roomId, remark, 0})
}

func (app *App) shotcut(room Room) {
	app.button.Text = "查询中......"
	app.button.Disable()
	defer func() {
		app.button.Text = "查询&打开"
		app.button.Enable()
	}()
	app.openRoom(room)
}

func (app *App) openRoom(room Room) {
	urls, err := app.api.GetRealUrl(room.Id)
	if err != nil {
		app.alert(err.Error())
		return
	}
	randUrl := urls[rand.Intn(len(urls)-1)]
	var isExisted = false
	for _, recent := range app.recents {
		if recent.Id == room.Id || recent.Remark == room.Remark {
			recent.Count++
			recent.Remark = room.Remark
			isExisted = true
			break
		}
	}
	if !isExisted {
		app.recents = append(app.recents, &Room{Id: room.Id, Remark: room.Remark})
	}
	app.updateRecents()
	app.window.Clipboard().SetContent(randUrl.Url)
	err = exec.Command("smplayer", randUrl.Url).Start()
	if err != nil {
		err = exec.Command("mpv", randUrl.Url).Start()
	}
	if err != nil {
		app.alert("打开播放器失败，请确认是否安装smplayer、mpv，并确保在终端里面可以成功调用！")
		app.alert("直播地址已复制到粘贴板，也可以手动打开播放器播放！")
	}
	go app.saveRecent()
	time.Sleep(time.Second)
}

// 设置最近访问记录
func (app *App) setUpRecents() {
	recents, err := app.loadRecent()
	if err != nil {
		return
	}
	sort.Sort(recents)
	pos := 0
	length := float32(0.0)
	list := make([]*fyne.Container, len(recents)/5+1)
	for _, v := range recents {
		vv := v
		remark := vv.Remark
		if remark == "" {
			remark = vv.Id
		}
		bt := widget.NewButton(remark, func() {
			app.shotcut(*vv)
		})
		if list[pos] == nil {
			list[pos] = container.NewHBox()
			app.history.Add(list[pos])
		}
		list[pos].Add(bt)
		length += bt.Size().Width
		if length >= 500 {
			pos++
			length = 0
		}
	}
	app.recents = recents
	app.saveRecent()
}

// 更新最近访问记录
func (app *App) updateRecents() {
	sort.Sort(app.recents)
	pos := 0
	length := float32(0.0)
	list := make([]*fyne.Container, len(app.recents)/5+1)
	for _, v := range app.history.Objects {
		app.history.Remove(v)
	}
	for _, v := range app.recents {
		vv := v
		remark := vv.Remark
		if remark == "" {
			remark = vv.Id
		}
		bt := widget.NewButton(remark, func() {
			app.shotcut(*vv)
		})
		if list[pos] == nil {
			list[pos] = container.NewHBox()
			app.history.Add(list[pos])
		}
		list[pos].Add(bt)
		length += bt.Size().Width
		if length >= 500 {
			pos++
			length = 0
		}
	}
}

// 加载最近访问记录配置
func (app *App) loadRecent() (RoomList, error) {
	recents := fyne.CurrentApp().Preferences().String("recents")
	if len(recents) == 0 {
		return RoomList{{"lpl", "LPL", 0}, {"s4k",
			"LPL 4K", 0}, {"991111", "TheShy", 0}}, nil
	}
	var content []*Room
	err := json.Unmarshal([]byte(recents), &content)
	if err != nil {
		return nil, err
	}
	return content, nil
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
