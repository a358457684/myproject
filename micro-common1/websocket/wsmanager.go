package websocket

import (
	"common/log"
	"common/util"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"runtime"
	"time"
)

type WsClienter interface {
	//WebSocket的接口
	io.Closer
	// LocalAddr returns the local network address.
	LocalAddr() net.Addr
	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr
	Subprotocol() string
	WriteControl(messageType int, data []byte, deadline time.Time) error
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	WriteJSON(v interface{}) error
	ReadJSON(v interface{}) error

	//逻辑接口
	ClientID() int64 //唯一ID
	OnError(err error)
	OnDisConnect()
	OnRecvMessage(messageType int, data []byte) error
}

type clientInfo struct {
	client WsClienter
	quit   chan struct{}
}

type BroadCastFunction func(client WsClienter, params ...interface{}) bool

type broadcastIntf struct {
	value        interface{}
	canBroadCast BroadCastFunction //根据参数判定是否允许广播
	checkParams  []interface{}
}

type safeSendIntf struct {
	sendIds []int64
	value   interface{}
}

type WsManager struct {
	broadcast   chan broadcastIntf
	clientChan  chan clientInfo
	safeSend    chan safeSendIntf
	errClientID chan int64
}

func (manager *WsManager) run() {
	clients := make(map[int64]clientInfo, 10240)
	delCount := 0
	doResetMap := func() {
		if delCount > 819200 {
			//变动太大，重建map
			newMap := make(map[int64]clientInfo, 10240)
			for k, v := range clients {
				newMap[k] = v
			}
			clients = newMap
			delCount = 0
		}
	}
	for {
		select {
		case cinfo := <-manager.clientChan:
			cid := cinfo.client.ClientID()
			oldClientInfo, ok := clients[cid]
			if !ok {
				if cinfo.quit != nil {
					clients[cid] = cinfo
				}
			} else if cinfo.quit == nil {
				//删除
				close(oldClientInfo.quit)
				delete(clients, cid)
				delCount++
				doResetMap()
			}
		case safeValue := <-manager.safeSend:
			switch value := safeValue.value.(type) {
			case []byte:
				for i := 0; i < len(safeValue.sendIds); i++ {
					client, ok := clients[safeValue.sendIds[i]]
					if !ok {
						continue
					}
					err := client.client.WriteMessage(websocket.BinaryMessage, value)
					if err != nil {
						close(client.quit)
						client.client.OnError(err)
						delete(clients, safeValue.sendIds[i])
						log.WithError(err).Error("写入数据发生错误,关闭连接")
						delCount++
					}
				}
			case BroadCastCmd:
				if value == BroadCastRemove {
					for _, clientId := range safeValue.sendIds {
						if client, ok := clients[clientId]; ok {
							close(client.quit)
							delete(clients, clientId)
							delCount++
						}
					}
				}
			default:
				buffer := util.GetBuffer()
				err := json.NewEncoder(buffer).Encode(value)
				if err != nil {
					util.FreeBuffer(buffer)
					log.WithError(err).Error("Broadcast序列化失败")
					continue
				}
				for i := 0; i < len(safeValue.sendIds); i++ {
					client, ok := clients[safeValue.sendIds[i]]
					if !ok {
						continue
					}
					err = client.client.WriteMessage(websocket.TextMessage, buffer.Bytes())
					if err != nil {
						close(client.quit)
						client.client.OnError(err)
						delete(clients, safeValue.sendIds[i])
						log.WithError(err).Error("写入数据发生错误,关闭连接")
						delCount++
					}
				}
				util.FreeBuffer(buffer)
			}
		case errCid := <-manager.errClientID:
			for cid, client := range clients {
				if errCid == cid {
					//移除这个
					close(client.quit)
					delete(clients, cid)
					delCount++
					break
				}
			}
		case bvalue := <-manager.broadcast:
			if bvalue.value == nil {
				//是要自己处理发送内容的
				for _, client := range clients {
					if bvalue.canBroadCast != nil {
						bvalue.canBroadCast(client.client, bvalue.checkParams...) //在这里自定义发送
					}
				}
				continue
			}
			switch value := bvalue.value.(type) {
			case []byte:
				for cid, client := range clients {
					if bvalue.canBroadCast != nil && !bvalue.canBroadCast(client.client, bvalue.checkParams...) {
						continue
					}
					err := client.client.WriteMessage(websocket.BinaryMessage, value)
					if err != nil {
						//移除这个
						close(client.quit)
						client.client.OnError(err)
						delete(clients, cid)
						log.WithError(err).Error("写入数据发生错误,关闭连接")
						delCount++
					}
				}
			case BroadCastCmd:
				for cid, client := range clients {
					if bvalue.canBroadCast != nil {
						if bvalue.canBroadCast(client.client, bvalue.checkParams...) {
							switch value {
							case BroadCastCheckTimeout:
								//超时了，移除
								close(client.quit)
								delete(clients, cid)
								delCount++
							}
						}
					}
				}
			default:
				buffer := util.GetBuffer()
				err := json.NewEncoder(buffer).Encode(value)
				if err != nil {
					log.WithError(err).Error("Broadcast序列化失败")
				} else {
					for cid, client := range clients {
						if bvalue.canBroadCast != nil && !bvalue.canBroadCast(client.client, bvalue.checkParams...) {
							continue
						}
						err = client.client.WriteMessage(websocket.TextMessage, buffer.Bytes())
						if err != nil {
							close(client.quit)
							client.client.OnError(err)
							delete(clients, cid)
							log.WithError(err).Error("写入数据发生错误,关闭连接")
							delCount++
						}
					}
				}
				util.FreeBuffer(buffer)
			}
			doResetMap()
		}
	}
}

func (manager *WsManager) processSocket(clientCon WsClienter, quitchan <-chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			const size = 65535
			buf := make([]byte, size)
			ss := runtime.Stack(buf, false)
			if ss > size {
				ss = size
			}
			buf = buf[:ss]
			log.Errorf("WebSocket业务逻辑panic：%s", buf)
		}
		clientCon.Close()
		clientCon.OnDisConnect()
	}()
	for {
		select {
		case <-quitchan:
			return
		default:
			messageType, data, err := clientCon.ReadMessage()
			if len(data) > 0 {
				err = clientCon.OnRecvMessage(messageType, data)
				if _, ok := err.(net.Error); !ok && err != io.EOF {
					err = nil
				}
			}
			if err != nil {
				clientCon.OnError(err)
				manager.errClientID <- clientCon.ClientID()
				return
			}
		}
	}
}

func (manager *WsManager) RegisterClient(clientCon WsClienter) {
	quit := make(chan struct{})
	manager.clientChan <- clientInfo{
		client: clientCon,
		quit:   quit,
	}
	manager.processSocket(clientCon, quit)
}

func (manager *WsManager) UnRegisterClient(clientCon WsClienter) {
	manager.clientChan <- clientInfo{
		client: clientCon,
		quit:   nil,
	}
}

func (manager *WsManager) Broadcast(value interface{}, canBroadCast BroadCastFunction, checkParams ...interface{}) {
	manager.broadcast <- broadcastIntf{
		value:        value,
		canBroadCast: canBroadCast,
		checkParams:  checkParams,
	}
}

func (manager *WsManager) DirectSend(value interface{}, Clienters ...WsClienter) {
	if value == nil || len(Clienters) == 0 {
		return
	}
	buffer := util.GetBuffer()
	err := json.NewEncoder(buffer).Encode(value)
	if err != nil {
		log.WithError(err).Error("Broadcast序列化失败")
	} else {
		for i := 0; i < len(Clienters); i++ {
			err = Clienters[i].WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.WithError(err).Error("写入数据发生错误,关闭连接")
			}
		}
	}
	util.FreeBuffer(buffer)
}

func (manager *WsManager) SafeSend(value interface{}, Clienters ...WsClienter) {
	if value == nil || len(Clienters) == 0 {
		return
	}
	ids := make([]int64, len(Clienters))
	for i := 0; i < len(Clienters); i++ {
		ids[i] = Clienters[i].ClientID()
	}
	manager.safeSend <- safeSendIntf{
		value:   value,
		sendIds: ids,
	}
}

func createcanBroadCast(cmd BroadCastCmd, canBroadCast func(cmd BroadCastCmd, client WsClienter, params ...interface{}) bool) BroadCastFunction {
	return func(client WsClienter, params ...interface{}) bool {
		return canBroadCast(cmd, client, params...)
	}
}

type BroadCastCmd int8

const (
	BroadCastCheckTimeout BroadCastCmd = iota + 1
	BroadCastRemove                    //移除
)

//广播检查是否超时的
func (manager *WsManager) BroadCastCheckTimeout(broadcastFunc func(cmd BroadCastCmd, client WsClienter, params ...interface{}) bool, checkParams ...interface{}) {
	manager.broadcast <- broadcastIntf{
		value:        BroadCastCheckTimeout,
		canBroadCast: createcanBroadCast(BroadCastCheckTimeout, broadcastFunc),
		checkParams:  checkParams,
	}
}

//广播指令操作
func (manager *WsManager) BroadCastCommand(cmd BroadCastCmd, broadcastFunc func(cmd BroadCastCmd, client WsClienter, params ...interface{}) bool, checkParams ...interface{}) {
	manager.broadcast <- broadcastIntf{
		value:        cmd,
		canBroadCast: createcanBroadCast(cmd, broadcastFunc),
		checkParams:  checkParams,
	}
}

func NewWsManager() *WsManager {
	result := &WsManager{
		broadcast:   make(chan broadcastIntf, 32),
		clientChan:  make(chan clientInfo, 32),
		errClientID: make(chan int64, 32),
		safeSend:    make(chan safeSendIntf, 32),
	}
	go result.run()
	return result
}
