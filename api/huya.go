package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Huya struct {
	httpClient http.Client
}

func New() Huya {
	return Huya{
		httpClient: http.Client{Timeout: time.Second * 5},
	}
}

// GetRealUrl 解析虎牙直播间地址
func (r *Huya) GetRealUrl(roomId string) (*Room, error) {
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
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	reg := regexp.MustCompile("<script> window.HNF_GLOBAL_INIT = (.*)</script>")
	matches := reg.FindStringSubmatch(string(result))
	if matches == nil || len(matches) < 2 {
		return nil, errors.New("查询失败")
	}
	return r.extractInfo(matches[1])
}

func (r *Huya) extractInfo(content string) (*Room, error) {
	parse := gjson.Parse(content)

	var (
		//直播间配置
		nickName    = parse.Get("roomInfo.tProfileInfo.sNick").String()
		description = parse.Get("roomInfo.tLiveInfo.sIntroduction").String()
		screenshot  = parse.Get("roomInfo.tLiveInfo.sScreenshot").String()
		streamInfo  = parse.Get("roomInfo.tLiveInfo.tLiveStreamInfo.vStreamInfo.value")
	)
	if description == "" {
		description = "主播暂不在直播哦～"
	}
	if screenshot == "" {
		screenshot = "https://a.msstatic.com/huya/main/assets/img/default/338x190.jpg"
	}

	var urls []string
	streamInfo.ForEach(func(key, value gjson.Result) bool {
		sFlvUrl := value.Get("sFlvUrl").String()
		if strings.Contains(sFlvUrl, "huyalive") { //视频源有问题
			return true
		}
		urlStr := fmt.Sprintf("%s/%s.%s?%s",
			sFlvUrl,
			value.Get("sStreamName").String(),
			value.Get("sFlvUrlSuffix").String(),
			parseAntiCode(value.Get("sFlvAntiCode").String(), value.Get("sStreamName").String()))
		urls = append(urls, urlStr)
		return true
	})
	return &Room{
		Urls:        urls,
		Name:        nickName,
		Screenshot:  screenshot,
		Description: description,
	}, nil
}

func parseAntiCode(antiCode, streamName string) string {
	qr, err := url.ParseQuery(antiCode)
	if err != nil {
		return ""
	}

	t := "0"
	f := strconv.FormatInt(time.Now().UnixNano()/100, 10)
	wsTime := qr.Get("wsTime")

	decodeString, _ := base64.StdEncoding.DecodeString(qr.Get("fm"))
	fm := string(decodeString)
	fm = strings.ReplaceAll(fm, "$0", t)
	fm = strings.ReplaceAll(fm, "$1", streamName)
	fm = strings.ReplaceAll(fm, "$2", f)
	fm = strings.ReplaceAll(fm, "$3", wsTime)

	return fmt.Sprintf("wsSecret=%s&wsTime=%s&u=%s&seqid=%s&txyp=%s&fs=%s&sphdcdn=%s&sphdDC=%s&sphd=%s&u=0&t=100&ratio=0",
		MD5([]byte(fm)), wsTime, t, f, qr.Get("txyp"), qr.Get("fs"), qr.Get("sphdcdn"), qr.Get("sphdDC"), qr.Get("sphd"))
}
