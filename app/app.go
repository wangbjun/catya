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
	"log"
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

	Window  fyne.Window
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
		Window:  a.NewWindow("Catya"),
		history: container.NewVBox(),
	}
}

func (app *App) SetUp() {
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
	app.initRecents()
	app.Window.SetContent(
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
	go app.updateStatus()
}

// 自动更新直播间状态
func (app *App) updateStatus() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("something happend: %s\n", err)
		}
	}()
	ticker := time.NewTicker(time.Minute * 3)
	for {
		for _, v := range app.recents {
			realUrl, err := app.api.GetRealUrl(v.Id)
			if err != nil {
				log.Printf("update live status error: %s => %s", v.Remark, err)
				continue
			}
			if len(realUrl) != 0 {
				v.Status = 1
				v.RealUrl = realUrl
				continue
			}
			time.Sleep(time.Millisecond * 300)
		}
		sort.Sort(app.recents)
		app.updateRecents()
		app.Window.Content().Refresh()
		log.Println("update live status finished")
		<-ticker.C
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
	app.openRoom(Room{Id: roomId, Remark: remark})
}

func (app *App) shortcut(room Room) {
	app.button.Text = "查询中......"
	app.button.Disable()
	defer func() {
		app.button.Text = "查询&打开"
		app.button.Enable()
	}()
	app.openRoom(room)
}

func (app *App) openRoom(room Room) {
	log.Printf("open room: %s", room.Id)
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
	app.saveRecent()
	app.updateRecents()
	// 获取直播地址
	urls := app.getFromRecents(room.Id)
	if urls == nil {
		urls, _ = app.api.GetRealUrl(room.Id)
	}
	if urls == nil || len(urls) == 0 {
		app.alert("未开播或不存在")
		return
	}
	randUrl := urls[rand.Intn(len(urls)-1)]
	app.Window.Clipboard().SetContent(randUrl.Url)
	err := exec.Command("smplayer", randUrl.Url).Start()
	if err != nil {
		err = exec.Command("mpv", randUrl.Url).Start()
	}
	if err != nil {
		app.alert("直播地址已复制到粘贴板，可以手动打开播放器播放！")
		app.alert("播放失败，请确认是否安装smplayer或mpv，并确保在终端可以调用！")
	}
	time.Sleep(time.Second)
}

// 获取roomId
func (app *App) getFromRecents(roomId string) []api.ResultUrl {
	for _, v := range app.recents {
		if v.Id == roomId {
			return v.RealUrl
		}
	}
	return nil
}

// 设置最近访问记录
func (app *App) initRecents() {
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
		bt := widget.NewButtonWithIcon(remark, theme.ResourceOfflineSvg, func() {
			app.shortcut(*vv)
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
	pos := 0
	length := float32(0.0)
	list := make([]*fyne.Container, len(app.recents)/5+1)
	for { // 清除旧记录
		if len(app.history.Objects) == 0 {
			break
		}
		for _, v := range app.history.Objects {
			app.history.Remove(v)
		}
	}
	for _, v := range app.recents {
		vv := v
		remark := vv.Remark
		if remark == "" {
			remark = vv.Id
		}
		statusIcon := theme.ResourceOfflineSvg
		if vv.Status == 1 {
			statusIcon = theme.ResourceOnlineSvg
		}
		bt := widget.NewButtonWithIcon(remark, statusIcon, func() {
			app.shortcut(*vv)
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
		return RoomList{{Id: "lpl", Remark: "LPL"}, {Id: "s4k",
			Remark: "LPL 4K"}, {Id: "991111", Remark: "TheShy"}}, nil
	}
	var content []*Room
	err := json.Unmarshal([]byte(recents), &content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (app *App) alert(msg string) {
	info := dialog.NewInformation("提示", msg, app.Window)
	info.Show()
}

func (app *App) saveRecent() {
	text, err := json.Marshal(&app.recents)
	if err != nil {
		return
	}
	fyne.CurrentApp().Preferences().SetString("recents", string(text))
}
