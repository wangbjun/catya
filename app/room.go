package app

type Room struct {
	Id     string   `json:"id"`
	Remark string   `json:"remark"`
	Count  int      `json:"count"`
	Status int      `json:"-"`
	Url    []string `json:"-"`
}

type Rooms []*Room

func (r Rooms) Len() int {
	return len(r)
}

func (r Rooms) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Rooms) Less(i, j int) bool {
	return r[i].Count > r[j].Count
}
