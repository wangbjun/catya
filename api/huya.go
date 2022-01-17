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
	cache      map[string][]ResultUrl
}

var ErrorNotExist = errors.New("未开播或不存在")

func New() Huya {
	return Huya{
		httpClient: http.Client{
			Timeout: time.Second * 5,
		},
		cache: make(map[string][]ResultUrl, 20),
	}
}

type ResultUrl struct {
	CdnType string
	Url     string
}

func (r *Huya) GetRealUrl(roomId string) ([]ResultUrl, error) {
	// 直接从缓存里面取
	cacheUrl, ok := r.cache[roomId]
	if ok {
		return cacheUrl, nil
	}
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
	resultUrl, err := extractUrl(submatch[1])
	if err != nil {
		return nil, err
	}
	r.cache[roomId] = resultUrl
	return resultUrl, nil
}

func extractUrl(content string) ([]ResultUrl, error) {
	parse := gjson.Parse(content)
	streamInfo := parse.Get("roomInfo.tLiveInfo.tLiveStreamInfo.vStreamInfo.value")
	if !streamInfo.Exists() || len(streamInfo.Array()) == 0 {
		return nil, ErrorNotExist
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
