package wechat

import (
	"common/config"
	"common/log"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chanxuehong/wechat/mp/message/template"
	"io/ioutil"
	"net/http"
	"time"
)

func Init() {
	var err error
	defer func() {
		if err != nil {
			log.WithError(err).Error("微信通知初始化失败")
			panic(err)
		}
		log.Info("微信通知初始化成功")
	}()
	if config.Data.Wechat == nil {
		err = errors.New("读取微信通知配置失败")
		return
	}
	err = initwechat(config.Data.Wechat)
}

var client = http.Client{
	Timeout: 10 * time.Second,
}

type wxBody struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func Sendwechat(user string, info map[string]interface{}, templateflag int) error {
	templatemsg := template.TemplateMessage2{}
	templatemsg.ToUser = user
	if templateflag == 0 {
		templatemsg.TemplateId = arriveTemplateId
	} else {
		templatemsg.TemplateId = recieveTemplateId
	}

	mini := template.MiniProgram{miniAppId, "/pages/login/login"}
	templatemsg.MiniProgram = &mini
	templatemsg.Data = info
	id, err := template.Send(WechatClient, templatemsg)
	log.LogWithErrorf(err, "Sendwechat fail. id: %d, %v", id, templatemsg)
	return err
}

func SendWeChat(user string, TemplateId string, appId, pagePath string, url string, data interface{}) error {
	mini := &template.MiniProgram{appId, pagePath}
	if appId == "" {
		mini = nil
	}
	templatemsg := template.TemplateMessage2{}
	templatemsg.ToUser = user
	templatemsg.TemplateId = TemplateId
	templatemsg.MiniProgram = mini
	templatemsg.Data = data
	templatemsg.URL = url

	id, err := template.Send(WechatClient, templatemsg)
	log.LogWithErrorf(err, "Sendwechat fail. id: %d, %v", id, templatemsg)
	return err
}

func Wxget(code string) (string, string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		AppId, secret, code)
	res, err := client.Get(url)
	if err != nil {
		log.WithError(err).Errorf("获取微信用户信息异常")
		return "", "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Errorf("读取微信用户信息失败")
		return "", "", err
	}

	var data wxBody
	if err = json.Unmarshal(body, &data); err != nil {
		log.WithError(err).Errorf("序列化微信用户信息失败")
		return "", "", err
	}

	if data.ErrCode != 0 {
		log.Errorf("访问微信接口异常：%d -> %s", data.ErrCode, data.ErrMsg)
		return "", "", err
	}

	return data.OpenId, data.SessionKey, nil
}
