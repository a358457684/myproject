package monitor_websocket

import (
	"context"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/model"
	"epshealth-airobot-monitor/result"
	"epshealth-airobot-monitor/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"micro-common1/log"
	"micro-common1/redis"
	wsManager "micro-common1/websocket"
	"net/http"
)

var _manager *wsManager.WsManager

func init() {
	_manager = wsManager.NewWsManager()
}

func WsPage(c *gin.Context) {
	token := c.Request.Header.Get("Sec-WebSocket-Protocol")
	log.Infof("ws connect token info %s", token)
	_, err := utils.ParseToken(token)
	if err != nil || token == "" {
		log.WithError(err).Errorf("ws connect token error")
		result.Custom(c, result.InvalidToken, "访问未授权")
		return
	}
	upgrade := &websocket.Upgrader{
		// cross origin domain 这个是解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Subprotocols: []string{token},
	}
	conn, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	// parseToken 解析token包含的信息
	_, err = utils.ParseToken(token)
	if err != nil || token == "" {
		log.WithError(err).Errorf("ws connect token error")
		result.Custom(c, result.InvalidToken, "访问未授权")
		_ = conn.WriteMessage(websocket.TextMessage, []byte("999"))
		_ = conn.Close()
		return
	}
	var client RobotWebSocket
	client.Conn = conn
	_manager.RegisterClient(&client)
}

// 机器人推送列表发生了变化，推送到web端
func SendRobotPush(isUpdate bool, data model.ElasticRobotPushMessage) {
	sendRobotPush(isUpdate, data)
}

func SubscribeWebSocketQueue() {
	pubSub := redis.Client.Subscribe(context.Background(), constant.WebsocketQueues...)
	defer func() {
		_ = pubSub.Close()
	}()
	for msg := range pubSub.Channel() {
		sendMsg(msg)
	}
}
