package app

type Room struct {
	Id     string
	Remark string
	Count  int
	Status int
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
