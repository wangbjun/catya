package app

import (
	"catya/theme"
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"log"
	"sort"
	"time"
)

type History struct {
	app   *App
	rooms Rooms
}

// Load 加载历史访问记录
func (r *History) Load() {
	rooms := Rooms{{Id: "lpl", Remark: "LPL"}, {Id: "s4k", Remark: "LPL 4K"}, {Id: "991111", Remark: "TheShy"}}
	config := r.app.fyne.Preferences().String("recents")
	if len(config) != 0 {
		err := json.Unmarshal([]byte(config), &rooms)
		if err != nil {
			log.Printf("load history room failed: %s", err)
		}
	}
	r.rooms = rooms

	r.refreshWidget()

	go r.updateStatus()
}

// Add 添加访问记录
func (r *History) Add(room Room) {
	var isExisted = false
	for _, item := range r.rooms {
		if item.Id == room.Id || item.Remark == room.Remark {
			item.Count++
			item.Remark = room.Remark
			isExisted = true
			break
		}
	}
	if !isExisted {
		r.rooms = append(r.rooms, &Room{Id: room.Id, Remark: room.Remark, Status: 1})
	}
	r.update()
	r.save()
}

// Get 获取room信息
func (r *History) Get(roomId string) []string {
	for _, v := range r.rooms {
		if v.Id == roomId {
			return v.Url
		}
	}
	return nil
}

// update 更新访问记录
func (r *History) update() {
	for { // 清除旧记录
		if len(r.app.historyList.Objects) == 0 {
			break
		}
		for _, v := range r.app.historyList.Objects {
			r.app.historyList.Remove(v)
		}
	}
	r.refreshWidget()
}

func (r *History) refreshWidget() {
	var (
		index = 0
		width = 0.0
		list  = make([]*fyne.Container, len(r.rooms))
	)
	for _, v := range r.rooms {
		vv := v
		name := vv.Remark
		if name == "" {
			name = vv.Id
		}
		statusIcon := theme.ResourceOfflineSvg
		if vv.Status == 1 {
			statusIcon = theme.ResourceOnlineSvg
		}
		bt := widget.NewButtonWithIcon(name, statusIcon, func() {
			r.app.submitHistory(*vv)
		})
		if list[index] == nil {
			list[index] = container.NewHBox()
			r.app.historyList.Add(list[index])
		}
		list[index].Add(bt)
		width += float64(bt.Size().Width)
		if width >= 500 {
			index++
			width = 0
		}
	}
	r.app.window.Content().Refresh()
}

// 自动更新直播间状态
func (r *History) updateStatus() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("something error happend: %s\n", err)
		}
	}()
	ticker := time.NewTicker(time.Minute * 2)
	for {
		for i, room := range r.rooms {
			liveUrl, err := r.app.api.GetLiveUrl(room.Id)
			if err != nil {
				log.Printf("update status error: %s => %s", room.Remark, err)
				continue
			}
			if len(liveUrl) == 0 {
				log.Printf("result url empty")
				continue
			}
			room.Url = liveUrl
			room.Status = 1
			if i%5 == 0 {
				r.update()
				time.Sleep(time.Millisecond * 200)
			}
		}
		sort.Sort(r.rooms)
		r.save()
		<-ticker.C
	}
}

// save 保存访问记录
func (r *History) save() {
	text, err := json.Marshal(&r.rooms)
	if err != nil {
		return
	}
	r.app.fyne.Preferences().SetString("recents", string(text))
}
