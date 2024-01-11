package app

import (
	"catya/api"
	"catya/theme"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"math/rand"
	"net"
	"net/url"
	"os/exec"
	"strings"
)

const (
	preferenceKeyWindowSize = "windowSize"
	preferenceKeyHistory    = "recentsList"
)

type App struct {
	api     api.LiveApi
	history *History
	ipc     chan string

	fyne         fyne.App
	window       fyne.Window
	inputRoom    *widget.Entry
	inputName    *widget.Entry
	submitButton *widget.Button
	recentsList  *fyne.Container
}

type MPVCommand struct {
	Command []interface{} `json:"command"`
}

func New(api api.LiveApi) *App {
	catya := app.NewWithID("catya")
	catya.Settings().SetTheme(&theme.MyTheme{})

	application := &App{
		api:         api,
		ipc:         make(chan string),
		fyne:        catya,
		window:      catya.NewWindow("Catya"),
		recentsList: container.New(NewHistoryLayout()),
	}

	application.history = NewHistory(application)
	return application
}

func (app *App) Run() {
	app.init()
	app.setUpSize()
	app.window.CenterOnScreen()
	app.window.ShowAndRun()
}

func (app *App) init() {
	app.history.LoadConf()

	app.inputName = widget.NewEntry()
	app.inputName.PlaceHolder = "请输入直播间备注名称，可选"

	app.inputRoom = widget.NewEntry()
	app.inputRoom.PlaceHolder = "请输入直播间URL或ID，比如：991111、uzi"
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
		container.NewVBox(
			container.New(NewProportionLayout([]float64{0.8, 0.2}), container.NewVBox(app.inputRoom, app.inputName), app.submitButton),
			app.recentsList,
		))
}

func (app *App) setUpSize() {
	app.window.SetOnClosed(func() {
		app.saveSize()
		app.history.Save()
		app.window.Close()
	})

	windowSize := app.fyne.Preferences().FloatList(preferenceKeyWindowSize)
	if len(windowSize) == 2 {
		app.window.Resize(fyne.NewSize(float32(windowSize[0]), float32(windowSize[1])))
	} else {
		app.window.Resize(fyne.NewSize(1280, 720))
	}
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
	title := fmt.Sprintf("%s: %s", roomInfo.Name, roomInfo.Description)
	err = app.sendIPC([]interface{}{"loadfile", randUrl})
	err = app.sendIPC([]interface{}{"set_property", "title", title})
	if err == nil {
		return
	}
	err = exec.Command("mpv", "--title="+title, "--input-ipc-server=/tmp/mpv_socket", randUrl).Start()
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
}

func (app *App) alert(msg string) {
	info := dialog.NewInformation("提示", msg, app.window)
	info.Show()
}

func (app *App) saveSize() {
	currentSize := app.window.Canvas().Size()
	app.fyne.Preferences().SetFloatList(preferenceKeyWindowSize, []float64{float64(currentSize.Width), float64(currentSize.Height)})
}

// 给mpv发送IPC命令
func (app *App) sendIPC(command []interface{}) error {
	conn, err := net.Dial("unix", "/tmp/mpv_socket")
	if err != nil {
		return err
	}
	defer conn.Close()
	// 构建一个命令
	cmd := MPVCommand{
		Command: command,
	}
	cmdJSON, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(cmdJSON, '\n'))
	return err
}
