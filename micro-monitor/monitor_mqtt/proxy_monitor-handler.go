package monitor_mqtt

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/utils"
	"fmt"
	"micro-common1/log"
	"micro-common1/redis"
	"strings"
	"time"
)

func proxyMonitorHandler(topic string, msg *MqttMsgVo) {

	topics := strings.Split(topic, "/")

	_, err := utils.ParseToken(msg.Token)
	if err != nil {
		log.WithError(err).Errorf("topic:%s, token:%s校验失败", topic, msg.Token)
		publishInvalidToken(topics, msg.MsgID)
		return
	}

	server := ProxyServer{
		OfficeId:   topics[3],
		ProxyId:    topics[4],
		UploadTime: time.Now(),
	}

	ctx := context.Background()
	jsonData, _ := json.Marshal(server)
	redis.HSet(ctx, constant.ProxyStatus, fmt.Sprintf("%s:%s", server.OfficeId, server.ProxyId), jsonData)
}
