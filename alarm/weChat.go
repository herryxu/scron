package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/olaola-chat/slp-tools/limiter"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	//ChargeWebhook = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=e55d27d5-62de-4039-a119-eb88a17f219f"
	prodEnv = "prod"
)

type weChatAlarm struct {
	webhook string
}

func newWeChatAlarm() IAlarm {
	w := &weChatAlarm{
		webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=872b1013-e950-49ed-98fe-dd95397482ee",
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
