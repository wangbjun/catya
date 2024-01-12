package api

import (
	"fmt"
	"testing"
)

func TestGetRealUrl(t *testing.T) {
	huya := New()
	roomInfo, err := huya.GetRealUrl("lpl")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", roomInfo.Name)
	for _, url := range roomInfo.Urls {
		fmt.Printf("%v\n", url)
	}
}
