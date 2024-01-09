package app

import (
	"catya/api"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
	"log"
	"sort"
	"time"
	"unicode/utf8"
)

type History struct {
	app   *App
	rooms api.Rooms
}

// LoadConfig 加载历史访问记录
func (r *History) LoadConfig() {
	rooms := api.Rooms{{Id: "lpl", Name: "LPL赛事"}, {Id: "991111", Name: "TheShy"}}
	config := r.app.fyne.Preferences().String("recents")
	if len(config) != 0 {
		err := json.Unmarshal([]byte(config), &rooms)
		if err != nil {
			log.Printf("load history room failed: %s", err)
		}
	}
	r.rooms = rooms

	r.update()

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
		room.Status = 1
		r.rooms = append(r.rooms, room)
		r.update()
	}
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

// Delete 删除一个room信息
func (r *History) Delete(roomId string) {
	var result = api.Rooms{}
	for _, v := range r.rooms {
		if v.Id == roomId {
			continue
		}
		result = append(result, v)
	}
	r.rooms = result
}

func (r *History) update() {
	r.app.historyList.RemoveAll()
	for _, room := range r.rooms {
		name := room.Name
		if name == "" {
			name = room.Id
		}
		if utf8.RuneCountInString(name) > 8 {
			name = string([]rune(name)[:8]) + "..."
		}
		uri, err := storage.ParseURI(room.Screenshot)
		if err != nil {
			fmt.Printf("parse image url failed:%s\n", err)
			continue
		}
		image := canvas.NewImageFromURI(uri)
		image.FillMode = canvas.ImageFillOriginal

		roomId := room.Id
		card := NewTappedCard(name, room.Description, image, func() {
			r.app.submit(roomId)
		}, func() {
			r.app.remove(roomId)
		})
		card.Resize(fyne.Size{
			Width:  255,
			Height: 200,
		})
		r.app.historyList.Add(card)
	}
	r.app.historyList.Refresh()
}

// 自动更新直播间状态
func (r *History) updateStatus() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("something error happend: %s\n", err)
		}
	}()
	ticker := time.NewTicker(time.Minute * 1)
	for {
		for i, room := range r.rooms {
			roomInfo, err := r.app.api.GetRealUrl(room.Id)
			if err != nil {
				log.Printf("update status error: [%s] %s", room.Name, err)
				continue
			}
			room.Urls = roomInfo.Urls
			room.Screenshot = roomInfo.Screenshot
			room.Description = roomInfo.Description
			if len(roomInfo.Urls) > 0 {
				room.Status = 1
			} else {
				room.Status = 0
			}
			log.Printf("update status success: [%s]", room.Name)
			if len(r.rooms) > 10 || (i+1)%10 == 0 {
				r.update()
			}
		}
		sort.Sort(r.rooms)
		r.update()
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
