package controller

import (
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/dao"
	monitorMqtt "epshealth-airobot-monitor/monitor_mqtt"
	"epshealth-airobot-monitor/result"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"micro-common1/biz/enum"
	"micro-common1/redis"
	"time"
)

type ProxyVo struct {
	Id            string             `json:"id"`
	Ip            string             `json:"ip"`
	Account       string             `json:"account"`
	NetStatus     enum.NetStatusEnum `json:"netStatus"`
	NetStatusText string             `json:"netStatusText"`
	UploadTime    time.Time          `json:"uploadTime"`
}

// @Tags proxy
// @Summary 获取代理服务信息
// @Description 获取代理服务信息
// @Security ApiKeyAuth
// @Param officeId query string true "机构信息"
// @Success 200 {object} result.Result{data=[]ProxyVo}
// @Router /proxyServer/list [get]
func FindProxyServer(c *gin.Context) {

	officeId := c.Query("officeId")
	if officeId == "" {
		result.BadRequest(c, errors.New("机构信息不能为空"))
		return
	}

	dataList := dao.FindFrontServerByOfficeId(officeId)
	var proxies []ProxyVo
	for _, data := range dataList {
		vo := ProxyVo{
			Id:            data.Id,
			Ip:            data.Ip,
			Account:       data.Account,
			NetStatus:     enum.NsOffline,
			NetStatusText: enum.NsOffline.Description(),
		}
		res, err := redis.HGet(c, constant.ProxyStatus, fmt.Sprintf("%s:%s", officeId, data.Id)).Result()
		if err != nil {
			continue
		}
		var proxyServer monitorMqtt.ProxyServer
		_ = json.Unmarshal([]byte(res), &proxyServer)
		vo.UploadTime = proxyServer.UploadTime
		if proxyServer.UploadTime.After(time.Now().Add(-30 * time.Second)) {
			vo.NetStatus = enum.NsOnline
			vo.NetStatusText = enum.NsOnline.Description()
		}
		proxies = append(proxies, vo)
	}
	result.Success(c, proxies)
}
