package api

type Room struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Screenshot  string   `json:"screenshot"`
	Count       int      `json:"count"`
	Status      int      `json:"status"`
	Urls        []string `json:"urls"`
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
