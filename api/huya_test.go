package api

import (
	"fmt"
	"testing"
)

func TestGetRealUrl(t *testing.T) {
	huya := New()
	roomInfo, err := huya.GetLiveUrl("lpl")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", roomInfo)
}
