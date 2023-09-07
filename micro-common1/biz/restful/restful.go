package restful

import (
	"bytes"
	"common/log"
	"common/util"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type ErrorResultCode uint64

const (
	ResultOk           ErrorResultCode = iota
	ErrInvaidParam     ErrorResultCode = iota + 1000 //无效的参数
	ErrNoFreeRobot                                   //机器人不存在
	ErrNoRobotPosPoint                               //机器人点位不存在
	ErrCacheRedis
	ErrNoJob         //不存在的任务
	ErrRobotNoPoint  //指定的机器人类型没有指定位置的点位
	ErrRobotOffline  //机器人离线了
	ErrRobotCharging //机器人充电中
	ErrRobotBack     //机器人返程中
	ErrRobotBattery  //机器人电量不足
)

type ErrResult struct {
	Code    ErrorResultCode //错误编码
	Message string          //描述信息
	Error   string          //错误信息
}

func Get(url string, responseVo interface{}) error {
	return doRequest(url, http.MethodGet, nil, responseVo)
}

func Post(url string, requestVo, responseVo interface{}) error {
	return doRequest(url, http.MethodPost, requestVo, responseVo)
}

func Put(url string, requestVo, responseVo interface{}) error {
	return doRequest(url, http.MethodPut, requestVo, responseVo)
}

func Delete(url string, requestVo, responseVo interface{}) error {
	return doRequest(url, http.MethodDelete, requestVo, responseVo)
}

func doRequest(url, method string, requestVo, responseVo interface{}) error {
	var buffer io.Reader
	if requestVo != nil {
		data, err := json.Marshal(requestVo)
		if err != nil {
			return util.WrapErr(err, "请求参数序列化失败")
		}
		buffer = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return util.WrapErr(err, "请求创建失败")
	}
	req.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return util.WrapErr(err, "请求发送失败")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("服务不可达，%s", url)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return util.WrapErr(err, "请求结果读取失败")
	}
	log.Debug(string(body))
	if response.StatusCode != http.StatusOK {
		errResult := ErrResult{}
		if err := json.Unmarshal(body, &errResult); err != nil {
			return util.WrapErr(err, "请求失败，错误信息反序列化失败")
		}
		return util.WrapErr(errors.New(errResult.Error), errResult.Message)
	}
	if responseVo == nil {
		return nil
	}
	if err := json.Unmarshal(body, responseVo); err != nil {
		return util.WrapErr(err, "请求成功，结果反序列化失败")
	}
	return nil
}
