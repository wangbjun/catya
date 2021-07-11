package api

import (
	"errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Huya struct {
	httpClient http.Client
}

func New() Huya {
	return Huya{
		httpClient: http.Client{
			Timeout: time.Second * 10,
		},
	}
}

type ResultUrl struct {
	CdnType string
	Url     string
}

func (r Huya) GetRealUrl(roomId string) ([]ResultUrl, error) {
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

	reg := regexp.MustCompile("<script> window.HNF_GLOBAL_INIT = (.*)</script>")
	submatch := reg.FindStringSubmatch(string(result))

	if submatch == nil || len(submatch) < 2 {
		return nil, errors.New("查询失败")
	}
	return extractUrl(submatch[1])
}

func extractUrl(content string) ([]ResultUrl, error) {
	parse := gjson.Parse(content)
	streamInfo := parse.Get("roomInfo.tLiveInfo.tLiveStreamInfo.vStreamInfo.value")
	if !streamInfo.Exists() || len(streamInfo.Array()) == 0 {
		return nil, errors.New("未开播或直播间不存在")
	}
	var result []ResultUrl
	streamInfo.ForEach(func(key, value gjson.Result) bool {
		cdnType := strings.ToLower(value.Get("sCdnType").String())
		urlPart := value.Get("sStreamName").String() + "." + value.Get("sFlvUrlSuffix").String() + "?" + value.Get("sFlvAntiCode").String()
		result = append(result, ResultUrl{
			CdnType: cdnType + "_flv",
			Url:     value.Get("sFlvUrl").String() + "/" + urlPart,
		})
		return true
	})

	return result, nil
}
