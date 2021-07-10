package huya

import (
	"fmt"
	"testing"
)

func TestGetRealUrl(t *testing.T) {
	huya := New()
	urls, err := huya.GetRealUrl("kaerlol")
	if err != nil {
		panic(err)
	}
	for _, v := range urls {
		fmt.Printf("%v\n", v.Url)
	}
}
