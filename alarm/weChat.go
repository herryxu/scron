package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/henryxu/tools/limiter"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	prodEnv = "prod"
)

type weChatAlarm struct {
	webhook string
}

func newWeChatAlarm() IAlarm {
	w := &weChatAlarm{
		webhook: "xxxxx",
	}
	return w
}

func (s *weChatAlarm) SendAlarm(content string, limit ...interface{}) {
	var key string
	if len(limit) >= 2 {
		key = strings.TrimSpace(limit[0].(string))
	}
	//runMode := g.Cfg().GetString("server.RunMode")
	//var envDesc = ""
	//if common.RunMode != prodEnv {
	//	return
	//}
	//content = envDesc + content
	if key == "" {
		s.doSend(content)
		return
	}
	// 1h 1次
	if limiter.CheckLimiter(key, 3600) {
		s.doSend(content)
	}
}

func (s *weChatAlarm) doSend(content string) error {
	data := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": content,
			// "mentioned_list": []string{"@all"}, // @所有人
		},
	}
	js, _ := json.Marshal(data)
	//res, err := g.Client().Post(s.webhook, string(js))
	res, err := httpPost(s.webhook, string(js))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http.doSend.res:%+v", res)
	}
	return nil
}

func (s *weChatAlarm) SetWebhook(w string) {
	s.webhook = w
}

func httpPost(url, data string) (response *http.Response, err error) {
	fmt.Println("HTTP JSON POST URL:", url)

	var jsonData = []byte(data)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
	return
}
