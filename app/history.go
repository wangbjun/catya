package app

import (
	"catya/api"
	"catya/theme"
	"encoding/json"
	"fyne.io/fyne/v2/widget"
	"log"
	"sort"
	"time"
	"unicode/utf8"
)

type History struct {
	app   *App
	rooms api.Rooms
}

// Load 加载历史访问记录
func (r *History) Load() {
	rooms := api.Rooms{{Id: "lpl", Name: "LPL赛事"}, {Id: "991111", Name: "TheShy"}}
	config := r.app.fyne.Preferences().String("recents")
	if len(config) != 0 {
		err := json.Unmarshal([]byte(config), &rooms)
		if err != nil {
			log.Printf("load history room failed: %s", err)
		}
	}
	r.rooms = rooms

	sort.Sort(r.rooms)

	r.updateHistory()

	go r.updateStatus()
}

// Add 添加访问记录
func (r *History) Add(room *api.Room) {
	var isExisted = false
	for _, item := range r.rooms {
		if item.Id == room.Id {
			item.Count++
			isExisted = true
			break
		}
	}
	if !isExisted {
		r.rooms = append(r.rooms, &api.Room{Id: room.Id, Name: room.Name, Status: 1})
	}
	r.update()
	r.save()
}

// Get 获取room信息
func (r *History) Get(roomId string) *api.Room {
	for _, v := range r.rooms {
		if v.Id == roomId {
			return v
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
	r.updateHistory()
}

func (r *History) updateHistory() {
	for _, v := range r.rooms {
		vv := v
		name := vv.Name
		if name == "" {
			name = vv.Id
		}
		statusIcon := theme.ResourceOfflineSvg
		if vv.Status == 1 {
			statusIcon = theme.ResourceOnlineSvg
		}
		if utf8.RuneCountInString(name) > 8 {
			name = string([]rune(name)[:8]) + "..."
		}
		bt := widget.NewButtonWithIcon(name, statusIcon, func() {
			r.app.submit(vv.Id)
		})
		bt.Alignment = widget.ButtonAlignLeading
		r.app.historyList.Add(bt)
	}
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
			roomInfo, err := r.app.api.GetLiveUrl(room.Id)
			if err != nil {
				log.Printf("update status error: [%s] %s", room.Name, err)
				continue
			}
			room.Urls = roomInfo.Urls
			if len(roomInfo.Urls) > 0 {
				room.Status = 1
			} else {
				room.Status = 0
			}
			if i%5 == 0 || i == len(r.rooms)-1 {
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
	r.app.window.Content().Refresh()
}
