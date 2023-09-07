package mq

import (
	"common/log"
	"common/rabbitmq"
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
	"strings"
	"time"
)

//机器人对资源的锁定和释放的通知信息

var (
	ResourceStatusExchange = rabbitmq.Exchange{
		Name:  "resUseChange",
		Model: rabbitmq.ET_Fanout,
	}
	MonitorRouter = "dispatch_srv/monitor/"
)

//资源类型
type ResourceType uint8

const (
	ResTypeCtrlArea     ResourceType = iota + 1 //交通管制区域资源
	ResTypeLiftCtrlArea                         //电梯管制区资源  等待点在这里一起返回
	ResTypeLift                                 //电梯类型资源
	ResTypeSafeWait                             //安全距离的停靠点类型
	ResTypeFloor                                //楼层资源，用来判定管理楼层最大的机器人数量
)

//资源使用者的信息
type ResourceUseRobotInfo struct {
	RobotID            string
	ReleaseDescription string    `json:",omitempty"` //释放的描述，一般异常情况，会有这个描述
	LockStart          time.Time `json:",omitempty"`
	ReleaseTime        time.Time `json:",omitempty"`
}

type WaitPointResourceInfo struct {
	ResourceUseRobotInfo
	WaitPointID string
}

//管制区的通知,
type CtrlAreaResNotify struct {
	MaxRobotCount        int                     //资源区能够容纳的最大机器人数量
	CtrlAreaID           string                  `json:",omitempty"` //资源管制区ID
	CurrentRobotIDs      []ResourceUseRobotInfo  `json:",omitempty"` //当前区域的机器人信息
	ReleaseRobotIDs      []ResourceUseRobotInfo  `json:",omitempty"` //当前释放了资源的机器人信息
	CurrentUseWaitPoints []WaitPointResourceInfo `json:",omitempty"` //当前正在使用的等待点的信息
	ReleaseWaitPoints    []WaitPointResourceInfo `json:",omitempty"` //当前通知释放的点
}

func (ctrareaResNotify *CtrlAreaResNotify) Empty() bool {
	return ctrareaResNotify.CtrlAreaID == "" || len(ctrareaResNotify.CurrentRobotIDs) == 0 && len(ctrareaResNotify.ReleaseRobotIDs) == 0 &&
		len(ctrareaResNotify.CurrentUseWaitPoints) == 0 && len(ctrareaResNotify.ReleaseWaitPoints) == 0
}

//电梯资源的通知信息
type LiftResourceNotify struct {
	LiftID               string                  `json:",omitempty"` //电梯ID
	CurrentRobot         *ResourceUseRobotInfo   `json:",omitempty"` //当前的机器人使用的信息
	ReleaseRobot         *ResourceUseRobotInfo   `json:",omitempty"` //释放电梯的机器人信息
	CurrentUseWaitPoints []WaitPointResourceInfo `json:",omitempty"` //当前正在使用的等待点的信息
	ReleaseWaitPoints    []WaitPointResourceInfo `json:",omitempty"` //当前通知释放的点
}

func (liftResNotify *LiftResourceNotify) Empty() bool {
	return liftResNotify.LiftID == "" || liftResNotify.CurrentRobot == nil && liftResNotify.ReleaseRobot == nil &&
		len(liftResNotify.CurrentUseWaitPoints) == 0 && len(liftResNotify.ReleaseWaitPoints) == 0
}

//安全资源的通知信息
type SafeResourceNotify struct {
	CurrentUseWaitPoints []WaitPointResourceInfo `json:",omitempty"` //当前正在使用的等待点的信息
	ReleaseWaitPoints    []WaitPointResourceInfo `json:",omitempty"` //当前通知释放的点
}

func (safeResNotify *SafeResourceNotify) Empty() bool {
	return len(safeResNotify.CurrentUseWaitPoints) == 0 && len(safeResNotify.ReleaseWaitPoints) == 0
}

//楼层资源通知
type FloorResourceNotify struct {
	CurrentRobots []ResourceUseRobotInfo `json:",omitempty"` //当前正在使用楼层的机器人
	ReleaseRobots []ResourceUseRobotInfo `json:",omitempty"` //当前释放机器人的楼层
}

func (floorResNotify *FloorResourceNotify) Empty() bool {
	return len(floorResNotify.CurrentRobots) == 0 && len(floorResNotify.ReleaseRobots) == 0
}

type ResourceNotify struct {
	ResType        ResourceType         //资源类型
	Floor          int32                //楼层
	BuildID        string               `json:",omitempty"` //楼宇
	AreaResNotify  *CtrlAreaResNotify   `json:",omitempty"` //管制区域的通知（包括交通管制和电梯管制区）
	LiftResNotify  *LiftResourceNotify  `json:",omitempty"` //电梯区域信息
	SafeResNotify  *SafeResourceNotify  `json:",omitempty"` //安全资源的通知信息
	FloorResNotify *FloorResourceNotify `json:",omitempty"` //楼层资源
	childNotifys   []*ResourceNotify    `json:"-"`          //一个主通知下面有多个子通知，订阅的时候，不用关注，主要是推送通知的时候会顺序一起推送
}

func (resNotify *ResourceNotify) clear() {
	for i := 0; i < len(resNotify.childNotifys); i++ {
		resNotify.childNotifys[i] = nil
	}
	resNotify.childNotifys = resNotify.childNotifys[:0]
	resNotify.AreaResNotify = nil
	resNotify.LiftResNotify = nil
	resNotify.SafeResNotify = nil
	resNotify.FloorResNotify = nil
}

func (resNotify *ResourceNotify) Reset(floor int32, buildid string, resType ResourceType) {
	resNotify.Floor = floor
	resNotify.BuildID = buildid
	if resNotify.ResType == resType {
		for i := 0; i < len(resNotify.childNotifys); i++ {
			resNotify.childNotifys[i] = nil
		}
		resNotify.childNotifys = resNotify.childNotifys[:0]

		switch resNotify.ResType {
		case ResTypeFloor:
			if resNotify.FloorResNotify != nil {
				resNotify.FloorResNotify.ReleaseRobots = resNotify.FloorResNotify.ReleaseRobots[:0]
				resNotify.FloorResNotify.CurrentRobots = resNotify.FloorResNotify.CurrentRobots[:0]
			}
		case ResTypeSafeWait:
			if resNotify.SafeResNotify != nil {
				resNotify.SafeResNotify.CurrentUseWaitPoints = resNotify.SafeResNotify.CurrentUseWaitPoints[:0]
				resNotify.SafeResNotify.ReleaseWaitPoints = resNotify.SafeResNotify.ReleaseWaitPoints[:0]
			}
		case ResTypeLiftCtrlArea, ResTypeCtrlArea:
			if resNotify.AreaResNotify != nil {
				resNotify.AreaResNotify.ReleaseWaitPoints = resNotify.AreaResNotify.ReleaseWaitPoints[:0]
				resNotify.AreaResNotify.CurrentUseWaitPoints = resNotify.AreaResNotify.CurrentUseWaitPoints[:0]
				resNotify.AreaResNotify.ReleaseRobotIDs = resNotify.AreaResNotify.ReleaseRobotIDs[:0]
				resNotify.AreaResNotify.CurrentRobotIDs = resNotify.AreaResNotify.CurrentRobotIDs[:0]
			}
		case ResTypeLift:
			if resNotify.LiftResNotify != nil {
				resNotify.LiftResNotify.CurrentUseWaitPoints = resNotify.LiftResNotify.CurrentUseWaitPoints[:0]
				resNotify.LiftResNotify.ReleaseWaitPoints = resNotify.LiftResNotify.ReleaseWaitPoints[:0]
				resNotify.LiftResNotify.ReleaseRobot = nil
				resNotify.LiftResNotify.CurrentRobot = nil
			}
		}
		return
	}
	resNotify.clear()
	resNotify.ResType = resType
	switch resType {
	case ResTypeCtrlArea, ResTypeLiftCtrlArea:
		resNotify.AreaResNotify = &CtrlAreaResNotify{
			CurrentRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			ReleaseRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 3),
			ReleaseWaitPoints:    make([]WaitPointResourceInfo, 0, 3),
		}
	case ResTypeSafeWait:
		resNotify.SafeResNotify = &SafeResourceNotify{
			CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 4),
			ReleaseWaitPoints:    make([]WaitPointResourceInfo, 0, 4),
		}
	case ResTypeLift:
		resNotify.AreaResNotify = &CtrlAreaResNotify{
			CurrentRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			ReleaseRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 3),
			ReleaseWaitPoints:    make([]WaitPointResourceInfo, 0, 3),
		}
	case ResTypeFloor:
		resNotify.FloorResNotify = &FloorResourceNotify{
			CurrentRobots: make([]ResourceUseRobotInfo, 0, 8),
			ReleaseRobots: make([]ResourceUseRobotInfo, 0, 8),
		}
	}
}

func NewLiftResourceNotify(floor int32, buildId, LiftId string) *ResourceNotify {
	result := &ResourceNotify{
		ResType: ResTypeLift,
		Floor:   floor,
		BuildID: buildId,
		LiftResNotify: &LiftResourceNotify{
			LiftID: LiftId,
		},
	}
	return result
}

func NewResourceNotify(ResType ResourceType, floor int32, buildId string) *ResourceNotify {
	result := &ResourceNotify{
		ResType: ResType,
		Floor:   floor,
		BuildID: buildId,
	}
	switch ResType {
	case ResTypeCtrlArea, ResTypeLiftCtrlArea:
		result.AreaResNotify = &CtrlAreaResNotify{
			CurrentRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			ReleaseRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 3),
			ReleaseWaitPoints:    make([]WaitPointResourceInfo, 0, 3),
		}
	case ResTypeSafeWait:
		result.SafeResNotify = &SafeResourceNotify{
			CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 4),
			ReleaseWaitPoints:    make([]WaitPointResourceInfo, 0, 4),
		}
	case ResTypeLift:
		result.AreaResNotify = &CtrlAreaResNotify{
			CurrentRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			ReleaseRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 3),
			ReleaseWaitPoints:    make([]WaitPointResourceInfo, 0, 3),
		}
	case ResTypeFloor:
		result.FloorResNotify = &FloorResourceNotify{
			CurrentRobots: make([]ResourceUseRobotInfo, 0, 8),
			ReleaseRobots: make([]ResourceUseRobotInfo, 0, 8),
		}
	}
	return result
}

func NewCtrlAreaResourceNotify(ResType ResourceType, floor int32, buildId, AreaId string, maxRobotCount int) *ResourceNotify {
	result := &ResourceNotify{
		ResType: ResType,
		Floor:   floor,
		BuildID: buildId,
	}
	switch ResType {
	case ResTypeCtrlArea, ResTypeLiftCtrlArea:
		result.AreaResNotify = &CtrlAreaResNotify{
			CtrlAreaID:           AreaId,
			MaxRobotCount:        maxRobotCount,
			CurrentRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			ReleaseRobotIDs:      make([]ResourceUseRobotInfo, 0, 3),
			CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 3),
			ReleaseWaitPoints:    make([]WaitPointResourceInfo, 0, 3),
		}
	}
	return result
}

func NewSafeResourceNotify(floor int32, buildId string) *ResourceNotify {
	result := &ResourceNotify{
		ResType: ResTypeSafeWait,
		Floor:   floor,
		BuildID: buildId,
		SafeResNotify: &SafeResourceNotify{
			CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 4),
			ReleaseWaitPoints:    make([]WaitPointResourceInfo, 0, 4),
		},
	}
	return result
}

func (resnotify *ResourceNotify) RobotLockFloor(robotid, buildId string, floor int) {
	var child *ResourceNotify
	if resnotify.ResType != ResTypeFloor {
		child = NewResourceNotify(ResTypeFloor, int32(floor), buildId)
		resnotify.childNotifys = append(resnotify.childNotifys, child)
	} else {
		if resnotify.Floor == int32(floor) && resnotify.BuildID == buildId {
			child = resnotify
		} else {
			child = NewResourceNotify(ResTypeFloor, int32(floor), buildId)
			resnotify.childNotifys = append(resnotify.childNotifys, child)
		}
	}
	for i := 0; i < len(child.FloorResNotify.CurrentRobots); i++ {
		if child.FloorResNotify.CurrentRobots[i].RobotID == robotid {
			child.FloorResNotify.CurrentRobots[i].LockStart = time.Now()
			return
		}
	}
}

func (resnotify *ResourceNotify) RobotReleaseFloor(robotid, buildId, resDes string, floor int, lockStart time.Time) {
	var child *ResourceNotify
	if resnotify.ResType != ResTypeFloor {
		child = NewResourceNotify(ResTypeFloor, int32(floor), buildId)
		resnotify.childNotifys = append(resnotify.childNotifys, child)
	} else {
		if resnotify.Floor == int32(floor) && resnotify.BuildID == buildId {
			child = resnotify
		} else {
			child = NewResourceNotify(ResTypeFloor, int32(floor), buildId)
			resnotify.childNotifys = append(resnotify.childNotifys, child)
		}
	}
	for i := 0; i < len(child.FloorResNotify.ReleaseRobots); i++ {
		if child.FloorResNotify.ReleaseRobots[i].RobotID == robotid {
			child.FloorResNotify.ReleaseRobots[i].LockStart = lockStart
			child.FloorResNotify.ReleaseRobots[i].ReleaseTime = time.Now()
			child.FloorResNotify.ReleaseRobots[i].ReleaseDescription = resDes
			return
		}
	}
	child.FloorResNotify.ReleaseRobots = append(child.FloorResNotify.ReleaseRobots, ResourceUseRobotInfo{
		RobotID:            robotid,
		ReleaseDescription: resDes,
		LockStart:          lockStart,
		ReleaseTime:        time.Now(),
	})
}

func (resnotify *ResourceNotify) Add(reschildnotify *ResourceNotify) {
	resnotify.childNotifys = append(resnotify.childNotifys, reschildnotify)
}

func (resnotify *ResourceNotify) Empty() bool {
	switch resnotify.ResType {
	case ResTypeLift:
		return resnotify.LiftResNotify == nil || resnotify.LiftResNotify.Empty()
	case ResTypeLiftCtrlArea, ResTypeCtrlArea:
		return resnotify.AreaResNotify == nil || resnotify.AreaResNotify.Empty()
	case ResTypeSafeWait:
		return resnotify.SafeResNotify == nil || resnotify.SafeResNotify.Empty()
	case ResTypeFloor:
		return resnotify.FloorResNotify == nil || resnotify.FloorResNotify.Empty()
	}
	return true
}

func (resnotify *ResourceNotify) SafeWaitLock(robotId, waitPoint string, startTime time.Time) {
	resnotify.SafeResNotify.CurrentUseWaitPoints = append(resnotify.SafeResNotify.CurrentUseWaitPoints, WaitPointResourceInfo{
		ResourceUseRobotInfo: ResourceUseRobotInfo{
			RobotID:   robotId,
			LockStart: startTime,
		},
		WaitPointID: waitPoint,
	})
}

func (resnotify *ResourceNotify) SafeWaitRelease(robotId, waitPoint, releaseDes string, lockTime time.Time) {
	if robotId == "" || waitPoint == "" {
		return
	}
	for i := 0; i < len(resnotify.SafeResNotify.ReleaseWaitPoints); i++ {
		if resnotify.SafeResNotify.ReleaseWaitPoints[i].RobotID == robotId &&
			resnotify.SafeResNotify.ReleaseWaitPoints[i].WaitPointID == waitPoint {
			return
		}
	}
	resnotify.SafeResNotify.ReleaseWaitPoints = append(resnotify.SafeResNotify.ReleaseWaitPoints, WaitPointResourceInfo{
		ResourceUseRobotInfo: ResourceUseRobotInfo{
			RobotID:            robotId,
			ReleaseTime:        time.Now(),
			LockStart:          lockTime,
			ReleaseDescription: releaseDes,
		},
		WaitPointID: waitPoint,
	})
}

//管制区锁定
func (resnotify *ResourceNotify) CtrAreaLock(check bool, robotId string, startTime time.Time) {
	if check {
		for i := 0; i < len(resnotify.AreaResNotify.CurrentRobotIDs); i++ {
			if resnotify.AreaResNotify.CurrentRobotIDs[i].RobotID == robotId {
				return
			}
		}
	}
	resnotify.AreaResNotify.CurrentRobotIDs = append(resnotify.AreaResNotify.CurrentRobotIDs, ResourceUseRobotInfo{
		RobotID:   robotId,
		LockStart: startTime,
	})
}

//管制区释放
func (resnotify *ResourceNotify) CtrAreaRelease(robotId, releaseDes string, lockTime time.Time) {
	for i := 0; i < len(resnotify.AreaResNotify.ReleaseRobotIDs); i++ {
		if resnotify.AreaResNotify.ReleaseRobotIDs[i].RobotID == robotId {
			return
		}
	}
	resnotify.AreaResNotify.ReleaseRobotIDs = append(resnotify.AreaResNotify.ReleaseRobotIDs, ResourceUseRobotInfo{
		RobotID:            robotId,
		ReleaseTime:        time.Now(),
		LockStart:          lockTime,
		ReleaseDescription: releaseDes,
	})
}

//管制区的等待点锁定
func (resnotify *ResourceNotify) CtrAreaWaitLock(check bool, robotId, waitPoint string, startTime time.Time) {
	if check {
		for i := 0; i < len(resnotify.AreaResNotify.CurrentUseWaitPoints); i++ {
			if resnotify.AreaResNotify.CurrentUseWaitPoints[i].RobotID == robotId &&
				resnotify.AreaResNotify.CurrentUseWaitPoints[i].WaitPointID == waitPoint {
				return
			}
		}
	}
	resnotify.AreaResNotify.CurrentUseWaitPoints = append(resnotify.AreaResNotify.CurrentUseWaitPoints, WaitPointResourceInfo{
		ResourceUseRobotInfo: ResourceUseRobotInfo{
			RobotID:   robotId,
			LockStart: startTime,
		},
		WaitPointID: waitPoint,
	})
}

//释放管制区的等待点
func (resnotify *ResourceNotify) CtrAreaWaitRelease(robotId, waitPoint, releaseDes string, lockTime time.Time) {
	if robotId == "" || waitPoint == "" {
		return
	}
	for i := 0; i < len(resnotify.AreaResNotify.ReleaseWaitPoints); i++ {
		if resnotify.AreaResNotify.ReleaseWaitPoints[i].RobotID == robotId &&
			resnotify.AreaResNotify.ReleaseWaitPoints[i].WaitPointID == waitPoint {
			return
		}
	}
	resnotify.AreaResNotify.ReleaseWaitPoints = append(resnotify.AreaResNotify.ReleaseWaitPoints, WaitPointResourceInfo{
		ResourceUseRobotInfo: ResourceUseRobotInfo{
			RobotID:            robotId,
			ReleaseTime:        time.Now(),
			LockStart:          lockTime,
			ReleaseDescription: releaseDes,
		},
		WaitPointID: waitPoint,
	})
}

//锁定电梯
func (resnotify *ResourceNotify) LiftLock(LiftId, robotId, buildId string, floor int, startTime time.Time) {
	if resnotify.ResType != ResTypeLift {
		return
	}
	if resnotify.LiftResNotify.LiftID == LiftId || resnotify.LiftResNotify.LiftID == "" {
		resnotify.LiftResNotify.LiftID = LiftId
		resnotify.Floor = int32(floor)
		resnotify.BuildID = buildId
		if resnotify.LiftResNotify.CurrentRobot == nil {
			resnotify.LiftResNotify.CurrentRobot = &ResourceUseRobotInfo{
				RobotID:   robotId,
				LockStart: startTime,
			}
		} else {
			resnotify.LiftResNotify.CurrentRobot.RobotID = robotId
			resnotify.LiftResNotify.CurrentRobot.LockStart = startTime
		}
	} else {
		//增加Child
		resnotify.childNotifys = append(resnotify.childNotifys, &ResourceNotify{
			ResType: ResTypeLift,
			Floor:   int32(floor),
			BuildID: buildId,
			LiftResNotify: &LiftResourceNotify{
				LiftID: LiftId,
				CurrentRobot: &ResourceUseRobotInfo{
					RobotID:   robotId,
					LockStart: startTime,
				},
			},
		})
	}
}

//释放电梯
func (resnotify *ResourceNotify) LiftRelease(LiftId, buildId, robotId, resDescription string, floor int, startTime time.Time) {
	if resnotify.ResType != ResTypeLift {
		return
	}
	if resnotify.LiftResNotify.LiftID == LiftId || resnotify.LiftResNotify.LiftID == "" {
		resnotify.LiftResNotify.LiftID = LiftId
		resnotify.Floor = int32(floor)
		resnotify.BuildID = buildId
		if resnotify.LiftResNotify.ReleaseRobot == nil {
			resnotify.LiftResNotify.ReleaseRobot = &ResourceUseRobotInfo{
				RobotID:            robotId,
				LockStart:          startTime,
				ReleaseTime:        time.Now(),
				ReleaseDescription: resDescription,
			}
		} else {
			resnotify.LiftResNotify.ReleaseRobot.RobotID = robotId
			resnotify.LiftResNotify.ReleaseRobot.LockStart = startTime
			resnotify.LiftResNotify.ReleaseRobot.ReleaseDescription = resDescription
			resnotify.LiftResNotify.ReleaseRobot.ReleaseTime = time.Now()
		}
	} else {
		resnotify.childNotifys = append(resnotify.childNotifys, &ResourceNotify{
			ResType: ResTypeLift,
			Floor:   int32(floor),
			BuildID: buildId,
			LiftResNotify: &LiftResourceNotify{
				LiftID: LiftId,
				ReleaseRobot: &ResourceUseRobotInfo{
					RobotID:            robotId,
					LockStart:          startTime,
					ReleaseTime:        time.Now(),
					ReleaseDescription: resDescription,
				},
			},
		})
	}
}

//电梯等待点的锁定
func (resnotify *ResourceNotify) LiftWaitLock(LiftId, buildId, robotId, waitPoint string, floor int, startTime time.Time) {
	if resnotify.ResType != ResTypeLift {
		return
	}
	var curResNotify *ResourceNotify //当前正在使用的等待点的信息
	if resnotify.LiftResNotify.LiftID == LiftId || resnotify.LiftResNotify.LiftID == "" {
		resnotify.LiftResNotify.LiftID = LiftId
		resnotify.Floor = int32(floor)
		resnotify.BuildID = buildId
		curResNotify = resnotify
	} else {
		for i := 0; i < len(resnotify.childNotifys); i++ {
			if resnotify.childNotifys[i].Floor == int32(floor) && resnotify.childNotifys[i].BuildID == buildId &&
				resnotify.childNotifys[i].LiftResNotify.LiftID == LiftId {
				curResNotify = resnotify.childNotifys[i]
				break
			}
		}
		if curResNotify == nil {
			curResNotify = &ResourceNotify{
				ResType: ResTypeLift,
				Floor:   int32(floor),
				BuildID: buildId,
				LiftResNotify: &LiftResourceNotify{
					LiftID:               LiftId,
					CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 4),
				},
			}
			resnotify.childNotifys = append(resnotify.childNotifys, curResNotify)
		}
	}
	curResNotify.LiftResNotify.CurrentUseWaitPoints = append(curResNotify.LiftResNotify.CurrentUseWaitPoints, WaitPointResourceInfo{
		ResourceUseRobotInfo: ResourceUseRobotInfo{
			RobotID:   robotId,
			LockStart: startTime,
		},
		WaitPointID: waitPoint,
	})
}

//电梯等待点的释放
func (resnotify *ResourceNotify) LiftWaitRelease(LiftId, buildId, robotId, waitPoint, releaseDes string, floor int, lockTime time.Time) {
	if robotId == "" || waitPoint == "" || resnotify.ResType != ResTypeLift {
		return
	}
	var curResNotify *ResourceNotify //当前正在使用的等待点的信息
	if resnotify.LiftResNotify.LiftID == LiftId || resnotify.LiftResNotify.LiftID == "" {
		resnotify.LiftResNotify.LiftID = LiftId
		resnotify.Floor = int32(floor)
		resnotify.BuildID = buildId
		curResNotify = resnotify
	} else {
		for i := 0; i < len(resnotify.childNotifys); i++ {
			if resnotify.childNotifys[i].Floor == int32(floor) && resnotify.childNotifys[i].BuildID == buildId &&
				resnotify.childNotifys[i].LiftResNotify.LiftID == LiftId {
				curResNotify = resnotify.childNotifys[i]
				break
			}
		}
		if curResNotify == nil {
			curResNotify = &ResourceNotify{
				ResType: ResTypeLift,
				Floor:   int32(floor),
				BuildID: buildId,
				LiftResNotify: &LiftResourceNotify{
					LiftID:               LiftId,
					CurrentUseWaitPoints: make([]WaitPointResourceInfo, 0, 4),
				},
			}
			resnotify.childNotifys = append(resnotify.childNotifys, curResNotify)
		}
	}

	curResNotify.LiftResNotify.ReleaseWaitPoints = append(curResNotify.LiftResNotify.ReleaseWaitPoints, WaitPointResourceInfo{
		ResourceUseRobotInfo: ResourceUseRobotInfo{
			RobotID:            robotId,
			ReleaseTime:        time.Now(),
			LockStart:          lockTime,
			ReleaseDescription: releaseDes,
		},
		WaitPointID: waitPoint,
	})
}

func (resnotify *ResourceNotify) createRabbitMQMsg() rabbitmq.Message {
	var notifys []*ResourceNotify
	l := len(resnotify.childNotifys)

	if resnotify.ResType == ResTypeLift {
		if resnotify.LiftResNotify.LiftID == "" { //本身不提交
			notifys = make([]*ResourceNotify, 0, l)
		} else {
			notifys = make([]*ResourceNotify, 0, l+1)
			if !resnotify.Empty() {
				notifys = append(notifys, resnotify)
			}
		}
	} else {
		notifys = make([]*ResourceNotify, 0, l+1)
		if !resnotify.Empty() {
			notifys = append(notifys, resnotify)
		}
	}
	if l > 0 {
		for i := 0; i < len(resnotify.childNotifys); i++ {
			if !resnotify.childNotifys[i].Empty() {
				notifys = append(notifys, resnotify.childNotifys[i])
			}
		}
	}
	if len(notifys) == 0 {
		return rabbitmq.Message{}
	}
	return rabbitmq.NewMessage(notifys)
}

//资源占有类型变动的通知消息
// dispatch_srv/monitor/#
func NotifyResUseChange(officeId string, notifyInfo *ResourceNotify) error {
	//构造router
	router := strings.Join([]string{MonitorRouter, officeId}, "")
	msg := notifyInfo.createRabbitMQMsg()
	if msg.MessageID == "" {
		return nil
	}
	return rabbitmq.Publish(ResourceStatusExchange, router, msg)
}

//订阅资源占用变动消息
func SubResUseChangeNotify(handler func(officeId string, notifyInfo ...ResourceNotify)) error {
	if handler == nil {
		return errors.New("必须指定一个订阅处理过程")
	}
	consume, err := rabbitmq.DefaultRMQ.RegisterConsume(ResourceStatusExchange, "", true, func(delivery amqp.Delivery) {
		var msg rabbitmq.Message
		var notifyData []ResourceNotify
		msg.Data = &notifyData
		err := json.Unmarshal(delivery.Body, &msg)
		if err != nil {
			log.WithError(err).Error("解析资源占用通知信息错误")
			return
		}
		officeid := delivery.RoutingKey[len(MonitorRouter):]
		handler(officeid, notifyData...)
	})
	if err != nil {
		return err
	}
	return consume.Subscribe(MonitorRouter + "#")
}
