package app

import "catya/api"

type Room struct {
	Id      string          `json:"id"`
	Remark  string          `json:"remark"`
	Count   int             `json:"count"`
	Status  int             `json:"-"`
	RealUrl []api.ResultUrl `json:"-"`
}

type RoomList []*Room

func (r RoomList) Len() int {
	return len(r)
}

func (r RoomList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RoomList) Less(i, j int) bool {
	return r[i].Count > r[j].Count
}
