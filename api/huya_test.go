package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, roomInfo.Name, "虎牙英雄联盟赛事")
}
