package api

import (
	"errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type LiveApi interface {
	GetLiveUrl(string) (*Room, error)
}

type Huya struct {
	httpClient http.Client
}

func New() Huya {
	return Huya{
		httpClient: http.Client{
			Timeout: time.Second * 5,
		},
	}
}

// GetLiveUrl 解析虎牙直播间地址
func (r *Huya) GetLiveUrl(roomId string) (*Room, error) {
	roomUrl := "https://m.huya.com/" + roomId
	request, err := http.NewRequest("GET", roomUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36")
	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	reg := regexp.MustCompile("<script> window.HNF_GLOBAL_INIT = (.*)</script>")
	submatch := reg.FindStringSubmatch(string(result))
	if submatch == nil || len(submatch) < 2 {
		return nil, errors.New("查询失败")
	}
	return extractInfo(submatch[1])
}

func extractInfo(content string) (*Room, error) {
	parse := gjson.Parse(content)
	streamInfo := parse.Get("roomInfo.tLiveInfo.tLiveStreamInfo.vStreamInfo.value")
	nickName := parse.Get("roomInfo.tProfileInfo.sNick").String()
	var urls []string
	streamInfo.ForEach(func(key, value gjson.Result) bool {
		urlPart := value.Get("sStreamName").String() + "." + value.Get("sFlvUrlSuffix").String() + "?" + value.Get("sFlvAntiCode").String()
		urls = append(urls, value.Get("sFlvUrl").String()+"/"+urlPart)
		return true
	})
	return &Room{
		Urls: urls,
		Name: nickName,
	}, nil
}
