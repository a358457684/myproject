package monitor_websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"micro-common1/biz/manager"
	"micro-common1/log"
	"micro-common1/util"
	"net"
)

type RobotWebSocket struct {
	*websocket.Conn
	snowID     int64
	Path       string            `json:"path"`
	OfficeId   string            `json:"officeId"`
	RobotId    string            `json:"robotId"`
	RobotModel manager.RobotType `json:"robotModel"`
	BuildingId string            `json:"buildingId"`
	Floor      int               `json:"floor"`
}

// 唯一ID
func (socket *RobotWebSocket) ClientID() int64 {
	if socket.snowID == 0 {
		socket.snowID = util.GetSnowflakeID()
	}
	return socket.snowID
}

func (socket *RobotWebSocket) OnError(err error) {
	// 错误
	log.WithError(err).Errorf("%d: ==== ws error ====", socket.snowID)
}

func (socket *RobotWebSocket) OnDisConnect() {
	// 连接断开
	log.Infof("%d: ==== ws disconnect ====", socket.snowID)
}

func (socket *RobotWebSocket) OnRecvMessage(messageType int, data []byte) error {
	if messageType == websocket.TextMessage && data != nil {
		// 处理接收的内容
		var socketData RobotWebSocket
		err := json.Unmarshal(data, &socketData)
		if err != nil || socketData.Path == "" {
			log.WithError(err).Errorf("websocket的参数错误：%s", string(data))
			return net.UnknownNetworkError("参数错误")
		}
		// 客户端的心跳
		if socketData.Path == ping {
			return socket.WriteMessage(websocket.TextMessage, []byte(pongMessage))
		}
		if socketData.OfficeId == "" {
			log.Errorf("websocket的机构为空：%s", string(data))
			return net.UnknownNetworkError("机构为空")
		}
		socket.Path = socketData.Path
		socket.OfficeId = socketData.OfficeId
		socket.RobotId = socketData.RobotId
		socket.RobotModel = socketData.RobotModel
		socket.BuildingId = socketData.BuildingId
		socket.Floor = socketData.Floor
	}
	return nil
}
