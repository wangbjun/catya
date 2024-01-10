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

func NewHistory(app *App) *History {
	return &History{app: app}
}

// Init 加载历史访问记录
func (m *History) Init() {
	config := m.app.fyne.Preferences().String(preferenceKeyHistory)
	if config == "" {
		m.rooms = api.Rooms{{Id: "lpl", Name: "LPL赛事"}, {Id: "991111", Name: "TheShy", Count: 1000}}
	} else {
		err := json.Unmarshal([]byte(config), &m.rooms)
		if err != nil {
			log.Printf("load history room failed: %s", err)
		}
	}

	m.updateCard()

	go m.updateRoomStatus()
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
		m.updateCard()
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
	m.updateCard()
}

func (m *History) updateCard() {
	m.app.recentsList.RemoveAll()
	for _, room := range m.rooms {
		name := room.Name
		if name == "" {
			name = room.Id
		}
		if utf8.RuneCountInString(name) > 10 {
			name = string([]rune(name)[:10]) + "..."
		}
		if utf8.RuneCountInString(room.Description) > 12 {
			room.Description = string([]rune(room.Description)[:12]) + "..."
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
			m.app.submit(roomId)
		}, func() {
			m.app.remove(roomId)
		})
		card.Resize(fyne.Size{
			Width:  255,
			Height: 200,
		})
		m.app.recentsList.Add(card)
	}
	m.app.recentsList.Refresh()
}

// Save 保存访问记录
func (m *History) Save() {
	text, err := json.Marshal(&m.rooms)
	if err != nil {
		return
	}
	m.app.fyne.Preferences().SetString(preferenceKeyHistory, string(text))
}

// 自动更新直播间状态
func (m *History) updateRoomStatus() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("something error happend: %s\n", err)
		}
	}()
	ticker := time.NewTicker(time.Minute * 1)
	for {
		for i, room := range m.rooms {
			roomInfo, err := m.app.api.GetRealUrl(room.Id)
			if err != nil {
				log.Printf("updateCard status error: [%s] %s", room.Name, err)
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
			log.Printf("updateCard status success: [%s]", room.Name)
			if len(m.rooms) > 10 || (i+1)%10 == 0 {
				m.updateCard()
			}
		}
		log.Println("-----------updateCard status finished-----------")
		sort.Sort(m.rooms)
		m.updateCard()
		<-ticker.C
	}
}
