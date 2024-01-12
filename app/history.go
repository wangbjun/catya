package app

import (
	"catya/api"
	"encoding/json"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
	"log"
	"math/rand"
	"sort"
	"time"
	"unicode/utf8"
)

type History struct {
	app   *App
	rooms api.Rooms
}

func NewHistory(app *App) *History {
	return &History{app: app}
}

// LoadConf 加载历史访问记录
func (m *History) LoadConf() {
	config := m.app.fyne.Preferences().String(preferenceKeyHistory)
	if config == "" {
		m.rooms = api.Rooms{{Id: "lpl", Name: "LPL赛事"}, {Id: "991111", Name: "TheShy", Count: 1000}}
	} else {
		err := json.Unmarshal([]byte(config), &m.rooms)
		if err != nil {
			log.Printf("load history room failed: %s", err)
		}
	}

	m.UpdateCard()

	go m.UpdateRoomStatus()
}

// Add 添加访问记录
func (m *History) Add(room *api.Room) {
	var isExisted = false
	for _, item := range m.rooms {
		if item.Id == room.Id {
			item.Count++
			isExisted = true
			break
		}
	}
	if !isExisted {
		room.Status = 1
		m.rooms = append(m.rooms, room)
		m.UpdateCard()
	}
}

// Get 获取room信息
func (m *History) Get(roomId string) *api.Room {
	for _, v := range m.rooms {
		if v.Id == roomId {
			return v
		}
	}
	return nil
}

// Delete 删除一个room信息
func (m *History) Delete(roomId string) {
	var result = api.Rooms{}
	for _, v := range m.rooms {
		if v.Id == roomId {
			continue
		}
		result = append(result, v)
	}
	m.rooms = result
	m.UpdateCard()
}

func (m *History) UpdateCard() {
	m.app.recentsList.RemoveAll()
	for _, room := range m.rooms {
		id := room.Id
		name := room.Name
		description := room.Description

		if name == "" {
			name = id
		}
		if utf8.RuneCountInString(name) > 10 {
			name = string([]rune(name)[:10]) + "..."
		}
		if utf8.RuneCountInString(description) > 12 {
			description = string([]rune(description)[:12]) + "..."
		}
		uri, err := storage.ParseURI(room.Screenshot)
		if err != nil {
			log.Printf("parse image url failed:%s\n", err)
			continue
		}
		image := canvas.NewImageFromURI(uri)
		image.FillMode = canvas.ImageFillOriginal
		card := NewTappedCard(name, description, image, func() {
			m.app.Submit(id)
		}, func() {
			m.app.RemoveRoom(id)
		})
		m.app.recentsList.Add(card)
	}
	canvas.Refresh(m.app.recentsList)
}

// Save 保存访问记录
func (m *History) Save() {
	text, err := json.Marshal(&m.rooms)
	if err != nil {
		return
	}
	m.app.fyne.Preferences().SetString(preferenceKeyHistory, string(text))
}

// UpdateRoomStatus 自动更新直播间状态
func (m *History) UpdateRoomStatus() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("something error happend: %s\n", err)
		}
	}()
	ticker := time.NewTicker(time.Second * 50)
	for {
		time.Sleep(time.Second * time.Duration(rand.Intn(10)))
		for i, room := range m.rooms {
			roomInfo, err := m.app.api.GetRealUrl(room.Id)
			if err != nil {
				log.Printf("UpdateCard status error: [%s] %s", room.Name, err)
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
			log.Printf("UpdateCard status success: [%s]", room.Name)
			if len(m.rooms) > 10 && (i+1)%10 == 0 {
				m.UpdateCard()
			}
		}
		log.Println("-----------UpdateCard status finished-----------")
		sort.Sort(m.rooms)
		m.UpdateCard()
		<-ticker.C
	}
}
