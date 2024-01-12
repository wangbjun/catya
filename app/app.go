package app

import (
	"catya/api"
	"catya/theme"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"math/rand"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

const (
	mpvIPCPath              = "/tmp/mpv_socket"
	preferenceKeyWindowSize = "windowSize"
	preferenceKeyHistory    = "recentsList"
)

type App struct {
	api     api.LiveApi
	history *History
	mpvIPC  net.Conn

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
		fyne:        catya,
		window:      catya.NewWindow("Catya"),
		recentsList: container.New(NewHistoryLayout()),
	}

	application.history = NewHistory(application)
	return application
}

func (app *App) Run() {
	app.SetLayout()
	app.SetSize()
	app.window.CenterOnScreen()
	app.window.ShowAndRun()
}

func (app *App) SetLayout() {
	app.history.LoadConf()

	app.inputName = widget.NewEntry()
	app.inputName.PlaceHolder = "请输入直播间备注名称，可选"

	app.inputRoom = widget.NewEntry()
	app.inputRoom.PlaceHolder = "请输入直播间URL或ID，比如：991111、uzi"
	app.inputRoom.OnSubmitted = func(roomId string) {
		app.Submit(roomId)
		app.inputRoom.SetText("")
	}

	app.submitButton = widget.NewButton("查询&打开", func() {
		app.Submit("")
		app.inputRoom.SetText("")
		app.inputName.SetText("")
	})

	app.window.SetContent(
		container.NewVBox(
			container.New(NewProportionLayout([]float64{0.8, 0.2}), container.NewVBox(app.inputRoom, app.inputName), app.submitButton),
			app.recentsList,
		))
}

func (app *App) SetSize() {
	app.window.SetOnClosed(func() {
		app.SaveSize()
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

func (app *App) Submit(roomId string) {
	if roomId == "" {
		roomId = strings.TrimSpace(app.inputRoom.Text)
	}
	if roomId == "" {
		app.ShowInfo("请输入直播间地址或ID")
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
			app.ShowInfo("查询失败")
			return
		}
	}
	roomInfo.Id = roomId
	if app.inputName.Text != "" {
		roomInfo.Name = app.inputName.Text
	}
	if roomInfo.Name == "" {
		app.ShowInfo("直播间不存在")
		return
	}
	app.history.Add(roomInfo)
	if len(roomInfo.Urls) == 0 {
		app.ShowInfo("主播暂未开播")
		return
	}
	//随机取一个地址
	randUrl := roomInfo.Urls[0]
	if len(roomInfo.Urls) > 1 {
		randUrl = roomInfo.Urls[rand.Intn(len(roomInfo.Urls)-1)]
	}

	title := fmt.Sprintf("%s: %s", roomInfo.Name, roomInfo.Description)
	if app.mpvIPC != nil {
		err = app.SendIPC([]interface{}{"loadfile", randUrl})
		err = app.SendIPC([]interface{}{"set_property", "title", title})
		if err == nil {
			return
		}
		log.Printf("SendIPC error:%s\n", err)
	}
	err = exec.Command("mpv", "--title="+title, "--input-ipc-server="+mpvIPCPath, randUrl).Start()
	if err != nil {
		err = exec.Command("smplayer", randUrl).Start()
	}
	if err != nil {
		app.window.Clipboard().SetContent(randUrl)
		app.ShowInfo("直播地址已复制到粘贴板，可以手动打开播放器播放！")
		app.ShowInfo("播放失败，请确认是否安装smplayer或mpv，并确保在终端可以调用！")
		return
	}
	err = app.InitIPC()
	if err != nil {
		log.Printf("InitIPC error:%s\n", err)
	}
}

func (app *App) RemoveRoom(roomId string) {
	app.history.Delete(roomId)
}

func (app *App) ShowInfo(msg string) {
	info := dialog.NewInformation("提示", msg, app.window)
	info.Show()
}

func (app *App) SaveSize() {
	currentSize := app.window.Canvas().Size()
	app.fyne.Preferences().SetFloatList(preferenceKeyWindowSize, []float64{float64(currentSize.Width), float64(currentSize.Height)})
}

// InitIPC 初始化mpv ipc连接
func (app *App) InitIPC() error {
	var err error
	for i := 0; i < 5; i++ {
		conn, e := net.Dial("unix", mpvIPCPath)
		err = e
		if e != nil {
			time.Sleep(time.Second)
			continue
		}
		app.mpvIPC = conn
		return nil
	}
	return err
}

// SendIPC 给mpv发送ipc命令
func (app *App) SendIPC(command []interface{}) error {
	if app.mpvIPC == nil {
		app.mpvIPC, _ = net.Dial("unix", mpvIPCPath)
		if app.mpvIPC == nil {
			return errors.New("mpv ipc conn is null")
		}
	}
	cmd := MPVCommand{
		Command: command,
	}
	cmdJSON, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	_, err = app.mpvIPC.Write(append(cmdJSON, '\n'))
	return err
}
