package sys_res_mq

import (
	"common/log"
	"common/rabbitmq"
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/suiyunonghen/DxCommonLib"
	"github.com/suiyunonghen/dxsvalue"
	"sort"
	"sync"
	"time"
)

//系统资源变动的RabbitMQ通知

var (
	SystemResChangeExchange = rabbitmq.Exchange{
		Name:  "sysResChange",
		Model: rabbitmq.ET_Fanout,
	}
	procedure *rabbitmq.RbProcedure
)

type SysResType int8

const (
	SysResOffice            SysResType = iota + 1 //机构信息，一般机构
	SysResBuilding                                //机构楼宇信息变动
	SysResRobotType                               //机器人的类型变动修改，以及底盘修改等
	SysResOfficeConfig                            //机构配置信息变动
	SysResFloor                                   //楼层相关信息变动
	SysResRobotPosition                           //机器人位置变动
	SysResRobot                                   //机器人信息变动
	SysResSpecialTimeJob                          //分时段任务
	SysResOfficeJobPriority                       //机构的任务优先级变动
)

type ChangeType int8

const (
	TypeAdd ChangeType = iota + 1
	TypeDel
	TypeModify
)

type SysResChangeNotify struct {
	SysResType SysResType
	ChangeType ChangeType
	OfficeID   string
	ChangeBody interface{} `json:",omitempty"` //通知内容，
}

//楼层资源
type FloorResType int16

const (
	FRTFloor          FloorResType = iota + 1 //楼层改变
	FRTCtrlAreas                              //楼层控制区域
	FRTLiftCtrAreas                           //电梯控制区域变动
	FRTLiftWait                               //电梯等待点变动
	FRTnoCtrlAreaWait                         //非交通管制冲突点变动
	FRTMapChange                              //楼层地图变动
	FRTMapConvertCfg                          //地图转换配置变动
	FRTStandby                                //待机区域变动
)

type SysChangeFloorBody struct {
	ResType FloorResType
	Floor   int16
	BuildId string `json:",omitempty"`
	ResId   string `json:",omitempty"` //变动资源的ID，一般是唯一键（最好是主键）,比如点位是点位guid,区域，是区域ID等,不清楚情况的或者说需要整体变动的可以不填
}

//如果为nil，就是修改的office的配置
type SysChangeConfigBody struct {
	RobotID string `json:",omitempty"` //如果机器人ID不为空，则是修改机构中机器人的配置
	BuildId string `json:",omitempty"` //如果楼宇ID不为空，则是修改机构楼宇配置
}

type SysRobotChassisBody struct {
	RobotType    string `json:",omitempty"` //机器人类型，可为空，为空，就是单纯操作底盘类型
	ChassisModel string //底盘型号不可为空
}

type SpecialTimeJobBody struct {
	RobotId   string
	TimeJobId string
}

type RobotPositionChangeBody struct {
	Floor      int
	BuildingId string
	RobotType  string
	PosGuid    string
}

type SysResChangeNotifyManager struct {
	notifys []SysResChangeNotify
}

func (sysChangeMgr *SysResChangeNotifyManager) Len() int {
	return len(sysChangeMgr.notifys)
}

func (sysChangeMgr *SysResChangeNotifyManager) Less(i, j int) bool {
	return sysChangeMgr.notifys[i].SysResType < sysChangeMgr.notifys[j].SysResType
}

func (sysChangeMgr *SysResChangeNotifyManager) Swap(i, j int) {
	sysChangeMgr.notifys[i], sysChangeMgr.notifys[j] = sysChangeMgr.notifys[j], sysChangeMgr.notifys[i]
}

func (sysChangeMgr *SysResChangeNotifyManager) AddOfficeChangeNotify(OfficeId string, changeType ChangeType) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == OfficeId && sysChangeMgr.notifys[i].SysResType == SysResOffice {
			sysChangeMgr.notifys[i].ChangeType = changeType
			return
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResOffice,
		ChangeType: changeType,
		OfficeID:   OfficeId,
	})
}

//resid标记为变更的资源ID，对应数据库中的ID
func (sysChangeMgr *SysResChangeNotifyManager) AddOfficeJobPriorityChangeNotify(OfficeId, ResId string, changeType ChangeType) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == OfficeId && sysChangeMgr.notifys[i].SysResType == SysResOfficeJobPriority {
			if v, ok := sysChangeMgr.notifys[i].ChangeBody.(string); ok && v == ResId {
				sysChangeMgr.notifys[i].ChangeType = changeType
				return
			}
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResOffice,
		ChangeType: changeType,
		OfficeID:   OfficeId,
		ChangeBody: ResId,
	})
}

func (sysChangeMgr *SysResChangeNotifyManager) AddRobotChangeNotify(officeId, robotId string, changeType ChangeType) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == officeId && sysChangeMgr.notifys[i].SysResType == SysResRobot {
			robotid := ""
			if sysChangeMgr.notifys[i].ChangeBody != nil {
				robotid = sysChangeMgr.notifys[i].ChangeBody.(string)
			}
			if robotid == robotId {
				sysChangeMgr.notifys[i].ChangeType = changeType
				return
			}
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResRobot,
		ChangeType: changeType,
		OfficeID:   officeId,
		ChangeBody: robotId,
	})
}

func (sysChangeMgr *SysResChangeNotifyManager) AddSpecialTimeJobNotify(officeId, robotId, TimeJobId string, changeType ChangeType) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == officeId && sysChangeMgr.notifys[i].SysResType == SysResSpecialTimeJob {
			var jobBody SpecialTimeJobBody
			if sysChangeMgr.notifys[i].ChangeBody != nil {
				jobBody = sysChangeMgr.notifys[i].ChangeBody.(SpecialTimeJobBody)
			}
			if jobBody.RobotId == robotId && jobBody.TimeJobId == TimeJobId {
				sysChangeMgr.notifys[i].ChangeType = changeType
				return
			}
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResSpecialTimeJob,
		ChangeType: changeType,
		OfficeID:   officeId,
		ChangeBody: SpecialTimeJobBody{
			RobotId:   robotId,
			TimeJobId: TimeJobId,
		},
	})
}

func (sysChangeMgr *SysResChangeNotifyManager) AddRefreshRobotPosNotify(officeId string, robotType string) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == officeId && sysChangeMgr.notifys[i].SysResType == SysResRobotPosition {
			var posBody RobotPositionChangeBody
			if sysChangeMgr.notifys[i].ChangeBody != nil {
				posBody = sysChangeMgr.notifys[i].ChangeBody.(RobotPositionChangeBody)
			}
			if posBody.RobotType == robotType && posBody.PosGuid == "" && posBody.BuildingId == "" {
				sysChangeMgr.notifys[i].ChangeType = TypeModify
				return
			}
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResRobotPosition,
		ChangeType: TypeModify,
		OfficeID:   "",
		ChangeBody: RobotPositionChangeBody{
			RobotType: robotType,
		},
	})
}

func (sysChangeMgr *SysResChangeNotifyManager) AddRobotPosChangeNotify(officeId string, poschangeBody RobotPositionChangeBody, changeType ChangeType) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == officeId && sysChangeMgr.notifys[i].SysResType == SysResRobotPosition {
			var posBody RobotPositionChangeBody
			if sysChangeMgr.notifys[i].ChangeBody != nil {
				posBody = sysChangeMgr.notifys[i].ChangeBody.(RobotPositionChangeBody)
			}
			if posBody.RobotType == poschangeBody.RobotType && posBody.PosGuid == poschangeBody.PosGuid && posBody.BuildingId == poschangeBody.BuildingId &&
				posBody.Floor == poschangeBody.Floor {
				sysChangeMgr.notifys[i].ChangeType = changeType
				return
			}
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResRobotPosition,
		ChangeType: changeType,
		OfficeID:   "",
		ChangeBody: poschangeBody,
	})
}

func (sysChangeMgr *SysResChangeNotifyManager) AddBuildChangeNotify(OfficeId, BuildId string, changeType ChangeType) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == OfficeId && sysChangeMgr.notifys[i].SysResType == SysResBuilding {
			buildid := ""
			if sysChangeMgr.notifys[i].ChangeBody != nil {
				buildid = sysChangeMgr.notifys[i].ChangeBody.(string)
			}
			if buildid == BuildId {
				sysChangeMgr.notifys[i].ChangeType = changeType
				return
			}
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResBuilding,
		ChangeType: changeType,
		OfficeID:   OfficeId,
		ChangeBody: BuildId,
	})
}

func (sysChangeMgr *SysResChangeNotifyManager) AddFloorChangeNotify(OfficeId string, changeType ChangeType, changeBody SysChangeFloorBody) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == OfficeId && sysChangeMgr.notifys[i].SysResType == SysResFloor {
			var floorbody SysChangeFloorBody
			if sysChangeMgr.notifys[i].ChangeBody != nil {
				floorbody = sysChangeMgr.notifys[i].ChangeBody.(SysChangeFloorBody)
			}
			if floorbody.BuildId == changeBody.BuildId && floorbody.Floor == changeBody.Floor &&
				floorbody.ResType == changeBody.ResType {
				sysChangeMgr.notifys[i].ChangeType = changeType
				sysChangeMgr.notifys[i].ChangeBody = changeBody
				return
			}
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResFloor,
		ChangeType: changeType,
		OfficeID:   OfficeId,
		ChangeBody: changeBody,
	})
}

func (sysChangeMgr *SysResChangeNotifyManager) AddOfficeCfgChangeNotify(officeId string, changeType ChangeType, changeBody SysChangeConfigBody) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == officeId && sysChangeMgr.notifys[i].SysResType == SysResOfficeConfig {
			var cfgbody SysChangeConfigBody
			if sysChangeMgr.notifys[i].ChangeBody != nil {
				cfgbody = sysChangeMgr.notifys[i].ChangeBody.(SysChangeConfigBody)
			} else if changeBody.BuildId == "" && changeBody.RobotID == "" {
				sysChangeMgr.notifys[i].ChangeType = changeType
				return
			}
			if cfgbody.BuildId == changeBody.BuildId && cfgbody.RobotID == changeBody.RobotID {
				sysChangeMgr.notifys[i].ChangeType = changeType
				sysChangeMgr.notifys[i].ChangeBody = changeBody
				return
			}
		}
	}
	newbody := SysResChangeNotify{
		SysResType: SysResOfficeConfig,
		ChangeType: changeType,
		OfficeID:   officeId,
	}
	if changeBody.BuildId != "" || changeBody.RobotID != "" {
		newbody.ChangeBody = changeBody
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, newbody)
}

func (sysChangeMgr *SysResChangeNotifyManager) AddRobotTypeNotify(changeType ChangeType, RobotType, ChassisModel string) {
	if cap(sysChangeMgr.notifys) == 0 {
		sysChangeMgr.notifys = make([]SysResChangeNotify, 0, 4)
	}
	for i := 0; i < len(sysChangeMgr.notifys); i++ {
		if sysChangeMgr.notifys[i].OfficeID == "" && sysChangeMgr.notifys[i].SysResType == SysResRobotType {
			var body SysRobotChassisBody
			if sysChangeMgr.notifys[i].ChangeBody != nil {
				body = sysChangeMgr.notifys[i].ChangeBody.(SysRobotChassisBody)
			}
			if body.RobotType == RobotType && body.ChassisModel == ChassisModel {
				sysChangeMgr.notifys[i].ChangeType = changeType
				sysChangeMgr.notifys[i].ChangeBody = SysRobotChassisBody{
					RobotType:    RobotType,
					ChassisModel: ChassisModel,
				}
				return
			}
		}
	}
	sysChangeMgr.notifys = append(sysChangeMgr.notifys, SysResChangeNotify{
		SysResType: SysResRobotType,
		ChangeType: changeType,
		OfficeID:   "",
		ChangeBody: SysRobotChassisBody{
			RobotType:    RobotType,
			ChassisModel: ChassisModel,
		},
	})
}

//发送通知到RMQ
func (sysChangeMgr *SysResChangeNotifyManager) NotifySysResChange() error {
	if len(sysChangeMgr.notifys) == 0 {
		return nil
	}
	if procedure == nil {
		proc, err := rabbitmq.DefaultRMQ.RegisterProcedure(SystemResChangeExchange)
		if err != nil {
			return err
		}
		procedure = proc
	}
	sort.Sort(sysChangeMgr)
	result, err := json.Marshal(sysChangeMgr.notifys)
	if err != nil {
		return err
	}
	return procedure.PublishSimple(SystemResChangeExchange.Name, result)
}

type notifyChangeConsume struct {
	consume                 *rabbitmq.RbConsume
	officeChangefunc        func(officeId string, changeType ChangeType) //变动消息
	buildchange             func(officeId, buildId string, changeType ChangeType)
	floorchange             func(officeid string, floorchangeinfo SysChangeFloorBody, changeType ChangeType)
	robotChange             func(officeId, robotId string, changeType ChangeType)
	configChange            func(officeId string, changeBody SysChangeConfigBody, changeType ChangeType)
	robotTypeChange         func(robotType, ChassisModel string, changeType ChangeType)
	specialTimeJobChange    func(officeId, robotId, timeJobId string, changeType ChangeType)
	robotPosChange          func(officeId string, changebody RobotPositionChangeBody, changeType ChangeType)
	officeJobPriorityChange func(officeId, resId string, changeType ChangeType) //机构任务优先级变动
	sync.RWMutex
}

func (changeConsume *notifyChangeConsume) createConsume(lock bool) error {
	if lock {
		changeConsume.Lock()
		defer changeConsume.Unlock()
	}
	if changeConsume.consume == nil {
		consume, err := rabbitmq.DefaultRMQ.RegisterConsume(SystemResChangeExchange, "", true, changeConsume.subNotify)
		if err != nil {
			return err
		}
		err = consume.Subscribe(SystemResChangeExchange.Name)
		if err == nil {
			changeConsume.consume = consume
		}
		return err
	}
	return nil
}

func (changeConsume *notifyChangeConsume) subNotify(delivery amqp.Delivery) {
	notifys := dxsvalue.NewCacheValue(dxsvalue.VT_Array)
	err := notifys.LoadFromJson(delivery.Body, true)
	if err != nil {
		log.WithError(err).Error("解析系统资源变更通知错误")
		return
	}
	//稍微等一等，防止同步未成功
	DxCommonLib.Sleep(time.Second)

	for i := 0; i < notifys.Count(); i++ {
		notifyValue := notifys.ValueByIndex(i)
		tp := notifyValue.AsInt("SysResType", 0)
		if tp == 0 {
			continue
		}
		restype := SysResType(tp)
		tp = notifyValue.AsInt("ChangeType", 0)
		if tp == 0 {
			continue
		}
		officeId := notifyValue.AsString("OfficeID", "")
		if officeId == "" && restype != SysResRobotType {
			continue
		}
		switch restype {
		case SysResOffice:
			changeConsume.RLock()
			officechange := changeConsume.officeChangefunc
			changeConsume.RUnlock()
			if officechange != nil {
				officechange(officeId, ChangeType(tp))
			}
		case SysResOfficeJobPriority:
			changeConsume.RLock()
			officeJobPriorityChange := changeConsume.officeJobPriorityChange
			changeConsume.RUnlock()

			if officeJobPriorityChange != nil {
				bodyValue := notifyValue.ValueByName("ChangeBody")
				resId := ""
				if bodyValue != nil {
					resId = bodyValue.String()
				}
				officeJobPriorityChange(officeId, resId, ChangeType(tp))
			}
		case SysResBuilding:
			buildid := notifyValue.AsString("ChangeBody", "")
			if buildid == "" {
				continue
			}
			changeConsume.RLock()
			buildchange := changeConsume.buildchange
			changeConsume.RUnlock()
			if buildchange != nil {
				buildchange(officeId, buildid, ChangeType(tp))
			}
		case SysResRobot:
			//为空的时候，重新拉取所有机器人信息
			robotId := notifyValue.AsString("ChangeBody", "")
			changeConsume.RLock()
			robotChange := changeConsume.robotChange
			changeConsume.RUnlock()
			if robotChange != nil {
				robotChange(officeId, robotId, ChangeType(tp))
			}
		case SysResFloor:
			bodyValue := notifyValue.ValueByName("ChangeBody")
			buildid := bodyValue.AsString("BuildId", "")
			Floor := bodyValue.AsInt("Floor", -1000)
			floorResType := bodyValue.AsInt("FloorResType", 0)
			if buildid != "" && Floor != -1000 && floorResType != 0 {
				changeConsume.RLock()
				floorchange := changeConsume.floorchange
				changeConsume.RUnlock()
				if floorchange != nil {
					floorchange(officeId, SysChangeFloorBody{
						BuildId: buildid,
						Floor:   int16(Floor),
						ResType: FloorResType(floorResType),
						ResId:   bodyValue.AsString("ResId", ""),
					}, ChangeType(tp))
				}
			}
		case SysResOfficeConfig:
			changeConsume.RLock()
			configChange := changeConsume.configChange
			changeConsume.RUnlock()
			if configChange != nil {
				bodyValue := notifyValue.ValueByName("ChangeBody")
				configChange(officeId, SysChangeConfigBody{
					RobotID: bodyValue.AsString("RobotID", ""),
					BuildId: bodyValue.AsString("BuildId", ""),
				}, ChangeType(tp))
			}
		case SysResRobotType:
			bodyValue := notifyValue.ValueByName("ChangeBody")
			ChassisModel := bodyValue.AsString("ChassisModel", "")
			if ChassisModel == "" {
				continue
			}
			changeConsume.RLock()
			robotTypeChange := changeConsume.robotTypeChange
			changeConsume.RUnlock()
			if robotTypeChange != nil {
				robotTypeChange(bodyValue.AsString("RobotType", ""), ChassisModel, ChangeType(tp))
			}
		case SysResSpecialTimeJob:
			bodyValue := notifyValue.ValueByName("ChangeBody")
			if bodyValue == nil {
				return
			}
			robotid := bodyValue.AsString("RobotId", "")
			timeJobId := bodyValue.AsString("TimeJobId", "")
			if robotid == "" || timeJobId == "" {
				return
			}
			changeConsume.RLock()
			specialTimeJobChange := changeConsume.specialTimeJobChange
			changeConsume.RUnlock()
			if specialTimeJobChange != nil {
				specialTimeJobChange(officeId, robotid, timeJobId, ChangeType(tp))
			}
		case SysResRobotPosition:
			bodyValue := notifyValue.ValueByName("ChangeBody")
			if bodyValue == nil {
				return
			}
			RobotType := bodyValue.AsString("RobotType", "")
			PosGuid := bodyValue.AsString("PosGuid", "")
			Floor := bodyValue.AsInt("Floor", -1000)
			buildid := bodyValue.AsString("BuildingId", "")
			if RobotType == "" || Floor == -1000 || buildid == "" {
				return
			}
			changeConsume.RLock()
			robotPosChange := changeConsume.robotPosChange
			changeConsume.RUnlock()
			if robotPosChange != nil {
				robotPosChange(officeId, RobotPositionChangeBody{
					Floor:      Floor,
					BuildingId: buildid,
					PosGuid:    PosGuid,
					RobotType:  RobotType,
				}, ChangeType(tp))
			}
		}
	}
	notifys.Clear()
	dxsvalue.FreeValue(notifys)

}

var (
	changeConsume notifyChangeConsume
)

func SubOfficeChange(officeChange func(officeId string, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.officeChangefunc = officeChange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}

func SubOfficeJobPriorityChange(officeJobPriorityChange func(officeId, resId string, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.officeJobPriorityChange = officeJobPriorityChange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}

func SubBuildingChange(buildchange func(officeId, BuildId string, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.buildchange = buildchange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}

func SubRobotChange(robotChange func(officeId, robotid string, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.robotChange = robotChange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}

func SubSpecialTimeJobChange(specialTimeJobChange func(officeId, robotid, timeJobId string, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.specialTimeJobChange = specialTimeJobChange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}

func SubFloorChange(floorchange func(officeid string, floorchangeinfo SysChangeFloorBody, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.floorchange = floorchange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}

func SubOfficeCfgChange(configChange func(officeId string, changeBody SysChangeConfigBody, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.configChange = configChange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}

func SubRobotTypeChange(robotTypeChange func(robotType, ChassisModel string, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.robotTypeChange = robotTypeChange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}

func SubRobotPosChange(robotPosChange func(officeId string, changebody RobotPositionChangeBody, changeType ChangeType)) error {
	changeConsume.Lock()
	changeConsume.robotPosChange = robotPosChange
	err := changeConsume.createConsume(false)
	changeConsume.Unlock()
	return err
}
