package api

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
		urlStr := fmt.Sprintf("%s/%s.%s?%s",
			value.Get("sFlvUrl").String(),
			value.Get("sStreamName").String(),
			value.Get("sFlvUrlSuffix").String(),
			parseAntiCode(value.Get("sHlsAntiCode").String(), getAnonymousUid(), value.Get("sStreamName").String()))
		urls = append(urls, urlStr)
		return true
	})
	return &Room{
		Urls: urls,
		Name: nickName,
	}, nil
}

func parseAntiCode(anticode, uid, streamName string) string {
	qr, err := url.ParseQuery(anticode)
	if err != nil {
		return ""
	}
	uidInt, _ := strconv.Atoi(uid)
	qr.Set("ver", "1")
	qr.Set("sv", "2110211124")
	qr.Set("seqid", strconv.FormatInt(time.Now().Unix()*1000+int64(uidInt), 10))
	qr.Set("uid", uid)
	qr.Set("uuid", strconv.Itoa(getUuid()))
	ss := MD5([]byte(fmt.Sprintf("%s|%s|%s", qr.Get("seqid"), qr.Get("ctype"), qr.Get("t"))))

	decodeString, _ := base64.StdEncoding.DecodeString(qr.Get("fm"))
	fm := string(decodeString)
	fm = strings.ReplaceAll(fm, "$0", qr.Get("uid"))
	fm = strings.ReplaceAll(fm, "$1", streamName)
	fm = strings.ReplaceAll(fm, "$2", ss)
	fm = strings.ReplaceAll(fm, "$3", qr.Get("wsTime"))

	qr.Del("fm")
	qr.Set("wsSecret", MD5([]byte(fm)))
	if qr.Has("txyp") {
		qr.Del("txyp")
	}
	return qr.Encode()
}

func getAnonymousUid() string {
	urlStr := "https://udblgn.huya.com/web/anonymousLogin"
	body := "{\n        \"appId\": 5002,\n        \"byPass\": 3,\n        \"context\": \"\",\n        \"version\": \"2.4\",\n        \"data\": {}\n    }"
	resp, err := http.Post(urlStr, "application/json", strings.NewReader(body))
	if err != nil {
		return ""
	}
	result, _ := ioutil.ReadAll(resp.Body)
	return gjson.Parse(string(result)).Get("data.uid").String()
}

func getUuid() int {
	now := time.Now().Unix()
	random := int64(rand.Intn(1000))
	return int((now%10000000000*1000 + random) % 4294967295)
}

func MD5(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}
