package cache

import (
	"common/biz/dto"
	"common/biz/enum"
	"common/biz/manager"
	"common/log"
	"common/redis"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/suiyunonghen/DxCommonLib"
	"github.com/suiyunonghen/dxsvalue"
	"strings"
	"time"
)

//机器人的任务信息缓存

type JsonTime struct {
	time.Time
}

var (
	ErrJobUnExists = errors.New("不存在的任务")
)

func (jtime *JsonTime) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(`"2006-01-02 15:04:05"`, DxCommonLib.FastByte2String(data))
	(*jtime).Time = t
	return err
}

func (jtime JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(jtime.Format(`"2006-01-02 15:04:05"`)), nil
}

type BaseJobInfo struct {
	Back           bool                  //是否需要返程
	Origin         enum.MsgOriginEnum    //任务来源
	JobType        enum.JobTypeEnum      //任务类型
	JobState       enum.JobStatusEnum    //任务状态
	AcceptState    enum.AcceptStatusEnum `json:",omitempty"` //物品接收状态
	RobotID        string                `json:",omitempty"` //这个任务绑定的机器人，如果任务开始执行的话
	JobCreator     string                `json:",omitempty"` //任务创建者
	StartPosId     string                `json:",omitempty"` //开始的位置
	StartPosName   string                `json:",omitempty"` //开始的位置名称
	StartBuild     string                `json:",omitempty"` //开始的楼宇
	StartBuildName string                `json:",omitempty"` //开始的楼宇名称
	StartFloor     int                   `json:",omitempty"` //开始的楼层
	EndPositionID  string                `json:",omitempty"` //任务要去的目标地点
	EndPosName     string                `json:",omitempty"` //地址
	EndBuildId     string                `json:",omitempty"` //任务要去的目标楼宇
	EndBuild       string                `json:",omitempty"` //任务要去的楼宇名称
	Endfloor       int                   `json:",omitempty"` //任务的目标楼层
	Description    string                `json:",omitempty"` //任务描述
	JobId          string                `json:",omitempty"` //当前的任务ID
	Jobgroup       string                `json:",omitempty"` //任务组ID
	CreateTime     JsonTime              `json:",omitempty"` //任务创建时间
	ArrivedTime    JsonTime              `json:",omitempty"` //到达时间
	ReplayTime     JsonTime              `json:",omitempty"` //机器人回应本任务的时间，一般是呼叫类型的任务会有
	StartTime      JsonTime              `json:",omitempty"` //任务的开始执行时间
	EndTime        JsonTime              `json:",omitempty"` //任务的完成时间
	EndUserID      string                `json:",omitempty"` //完成任务的用户ID
	Distance       int                   `json:",omitempty"` //任务里程（米）
	Remarks        string                `json:",omitempty"` //描述信息，任务结束原因
}

//机器人任务信息
type JobInfo struct {
	RobotStatus enum.RobotStatusEnum //任务状态变动过程中的机器人状态
	BaseJobInfo
	Order int //任务优先级
}

/**
获取某个机器人的所有的任务信息
robotId=""，的时候，返回这个机构下未分配的任务信息
*/
func GetRobotJobInfo(officeId, robotId string) ([]JobInfo, error) {
	if officeId == "" {
		return nil, nil
	}
	rediskey := ""
	if robotId == "" {
		rediskey = strings.Join([]string{"jobs", officeId, "_"}, ":")
	} else {
		rediskey = strings.Join([]string{"jobs", officeId, robotId}, ":")
	}
	jsonValues, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil {
		if redis.IsRedisNil(err) {
			err = nil
		}
		return nil, err
	}
	result := make([]JobInfo, len(jsonValues))
	for i := 0; i < len(jsonValues); i++ {
		json.Unmarshal(DxCommonLib.FastString2Byte(jsonValues[i]), &result[i])
	}
	return result, nil
}

//GetRobotJob 获取单个未完成的任务
func GetRobotJob(officeID, robotID, jobID string) (*JobInfo, error) {
	jobInfos, err := GetRobotJobInfo(officeID, robotID)
	if err != nil {
		return nil, err
	}
	if jobInfos != nil {
		for _, jobInfo := range jobInfos {
			if jobInfo.JobId == jobID {
				return &jobInfo, nil
			}
		}
	}
	return nil, ErrJobUnExists
}

//FindRobotJobByOfficeID 获取机构下所有正在执行和排队中的任务.
func FindRobotJobByOfficeID(officeID string) ([]JobInfo, error) {
	pattern := fmt.Sprintf("%s:%s:*", "jobs", officeID)
	keys, err := redis.Keys(context.Background(), pattern).Result()
	if err != nil {
		return nil, err
	}
	result := make([]JobInfo, 0, 32)
	var job JobInfo
	for _, key := range keys {
		dataArray, err := redis.LRange(context.Background(), key, 0, -1).Result()
		if err != nil {
			continue
		}
		for _, data := range dataArray {
			err := json.Unmarshal(DxCommonLib.FastString2Byte(data), &job)
			if err != nil {
				return result, err
			}
			result = append(result, job)
		}
	}
	return result, nil
}

//点位信息
type PointInfo struct {
	RobotType manager.RobotType //真实的机器人类型
	Pointguid string
	PointName string
	Building  string
	BuildName string
	Floor     int
}

func saveJobGroupInfo(robotJobGrop *dto.ApplyRobotJob, findpos func(office interface{}, pointinfo *PointInfo),
	office interface{}, rediskey, lockid string, posInfo *RobotPosInfo, canDispCharge func() bool) (runfirstjob bool, jobs []JobInfo, err error, cancelBackJobs []JobInfo) {
	//增加任务信息
	var jbinfo JobInfo
	jbinfo.RobotStatus = posInfo.Status
	replace := false
	var pipe redis2.Pipeliner
	if robotJobGrop.JobType == enum.JtBack {
		//返程任务，需要判定一下，后面是否有任务，如果有任务，就不执行返程
		runJob, err := redis.LIndex(context.Background(), rediskey, 0).Result()
		if err != nil && !redis.IsRedisNil(err) {
			redis.SimpleUnLock(lockid)
			return false, nil, err, nil
		}
		if runJob != "" {
			redis.SimpleUnLock(lockid)
			return false, nil, errors.New("后续还有任务，无法插入返程任务进行执行"), nil
		}
		//没有任务，可以返程
		addOK, jobs, err := addJobGroupInfo2Cache(robotJobGrop, true, findpos, office, posInfo, redis.Client.TxPipeline(), rediskey, lockid)
		return addOK, jobs, err, nil
	} else if robotJobGrop.JobType == enum.JtCall || robotJobGrop.JobType == enum.JtCharge {
		//合并
		values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
		if err != nil && !redis.IsRedisNil(err) {
			redis.SimpleUnLock(lockid)
			return false, nil, err, nil
		}
		totalLen := len(values)
		jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
		defer dxsvalue.FreeValue(jsovalue)
		if totalLen == 1 {
			//只有一条记录，判定一下这条记录，如果是返程或者充电的，那么直接覆盖掉这调记录
			err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[0]), true)
			if err != nil {
				pipe = redis.Client.TxPipeline()
				pipe.LSet(context.Background(), rediskey, 0, "delete")
				totalLen = 0
			} else {
				json2JobInfo(jsovalue, &jbinfo.BaseJobInfo)
				if robotJobGrop.JobType == enum.JtCharge && jbinfo.JobType == enum.JtCharge {
					redis.SimpleUnLock(lockid)
					return false, nil, errors.New("机器人已经在充电任务中"), nil
				}
				if robotJobGrop.JobType == enum.JtCall && (jbinfo.JobType == enum.JtBack || jbinfo.JobType == enum.JtCharge && (canDispCharge == nil || canDispCharge())) {
					//充电中，判定一下电量是否充足
					//返程中，直接返回
					//终止掉之前的任务
					cancelBackJobs = append(cancelBackJobs, jbinfo)
					pipe = redis.Client.TxPipeline()
					pipe.LPop(context.Background(), rediskey)
					totalLen = 0
				}
			}
		} else if totalLen > 1 {
			for i := 0; i < len(values); i++ {
				err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
				if err != nil {
					pipe.LSet(context.Background(), rediskey, int64(i), "delete")
					totalLen--
					continue
				}
				json2JobInfo(jsovalue, &jbinfo.BaseJobInfo)
				runfirstjob = i == 0
				if jbinfo.Jobgroup == robotJobGrop.GroupID && jbinfo.JobId == robotJobGrop.Jobs[0].JobID {
					redis.SimpleUnLock(lockid)
					jobs = append(jobs, jbinfo)
					runfirstjob = runfirstjob && jbinfo.JobState < enum.JsStated
					return runfirstjob, jobs, nil, cancelBackJobs
				}
				if (jbinfo.JobType == enum.JtCall && jbinfo.EndPositionID == robotJobGrop.Jobs[0].GUID || //都是呼叫，并且目标地点相同
					jbinfo.JobType == enum.JtCharge) && (jbinfo.JobState < enum.JsStated) { //都是充电任务，并且是还没执行的
					//原始任务合并为新任务，冲掉
					cancelBackJobs = append(cancelBackJobs, jbinfo)

					jobs = append(jobs, jbinfo)
					jb := &jobs[len(jobs)-1]
					jb.JobId = robotJobGrop.Jobs[0].JobID
					jb.Jobgroup = robotJobGrop.GroupID
					jb.JobCreator = robotJobGrop.UserID
					jb.CreateTime = JsonTime{Time: robotJobGrop.Time}
					jb.StartPosId = posInfo.PosId
					jb.StartPosName = posInfo.PosName
					jb.StartFloor = posInfo.Floor
					jb.StartBuild = posInfo.BuildId
					jb.StartBuildName = posInfo.BuildName
					jsovalue.SetKeyString("JobCreator", jb.JobCreator)
					jsovalue.SetKeyString("JobId", jb.JobId)
					jsovalue.SetKeyString("Jobgroup", jb.Jobgroup)
					newbt := make([]byte, 0, len(values[i]))
					newbt = jsovalue.ToJson(false, dxsvalue.JSE_NoEscape, false, newbt)

					pipe.LSet(context.Background(), rediskey, int64(i), newbt) //重置
					pipe.LRem(context.Background(), rediskey, 0, "delete")
					pipe.Del(context.Background(), lockid)
					_, err = pipe.Exec(context.Background())
					return runfirstjob, jobs, err, cancelBackJobs
				} else if jbinfo.JobType == enum.JtBack && i != len(values)-1 {
					//移除返程
					pipe.LSet(context.Background(), rediskey, int64(i), "delete")
					cancelBackJobs = append(cancelBackJobs, jbinfo)
					totalLen--
				}
			}
		}
		replace = totalLen == 0
		runfirstjob = replace //没有任务了，立即执行
	} else {
		//先判定一下，正在执行的任务，是否是返程和充电任务，如果是，先取消掉相关任务
		firstJobs, err := redis.LRange(context.Background(), rediskey, 0, 1).Result()
		if err != nil {
			if !redis.IsRedisNil(err) {
				redis.SimpleUnLock(lockid)
				return false, nil, err, nil
			}
			replace = true
			runfirstjob = true
		} else {
			replace = len(firstJobs) == 0
			runfirstjob = replace
			if len(firstJobs) == 1 {
				//只有一条记录，并且是充电和返程的，取消掉
				jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
				err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(firstJobs[0]), true)
				if err != nil {
					pipe = redis.Client.TxPipeline()
					pipe.LPop(context.Background(), rediskey)
					replace = true
					runfirstjob = true
				} else {
					json2JobInfo(jsovalue, &jbinfo.BaseJobInfo)
					if jbinfo.JobType == enum.JtBack || jbinfo.JobType == enum.JtCharge && (canDispCharge == nil || canDispCharge()) {
						//取消掉
						pipe = redis.Client.TxPipeline()
						cancelBackJobs = append(cancelBackJobs, jbinfo)
						pipe.LPop(context.Background(), rediskey)
						replace = true
						runfirstjob = true
					}
				}
				dxsvalue.FreeValue(jsovalue)
			}
		}
	}
	//增加新的
	if pipe == nil {
		pipe = redis.Client.TxPipeline()
	}
	addok, jobs, err := addJobGroupInfo2Cache(robotJobGrop, replace, findpos, office, posInfo, pipe, rediskey, lockid)
	return addok && runfirstjob, jobs, err, cancelBackJobs
}

func addJobGroupInfo2Cache(robotJobGrop *dto.ApplyRobotJob, replace bool,
	findpos func(office interface{}, pointinfo *PointInfo), office interface{}, posInfo *RobotPosInfo, pipe redis2.Pipeliner,
	rediskey, lockid string) (addOk bool, jobs []JobInfo, err error) {
	var jobinfo JobInfo
	jobinfo.StartTime = JsonTime{Time: time.Now()} //开始时间，立即开始执行
	jobs = make([]JobInfo, 0, len(robotJobGrop.Jobs))
	jobinfo.RobotID = robotJobGrop.RobotID
	jobinfo.Jobgroup = robotJobGrop.GroupID

	jobinfo.StartPosId = posInfo.PosId
	jobinfo.StartPosName = posInfo.PosName
	jobinfo.StartBuildName = posInfo.BuildName
	jobinfo.StartBuild = posInfo.BuildId
	jobinfo.StartFloor = posInfo.Floor

	jobinfo.CreateTime = JsonTime{Time: robotJobGrop.Time}

	jobinfo.JobType = robotJobGrop.JobType
	jobinfo.RobotStatus = posInfo.Status
	jobinfo.JobCreator = robotJobGrop.UserID
	var pinfo PointInfo
	pinfo.RobotType = posInfo.RobotType
	for i := 0; i < len(robotJobGrop.Jobs); i++ {
		if robotJobGrop.JobType != enum.JtBack && robotJobGrop.JobType != enum.JtCharge &&
			robotJobGrop.Jobs[i].GUID == "" {
			log.Errorf("任务%s未指定一个有效的目标地址", robotJobGrop.Jobs[i].JobID)
			continue
		}
		if i == 0 && replace && posInfo.Status >= 0 {
			jobinfo.JobState = enum.JsStated
		} else {
			jobinfo.JobState = enum.JsQueue
		}
		jobinfo.JobId = robotJobGrop.Jobs[i].JobID

		if robotJobGrop.Jobs[i].BuildID != "" && robotJobGrop.Jobs[i].Floor > -1000 {
			jobinfo.EndPositionID = robotJobGrop.Jobs[i].GUID
			jobinfo.EndBuildId = robotJobGrop.Jobs[i].BuildID
			jobinfo.EndBuild = robotJobGrop.Jobs[i].BuildName
			jobinfo.EndPosName = robotJobGrop.Jobs[i].PosName
			jobinfo.Endfloor = pinfo.Floor
		} else if findpos != nil && robotJobGrop.Jobs[i].GUID != "" && posInfo.RobotType != "" {
			pinfo.Pointguid = robotJobGrop.Jobs[i].GUID
			findpos(office, &pinfo)
			jobinfo.EndPositionID = pinfo.Pointguid
			jobinfo.EndBuildId = pinfo.Building
			jobinfo.EndBuild = pinfo.BuildName
			jobinfo.EndPosName = pinfo.PointName
			jobinfo.Endfloor = pinfo.Floor
		} else {
			jobinfo.EndPositionID = robotJobGrop.Jobs[i].GUID
			jobinfo.EndBuildId = ""
			jobinfo.EndBuild = ""
			jobinfo.EndPosName = ""
			jobinfo.Endfloor = 0
		}

		if robotJobGrop.JobType != enum.JtBack && robotJobGrop.JobType != enum.JtCharge &&
			jobinfo.EndPositionID == "" {
			log.Errorf("点位%s不存在", robotJobGrop.Jobs[i].GUID)
			continue
		}

		if i == len(robotJobGrop.Jobs)-1 {
			jobinfo.Back = robotJobGrop.Back //最后一个设定一下是否需要返程
		} else {
			jobinfo.Back = false
		}
		bt, _ := json.Marshal(jobinfo)
		pipe.RPush(context.Background(), rediskey, bt) //添加值
		jobs = append(jobs, jobinfo)
		jobinfo.StartPosId = ""
		jobinfo.StartPosName = ""
		jobinfo.StartBuildName = ""
		jobinfo.StartBuild = ""
	}
	if len(jobs) > 0 {
		pipe.Del(context.Background(), lockid) //删除锁
		_, err = pipe.Exec(context.Background())
		if err == nil {
			log.Debugf("groupid=%s,jobs=%s任务信息更新到Redis缓存,", jobs[0].Jobgroup, jobs[0].JobId)
			return true, jobs, nil
		}
		return err == nil, jobs, err
	}
	redis.SimpleUnLock(lockid)
	return false, nil, errors.New("未发现有效的任务内容")
}

type RobotPosInfo struct {
	Floor     int
	BuildId   string
	BuildName string
	PosId     string
	PosName   string
	Status    enum.RobotStatusEnum
	RobotType manager.RobotType
}

//将一系列的任务加入到redis中
func SaveJobGroupInfo(robotJobGrop *dto.ApplyRobotJob, replace bool, findpos func(office interface{}, pointinfo *PointInfo),
	office interface{}, posInfo *RobotPosInfo, canDispCharge func() bool) (runfirstjob bool, jobs []JobInfo, err error, cancelBackJobs []JobInfo) {
	if robotJobGrop.RobotID == "" || robotJobGrop.OfficeID == "" || len(robotJobGrop.Jobs) == 0 {
		return false, nil, nil, nil
	}
	rediskey := ""
	if robotJobGrop.RobotID == "" {
		rediskey = strings.Join([]string{"jobs", robotJobGrop.OfficeID, "_"}, ":")
	} else {
		//指定了机器人
		rediskey = strings.Join([]string{"jobs", robotJobGrop.OfficeID, robotJobGrop.RobotID}, ":")
	}
	lockid := rediskey + ":lock"
	if err = redis.SimpleLock(lockid); err != nil {
		log.WithError(err).Errorf("锁定Redis资源%s失败", lockid)
		return false, nil, err, nil
	}
	if !replace {
		//需要去重处理
		return saveJobGroupInfo(robotJobGrop, findpos, office, rediskey, lockid, posInfo, canDispCharge)
	}
	//replace模式
	pipe := redis.Client.Pipeline()
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil && !redis.IsRedisNil(err) {
		redis.SimpleUnLock(lockid)
		return false, nil, err, nil
	}
	jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	totalLen := len(values)
	var jbInfo JobInfo
	if totalLen > 0 {
		cancelBackJobs = make([]JobInfo, 0, totalLen)
	}
	for i := 0; i < totalLen; i++ {
		err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		if err != nil {
			continue
		}
		json2JobInfo(jsovalue, &jbInfo.BaseJobInfo)
		cancelBackJobs = append(cancelBackJobs, jbInfo)
	}
	//将redis中其他的任务，全部认为是清理掉的
	pipe.LTrim(context.Background(), rediskey, 1, 0) //清空list
	runfirstjob, jobs, err = addJobGroupInfo2Cache(robotJobGrop, true, findpos, office, posInfo, pipe, rediskey, lockid)
	return
}

//执行消毒任务ApplyDisinfectJobVO,是要立即执行的，直接将任务前提到第一个位置，然后执行
func RunningDisinfectJob(disinfectJob *dto.ApplyDisinfectJob, robotStatus enum.RobotStatusEnum) (err error, cancelBackJobs []BaseJobInfo) {
	if disinfectJob.RobotID == "" || disinfectJob.OfficeID == "" || len(disinfectJob.Areas) == 0 {
		return errors.New("无效的消毒任务信息"), nil
	}
	rediskey := strings.Join([]string{"jobs", disinfectJob.OfficeID, disinfectJob.RobotID}, ":")
	lockid := rediskey + ":lock"
	if err := redis.SimpleLock(lockid); err != nil {
		return err, nil
	}
	//先查看一下是否有一个同类型的任务，如果有同类型的任务，就不执行
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil && !redis.IsRedisNil(err) {
		redis.SimpleUnLock(lockid)
		return err, nil
	}

	pipe := redis.Client.TxPipeline()
	hasDel := false
	isFirst := true
	jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	for i := 0; i < len(values); i++ {
		err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		if err != nil {
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			continue
		}
		if isFirst {
			isFirst = false
			//取消正在执行的返程
			if enum.JobTypeEnum(jsovalue.AsInt("JobType", 0)) == enum.JtBack {
				var jbinfo BaseJobInfo
				json2JobInfo(jsovalue, &jbinfo)
				cancelBackJobs = []BaseJobInfo{jbinfo}
				pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			}
		}

		if jsovalue.AsInt("JobType", 0) == int(disinfectJob.JobType) &&
			jsovalue.AsString("JobId", "") == disinfectJob.JobID &&
			jsovalue.AsString("Jobgroup", "") == disinfectJob.GroupID &&
			jsovalue.AsString("RobotID", "") == disinfectJob.RobotID {
			//任务已经存在，
			hasDel = true
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			break
		}
	}
	err = nil
	jsovalue.Clear()
	dxsvalue.FreeValue(jsovalue)

	var jobinfo JobInfo
	jobinfo.RobotID = disinfectJob.RobotID
	jobinfo.JobState = enum.JsCalling
	jobinfo.JobId = disinfectJob.JobID
	jobinfo.RobotStatus = robotStatus
	jobinfo.Jobgroup = disinfectJob.GroupID
	jobinfo.JobType = disinfectJob.JobType
	jobinfo.Origin = disinfectJob.Origin
	jobinfo.JobCreator = disinfectJob.UserID
	jobinfo.Description = disinfectJob.TaskName + "[" + disinfectJob.TaskID + "]"
	jobinfo.CreateTime.Time = disinfectJob.Time

	bt, _ := json.Marshal(jobinfo)
	if hasDel {
		pipe.LRem(context.Background(), rediskey, 0, "delete")
	}
	pipe.LPush(context.Background(), rediskey, bt) //设置新值
	pipe.Del(context.Background(), lockid)         //删除锁
	_, err = pipe.Exec(context.Background())
	return err, cancelBackJobs
}

//将某个任务信息投递推送到某个机构的机器人下
func SaveJobInfo(officeId string, jobinfo BaseJobInfo, save2First bool) error {
	if jobinfo.JobType == enum.JtBack {
		return errors.New("返程任务不支持该接口接入")
	}
	if officeId == "" || jobinfo.JobId == "" {
		return errors.New("officeID和jobID不能为空")
	}
	if jobinfo.RobotID == "" {
		return errors.New("必须指定机器人信息")
	}
	rediskey := strings.Join([]string{"jobs", officeId, jobinfo.RobotID}, ":")
	lockid := rediskey + ":lock"
	if err := redis.SimpleLock(lockid); err != nil {
		return err
	}
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil && !redis.IsRedisNil(err) {
		redis.SimpleUnLock(lockid)
		return err
	}
	isSpecialTimeJob := jobinfo.Description == "分时段任务调度" //判定是否是分时段调度任务
	bt, _ := json.Marshal(jobinfo)
	jsonvalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	for i := 0; i < len(values); i++ {
		jsonvalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		if jsonvalue.AsString("JobId", "") == jobinfo.JobId &&
			jsonvalue.AsString("Jobgroup", "") == jobinfo.Jobgroup {
			//使用Pipe模式
			pipe := redis.Client.Pipeline()
			pipe.LSet(context.Background(), rediskey, int64(i), bt) //设置新值
			pipe.Del(context.Background(), lockid)
			_, err := pipe.Exec(context.Background())
			return err
		} else if jobinfo.JobType == enum.JtCharge && jsonvalue.AsInt("JobType", 0) == int(enum.JtCharge) {
			//有充电任务了
			pipe := redis.Client.Pipeline()
			if save2First {
				//判定第一个是否就是充电的
				if i == 0 { //如果就是充电的
					jsonvalue.SetKeyCached("JobState", dxsvalue.VT_Int, jsonvalue.ValueCache()).SetInt(int64(jobinfo.JobState))
					if !jobinfo.CreateTime.IsZero() {
						jsonvalue.SetKeyCached("CreateTime", dxsvalue.VT_DateTime, jsonvalue.ValueCache()).SetTime(jobinfo.CreateTime.Time)
					}
					if !jobinfo.StartTime.IsZero() {
						jsonvalue.SetKeyCached("StartTime", dxsvalue.VT_DateTime, jsonvalue.ValueCache()).SetTime(jobinfo.StartTime.Time)
					}
					if !jobinfo.ReplayTime.IsZero() {
						jsonvalue.SetKeyCached("ReplayTime", dxsvalue.VT_DateTime, jsonvalue.ValueCache()).SetTime(jobinfo.ReplayTime.Time)
					}
					newbt := make([]byte, 0, len(values[i]))
					newbt = jsonvalue.ToJson(false, dxsvalue.JSE_NoEscape, false, newbt)
					pipe.LSet(context.Background(), rediskey, int64(i), newbt)
				} else {
					pipe.LSet(context.Background(), rediskey, int64(i), "delete")
					pipe.LRem(context.Background(), rediskey, 0, "delete")
					pipe.LPush(context.Background(), rediskey, bt)
				}
			} else {
				pipe.LSet(context.Background(), rediskey, int64(i), bt)
			}
			pipe.Del(context.Background(), lockid)
			_, err := pipe.Exec(context.Background())
			return err
		} else if isSpecialTimeJob && jobinfo.Description == jsonvalue.AsString("Description", "") {
			//分时段调度任务，需要合并
			if i == 0 { //如果第一个就是当前正执行的任务，就投递
				redis.SimpleUnLock(lockid)
				return nil
			}
		}
	}
	pipe := redis.Client.Pipeline()
	if save2First {
		pipe.LPush(context.Background(), rediskey, bt) //设置新值
	} else {
		pipe.RPush(context.Background(), rediskey, bt) //设置新值
	}
	pipe.Del(context.Background(), lockid) //删除锁
	_, err = pipe.Exec(context.Background())
	if err != nil {
		err = fmt.Errorf("机器人%s任务组%s的任务%s存入redis失败：%w", jobinfo.RobotID, jobinfo.Jobgroup, jobinfo.JobId, err)
	}
	return err
}

//检查机器人的Redis缓存任务信息
func CheckRobotRedisJobs(officeId, robotId string, ntime *time.Time) (invalidJobs []BaseJobInfo, err error) {
	if robotId == "" {
		return nil, errors.New("必须指定机器人信息")
	}
	rediskey := ""
	rediskey = strings.Join([]string{"jobs", officeId, robotId}, ":")
	lockid := rediskey + ":lock"
	if err = redis.SimpleLock(lockid); err != nil {
		return nil, err
	}
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil && !redis.IsRedisNil(err) {
		redis.SimpleUnLock(lockid)
		return nil, err
	}
	pipe := redis.Client.TxPipeline()
	invalidJobs = make([]BaseJobInfo, 0, 4)
	jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	var jbinfo, lastBackJob BaseJobInfo
	lastBackIndex := -1
	hasDel := false
	for i := 0; i < len(values); i++ {
		err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		if err != nil {
			hasDel = true
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			continue
		}
		if lastBackIndex != -1 {
			//返程任务后面还有其他任务，返程任务终止掉，直接执行其他任务
			pipe.LSet(context.Background(), rediskey, int64(lastBackIndex), "delete")
			if lastBackJob.JobState < enum.JsArrived {
				lastBackJob.JobState = enum.JsNewTask
			} else {
				lastBackJob.JobState = enum.JsCancelExpiry
			}
			invalidJobs = append(invalidJobs, lastBackJob)
			lastBackIndex = -1
		}
		json2JobInfo(jsovalue, &jbinfo)
		if jbinfo.JobState <= enum.JsCalling && ntime.Sub(jbinfo.CreateTime.Time) >= time.Hour*3 {
			//超过3小时还没开始，就认为是过期了，删除任务
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			jbinfo.JobState = enum.JsCancelExpiry
			invalidJobs = append(invalidJobs, jbinfo)
		} else if jbinfo.JobType == enum.JtBack {
			//判定一下是否有返程的，并且返程之后还有其他任务的，如果返程之后，有其他任务，就需要删除上一个返程
			lastBackIndex = i
			lastBackJob = jbinfo
		}
	}
	if hasDel || len(invalidJobs) > 0 {
		pipe.LRem(context.Background(), rediskey, 0, "delete")
		pipe.Del(context.Background(), lockid) //删除锁
		_, err = pipe.Exec(context.Background())
	} else {
		redis.SimpleUnLock(lockid)
	}
	return invalidJobs, err
}

//更新某个机器人的任务的状态
func UpdateJobInfoStatus(officeId, jobId, groupId, robotId string, status enum.JobStatusEnum, statusTime time.Time) (result BaseJobInfo, invalidJobs []BaseJobInfo, err error) {
	if robotId == "" {
		return BaseJobInfo{}, nil, errors.New("必须指定机器人信息")
	}
	rediskey := ""
	rediskey = strings.Join([]string{"jobs", officeId, robotId}, ":")
	lockid := rediskey + ":lock"
	if err = redis.SimpleLock(lockid); err != nil {
		return BaseJobInfo{}, nil, err
	}
	//更新状态的一般是在第一个，如果不是第一个，要将他移动到第一个位置
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil && !redis.IsRedisNil(err) {
		redis.SimpleUnLock(lockid)
		return BaseJobInfo{}, nil, err
	}
	pipe := redis.Client.TxPipeline()
	hasSetCmd := false
	hasJob := false
	jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	beforeList := make([]BaseJobInfo, 0, 4)
	for i := 0; i < len(values); i++ {
		jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		json2JobInfo(jsovalue, &result)
		beforeList = append(beforeList, result)
		if result.RobotID == robotId && result.JobId == jobId && (groupId == "" || result.Jobgroup == groupId) {
			result.JobState = status
			hasJob = true
			if status >= enum.JsCompleted { //完成或者取消的任务
				//完成的任务，删除掉
				result.EndTime.Time = time.Now()
				pipe.LSet(context.Background(), rediskey, int64(i), "delete")
				pipe.LRem(context.Background(), rediskey, 0, "delete")
				hasSetCmd = true
			} else if jsovalue.AsInt("JobState", 0) == int(status) {
				//已经修改为这个状态了
				if i != 0 && status > enum.JsCalling { //说明之前的任务已经全部无效了，删除
					for k := 0; k < i; k++ {
						pipe.LPop(context.Background(), rediskey)
					}
					hasSetCmd = true
					invalidJobs = beforeList[:i]
				}
			} else {
				jsovalue.SetKeyCached("JobState", dxsvalue.VT_Int, jsovalue.ValueCache()).SetInt(int64(status))
				if statusTime.IsZero() {
					statusTime = time.Now()
				}
				switch status {
				case enum.JsCalling:
					result.StartTime.Time = statusTime
					jsovalue.SetKeyCached("StartTime", dxsvalue.VT_DateTime, jsovalue.ValueCache()).SetTime(statusTime)
				case enum.JsStated:
					result.ReplayTime.Time = statusTime
					jsovalue.SetKeyCached("ReplayTime", dxsvalue.VT_DateTime, jsovalue.ValueCache()).SetTime(statusTime)
				case enum.JsArrived:
					result.ArrivedTime.Time = statusTime
					jsovalue.SetKeyCached("ArrivedTime", dxsvalue.VT_DateTime, jsovalue.ValueCache()).SetTime(statusTime)
				}
				dstb := make([]byte, 0, 128)
				dstb = jsovalue.ToJson(false, dxsvalue.JSE_NoEscape, false, dstb)
				if i != 0 && status > enum.JsCalling {
					//说明之前的任务已经全部无效了，删除
					for k := 0; k < i; k++ {
						pipe.LPop(context.Background(), rediskey)
					}
					invalidJobs = beforeList[:i]
				} else {
					pipe.LSet(context.Background(), rediskey, int64(i), dstb)
				}
				hasSetCmd = true
			}
			break
		}

	}
	if !hasJob {
		result.JobId = ""
		result.Jobgroup = ""
		result.JobCreator = ""
		result.EndPositionID = ""
		err = ErrJobUnExists
	}
	dxsvalue.FreeValue(jsovalue)
	if hasSetCmd {
		pipe.Del(context.Background(), lockid) //删除锁
		_, err = pipe.Exec(context.Background())
	} else {
		redis.SimpleUnLock(lockid)
	}
	return result, invalidJobs, err
}

//删除某个机器人的充电任务,并返回机器人的下一个任务
func RemoveChargeJob(officeid, robotId string) JobInfo {
	if robotId == "" {
		return JobInfo{}
	}
	rediskey := strings.Join([]string{"jobs", officeid, robotId}, ":")
	lockid := rediskey + ":lock"
	if err := redis.SimpleLock(lockid); err != nil {
		log.WithError(err).Errorf("锁定资源%s失败", lockid)
		return JobInfo{}
	}
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil {
		redis.SimpleUnLock(lockid)
		return JobInfo{}
	}
	jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	for i := 0; i < len(values); i++ {
		jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		if jsovalue.AsInt("JobType", 0) == int(enum.JtCharge) { //充电的删除
			//删除这个,使用pipeline模式
			/*
				MULTI
				incr tx_pipeline_counter
				expire tx_pipeline_counter 3600
				EXEC
			*/
			pipe := redis.Client.TxPipeline()
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			pipe.LRem(context.Background(), rediskey, 0, "delete")
			pipe.Del(context.Background(), lockid) //删除锁
			_, err := pipe.Exec(context.Background())
			dxsvalue.FreeValue(jsovalue)
			if err != nil {
				return JobInfo{}
			}
			if i == 0 && i < len(values)-1 {
				var result JobInfo
				json.Unmarshal(DxCommonLib.FastString2Byte(values[i+1]), &result)
				return result
			}
			return JobInfo{}
		}
	}
	redis.SimpleUnLock(lockid)
	dxsvalue.FreeValue(jsovalue)
	return JobInfo{}
}

func createChargeJob(robotId, jobId, groupid, jobCreator string, robotStatus enum.RobotStatusEnum, jsovalue *dxsvalue.DxValue, destJob *BaseJobInfo) {
	destJob.JobCreator = jobCreator
	destJob.Jobgroup = groupid
	destJob.JobId = jobId
	destJob.RobotID = robotId
	destJob.JobType = enum.JtCharge
	destJob.CreateTime.Time = time.Now()
	destJob.Origin = enum.MoDispatch

	jsovalue.Clear()
	c := jsovalue.ValueCache()
	jsovalue.SetKeyCached("JobType", dxsvalue.VT_Int, c).SetInt(int64(enum.JtCharge))
	jsovalue.SetKeyCached("Origin", dxsvalue.VT_Int, c).SetInt(int64(enum.MoDispatch))
	jsovalue.SetKeyCached("RobotStatusToPAD", dxsvalue.VT_Int, c).SetInt(int64(robotStatus))
	jsovalue.SetKeyCached("Order", dxsvalue.VT_Int, c).SetInt(0)
	jsovalue.SetKeyCached("RobotID", dxsvalue.VT_String, c).SetString(robotId)
	jsovalue.SetKeyCached("CreateTime", dxsvalue.VT_DateTime, c).SetTime(destJob.CreateTime.Time)
	jsovalue.SetKeyCached("JobCreator", dxsvalue.VT_String, c).SetString(jobCreator)
	jsovalue.SetKeyCached("JobId", dxsvalue.VT_String, c).SetString(groupid)
	jsovalue.SetKeyCached("Jobgroup", dxsvalue.VT_String, c).SetString(groupid)
}

//插入充电任务并迅速执行,如果指定了是要必须立即执行的，就将充电任务提前执行，同时如果有其他正在执行的任务的，改变任务暂停，先充电
//否则就要等到正在执行完毕之后再执行充电
//否则，就立即执行充电任务
//返回是否需要调度执行充电
func InsertChargeJob(officeId, robotId, jobId, groupid, jobCreator string, robotStatus enum.RobotStatusEnum, immediateCharge bool) (immediatelyrun bool, jbinfo BaseJobInfo, pauseOrDelJobs []BaseJobInfo, err error) {
	if robotId == "" {
		return false, jbinfo, nil, errors.New("必须指定机器人")
	}
	rediskey := strings.Join([]string{"jobs", officeId, robotId}, ":")
	lockid := rediskey + ":lock"
	if err = redis.SimpleLock(lockid); err != nil {
		return false, jbinfo, nil, fmt.Errorf("插入充电任务失败:锁定资源%s失败，redisErr:%w", lockid, err)
	}
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil && !redis.IsRedisNil(err) {
		redis.SimpleUnLock(lockid)
		return false, jbinfo, nil, errors.New("redis检索充电任务失败")
	}
	json := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	pipe := redis.Client.TxPipeline()
	if len(values) == 0 {
		//直接插入
		createChargeJob(robotId, jobId, groupid, jobCreator, robotStatus, json, &jbinfo)
		newbt := make([]byte, 0, 128)
		jbinfo.JobState = enum.JsCalling
		json.SetKeyCached("JobState", dxsvalue.VT_Int, json.ValueCache()).SetInt(int64(enum.JsCalling))
		newbt = json.ToJson(false, dxsvalue.JSE_NoEscape, false, newbt[:0])
		pipe.LPush(context.Background(), rediskey, newbt) //插入充电任务到第一个
		pipe.Del(context.Background(), lockid)
		_, err = pipe.Exec(context.Background())
		dxsvalue.FreeValue(json)
		return true, jbinfo, nil, err
	}

	realFirstJob := make([]byte, 0, 128)
	//先删除掉所有的充电任务
	for i := 0; i < len(values); i++ {
		err = json.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		if err != nil || json.Count() == 0 {
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			continue
		}
		if json.AsInt("JobType", 0) == int(enum.JtCharge) {
			if i == 0 {
				redis.SimpleUnLock(lockid)
				json2JobInfo(json, &jbinfo)
				dxsvalue.FreeValue(json)
				if jbinfo.JobState > enum.JsQueue {
					//本来就是在充电状态
					return false, jbinfo, nil, errors.New("机器人已经在充电状态")
				} else {
					//还未执行充电
					return true, jbinfo, nil, errors.New("已经有充电任务准备执行")
				}
			}
			json2JobInfo(json, &jbinfo)
			jbinfo.JobState = enum.JsNewTask //设定为新任务终止
			pauseOrDelJobs = append(pauseOrDelJobs, jbinfo)
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
		} else if i == 0 {
			if immediateCharge {
				jobState := enum.JobStatusEnum(json.AsInt("JobState", -1))
				if jobState < enum.JsArrived {
					json2JobInfo(json, &jbinfo)
					jbinfo.JobState = enum.JsQueue //重新入队，等待充电执行完毕
					pauseOrDelJobs = append(pauseOrDelJobs, jbinfo)
					json.SetKeyCached("JobState", dxsvalue.VT_Int, json.ValueCache()).SetInt(int64(enum.JsQueue))
					realFirstJob = json.ToJson(false, dxsvalue.JSE_NoEscape, false, realFirstJob[:0])
				} else {
					immediateCharge = false //等待任务执行完毕之后再去执行
				}
			}
		}
	}
	jbinfo.JobId = ""
	jbinfo.Jobgroup = ""
	//删除掉其他的充电的
	pipe.LRem(context.Background(), rediskey, 0, "delete")
	createChargeJob(robotId, jobId, groupid, jobCreator, robotStatus, json, &jbinfo)
	//然后进行新的充电插入
	if immediateCharge && len(realFirstJob) > 0 {
		//是需要立即充电的，将充电任务插入到第一个
		pipe.LPop(context.Background(), rediskey)
		jbinfo.JobState = enum.JsCalling
		json.SetKeyCached("JobState", dxsvalue.VT_Int, json.ValueCache()).SetInt(int64(enum.JsCalling))
		chargeBt := json.ToJson(false, dxsvalue.JSE_NoEscape, false, make([]byte, 0, 128))
		pipe.LPush(context.Background(), rediskey, realFirstJob, chargeBt)
		immediatelyrun = true
	} else {
		immediatelyrun = false
		jbinfo.JobState = enum.JsQueue
		json.SetKeyCached("JobState", dxsvalue.VT_Int, json.ValueCache()).SetInt(int64(enum.JsQueue))
		chargeBt := json.ToJson(false, dxsvalue.JSE_NoEscape, false, make([]byte, 0, 128))
		pipe.LPush(context.Background(), rediskey, chargeBt)
	}
	pipe.Del(context.Background(), lockid)
	_, err = pipe.Exec(context.Background())
	dxsvalue.FreeValue(json)
	return immediatelyrun, jbinfo, pauseOrDelJobs, err
}

//立即执行一个返程任务
func RunningBackJob(officeId string, backJob *JobInfo, forceBack bool) (err error, pauseJob BaseJobInfo) {
	if backJob.RobotID == "" {
		return errors.New("返程任务必须指定一个机器人"), BaseJobInfo{}
	}
	if backJob.JobType != enum.JtBack {
		return errors.New("无效的返程任务"), BaseJobInfo{}
	}
	rediskey := strings.Join([]string{"jobs", officeId, backJob.RobotID}, ":")
	lockid := rediskey + ":lock"
	if err = redis.SimpleLock(lockid); err != nil {
		return fmt.Errorf("执行返程失败：redis锁定机器人任务资源%s失败:%w", lockid, err), BaseJobInfo{}
	}
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil && !redis.IsRedisNil(err) {
		redis.SimpleUnLock(lockid)
		return errors.New("redis检索充电任务失败"), BaseJobInfo{}
	}
	pipe := redis.Client.TxPipeline()
	hasDel := false
	var forceByte []byte
	canRunning := len(values) == 0
	if !canRunning {
		json := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
		for i := 0; i < len(values); i++ {
			err = json.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
			if err != nil {
				pipe.LSet(context.Background(), rediskey, int64(i), "delete")
				hasDel = true
				continue
			}
			if json.AsInt("JobType", 0) == int(enum.JtBack) {
				//正在执行
				backJob.JobId = json.AsString("JobId", backJob.JobId)
				backJob.Jobgroup = json.AsString("Jobgroup", backJob.Jobgroup)
				backJob.JobState = enum.JobStatusEnum(json.AsInt("JobState", 0))
				if backJob.JobState > enum.JsCalling {
					backJob.StartTime.Time = time.Now()
				} else {
					backJob.JobState = enum.JsCalling
				}
				backJob.CreateTime.Time = json.TimeByName("CreateTime", backJob.CreateTime.Time)
				if hasDel {
					pipe.LRem(context.Background(), rediskey, 0, "delete")
					pipe.Del(context.Background(), rediskey, lockid)
					_, err = pipe.Exec(context.Background())
				} else {
					redis.SimpleUnLock(lockid)
				}
				dxsvalue.FreeValue(json)
				return err, BaseJobInfo{}
			}
			//强制返程
			canRunning = forceBack
			if forceBack {
				json2JobInfo(json, &pauseJob)
				if pauseJob.JobState < enum.JsArrived {
					pauseJob.JobState = enum.JsQueue //变为等待状态
					json.SetKeyCached("JobState", dxsvalue.VT_Int, json.ValueCache()).SetInt(int64(pauseJob.JobState))
					forceByte = make([]byte, 0, 128)
					forceByte = json.ToJson(false, dxsvalue.JSE_NoEscape, true, forceByte)
					redis.LSet(context.Background(), rediskey, int64(i), "delete")
					hasDel = true
				} else {
					canRunning = false
				}
			}
			break
		}
	}
	if !canRunning {
		return errors.New("有其他任务执行，不允许执行返程任务"), pauseJob
	}
	if hasDel {
		pipe.LRem(context.Background(), rediskey, 0, "delete")
	}
	if len(forceByte) > 0 {
		pipe.LPush(context.Background(), rediskey, forceByte)
	}
	bt, _ := json.Marshal(backJob)
	pipe.LPush(context.Background(), rediskey, bt)
	pipe.Del(context.Background(), rediskey, lockid)
	_, err = pipe.Exec(context.Background())
	return err, pauseJob
}

//移除机器人的正在缓存中的任务,不移除当前正在执行的任务
func RemoveRobotWaitJob(officeid, robotid string) []BaseJobInfo {
	rediskey := strings.Join([]string{"jobs", officeid, robotid}, ":")
	lockid := rediskey + ":lock"
	if err := redis.SimpleLock(lockid); err != nil {
		log.WithError(err).Errorf("redis分布式锁%s失败,清空任务失败", lockid)
		return nil
	}
	jsonvalue, err := redis.Client.LIndex(context.Background(), rediskey, 0).Result()
	if err != nil {
		if !redis.IsRedisNil(err) {
			log.WithError(err).Error("清空机器人缓存中的任务信息失败")
		}
		return nil
	}
	willClearAll := false
	var runningJob JobInfo
	json.Unmarshal(DxCommonLib.FastString2Byte(jsonvalue), &runningJob)
	if runningJob.JobState < enum.JsCalling || runningJob.JobState > enum.JsCompleted {
		//正在执行中的任务，不管
		willClearAll = true
	}

	pipe := redis.Client.Pipeline()
	//先查询出当前机器人的任务列信息
	lrangeCmder := pipe.LRange(context.Background(), rediskey, 1, -1)
	if willClearAll {
		pipe.LTrim(context.Background(), rediskey, 1, 0) //清空
	} else {
		pipe.LTrim(context.Background(), rediskey, 0, 0) //留下第一个元素
	}
	pipe.Del(context.Background(), lockid)
	_, err = pipe.Exec(context.Background())
	if err != nil {
		if !redis.IsRedisNil(err) {
			log.WithError(err).Error("清空机器人缓存中的任务信息失败")
		}
		return nil
	}
	jsonValues, err := lrangeCmder.Result()
	result := make([]BaseJobInfo, len(jsonValues)+1)
	result[0] = runningJob.BaseJobInfo
	if redis.IsRedisNil(err) {
		return result
	}
	for i := 0; i < len(jsonValues); i++ {
		json.Unmarshal(DxCommonLib.FastString2Byte(jsonValues[i]), &result[i+1])
	}
	return result
}

//清空机器人的任务列表，包括正在执行的和未执行的任务,返回清除的任务列表信息,
func ClearRobotJobs(officeid, robotid string) []BaseJobInfo {
	rediskey := strings.Join([]string{"jobs", officeid, robotid}, ":")
	lockid := rediskey + ":lock"
	if err := redis.SimpleLock(lockid); err != nil {
		log.WithError(err).Errorf("清空机器人任务失败，锁定资源%s失败", lockid)
		return nil
	}

	pipe := redis.Client.TxPipeline()
	//先查询出当前机器人的任务列信息
	lrangeCmder := pipe.LRange(context.Background(), rediskey, 0, -1)
	pipe.LTrim(context.Background(), rediskey, 1, 0) //清空list
	pipe.Del(context.Background(), lockid)
	_, err := pipe.Exec(context.Background())

	if err != nil {
		if !redis.IsRedisNil(err) {
			log.WithError(err).Error("清空机器人任务信息失败")
		}
		return nil
	}
	jsonValues, err := lrangeCmder.Result()
	if redis.IsRedisNil(err) {
		return nil
	}
	result := make([]BaseJobInfo, len(jsonValues))
	for i := 0; i < len(jsonValues); i++ {
		json.Unmarshal(DxCommonLib.FastString2Byte(jsonValues[i]), &result[i])
	}
	return result
}

func json2JobInfo(jsovalue *dxsvalue.DxValue, jbinfo *BaseJobInfo) {
	jbinfo.JobType = enum.JobTypeEnum(jsovalue.AsInt("JobType", 0))
	jbinfo.Back = jsovalue.AsBool("Back", false)
	jbinfo.JobCreator = jsovalue.AsString("JobCreator", "")
	jbinfo.CreateTime = JsonTime{Time: jsovalue.TimeByName("CreateTime", time.Time{})}
	jbinfo.ArrivedTime = JsonTime{Time: jsovalue.TimeByName("ArrivedTime", time.Time{})}
	jbinfo.StartTime = JsonTime{Time: jsovalue.TimeByName("StartTime", time.Time{})}
	jbinfo.Origin = enum.MsgOriginEnum(jsovalue.AsInt("Origin", 0))
	jbinfo.JobState = enum.JobStatusEnum(jsovalue.AsInt("JobState", 0))
	jbinfo.StartPosId = jsovalue.AsString("StartPosId", "")
	jbinfo.StartBuild = jsovalue.AsString("StartBuild", "")
	jbinfo.StartPosName = jsovalue.AsString("StartPosName", "")
	jbinfo.StartFloor = jsovalue.AsInt("StartFloor", 0)
	jbinfo.EndPositionID = jsovalue.AsString("EndPositionID", "")
	jbinfo.EndBuildId = jsovalue.AsString("EndBuildId", "")
	jbinfo.EndPosName = jsovalue.AsString("EndPosName", "")
	jbinfo.Endfloor = jsovalue.AsInt("Endfloor", 0)
	jbinfo.Description = jsovalue.AsString("Description", "")
	jbinfo.RobotID = jsovalue.AsString("RobotID", "")
	jbinfo.JobId = jsovalue.AsString("JobId", "")
	jbinfo.Jobgroup = jsovalue.AsString("Jobgroup", "")
	jbinfo.ReplayTime = JsonTime{Time: jsovalue.TimeByName("StartTime", time.Time{})}
	jbinfo.Remarks = jsovalue.AsString("Remarks", "")
}

//删除某个机构下面的某个任务,并且返回删除的任务信息
//officeId和robotId不能为空
//如果jobid为空，就是清空整个组
func RemoveJob(officeid, jobId, groupId, robotId string) (deljob BaseJobInfo, nextJobGroup []BaseJobInfo, isrunningjob, willBack bool, err error) {
	if robotId == "" || jobId == "" {
		return deljob, nil, false, false, errors.New("必须指定机器人信息和任务信息")
	}
	rediskey := strings.Join([]string{"jobs", officeid, robotId}, ":")
	lockid := rediskey + ":lock"
	if err = redis.SimpleLock(lockid); err != nil {
		log.WithError(err).Errorf("删除任务失败，锁定资源%s失败", lockid)
		return deljob, nil, false, false, errors.New("锁定任务资源列表发生错误")
	}
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil {
		redis.SimpleUnLock(lockid)
		if redis.IsRedisNil(err) {
			err = nil
		}
		return deljob, nil, false, false, err
	}
	jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	maxcount := len(values)
	var nextRunning BaseJobInfo
	for i := 0; i < maxcount; i++ {
		err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		if err != nil {
			continue
		}
		if jsovalue.AsString("JobId", "") == jobId && jsovalue.AsString("RobotID", "") == robotId &&
			(groupId == "" || jsovalue.AsString("Jobgroup", "") == groupId) {
			json2JobInfo(jsovalue, &deljob)
			deljob.EndTime = JsonTime{Time: time.Now()} //结束时间
			//是否是正在执行的任务
			isrunningjob = i == 0 && deljob.JobState >= enum.JsStated && deljob.JobState < enum.JsLowElectric
			willBack = deljob.Back && i == maxcount-1 //是最后一个任务，并且这个任务是需要返程的，才可以返程
			//删除这个,使用pipeline模式
			/*
				MULTI
				incr tx_pipeline_counter
				expire tx_pipeline_counter 3600
				EXEC
			*/
			pipe := redis.Client.TxPipeline()
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			delGroup := ""
			if isrunningjob && i < maxcount-1 {
				//删除了正在执行的，需要继续执行下一个
				isFirst := true
				nextGroup := ""
				for k := i + 1; k < maxcount; k++ {
					err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i+1]), true)
					if err != nil {
						pipe.LSet(context.Background(), rediskey, int64(k), "delete")
						continue
					}
					json2JobInfo(jsovalue, &nextRunning)
					if nextRunning.JobId == "" {
						pipe.LSet(context.Background(), rediskey, int64(k), "delete")
						continue
					}
					if delGroup != "" {
						//要删除这个group
						if nextRunning.Jobgroup == delGroup {
							pipe.LSet(context.Background(), rediskey, int64(k), "delete")
							continue
						}
						delGroup = ""
						isFirst = true
					}
					if isFirst {
						isFirst = false
						nextGroup = nextRunning.Jobgroup
						//设置下一个任务的开始信息
						nextRunning.StartTime = JsonTime{Time: time.Now()} //开始时间
						nextRunning.StartPosId = deljob.EndPositionID
						nextRunning.StartFloor = deljob.Endfloor
						nextRunning.StartBuild = deljob.EndBuild
						nextRunning.StartBuildName = deljob.EndBuild
						nextRunning.StartPosName = deljob.EndPosName

						if nextRunning.Jobgroup == deljob.Jobgroup {
							nextRunning.JobState = enum.JsStated //同一组任务才会设定为自动开始运行，如果下一组任务换组了，需要等待调度推送到机器人之后才能继续
						} else {
							//不是同一组的，需要判定一下，这个组的任务是否已经超时了
							if time.Now().Sub(nextRunning.CreateTime.Time) >= time.Hour*3 { //这一组任务，超过2小时没有执行了，那么任务组超时
								log.Debugf("任务组%s的任务过期，删除", nextGroup)
								pipe.LSet(context.Background(), rediskey, int64(k), "delete")
								delGroup = nextGroup
								continue
							}
						}
						bt, _ := json.Marshal(nextRunning)
						//要重新构建一下
						pipe.LSet(context.Background(), rediskey, int64(k), bt)

					} else if nextRunning.Jobgroup != nextGroup { //换了组了
						break
					}
					nextJobGroup = append(nextJobGroup, nextRunning)
				}
			}
			pipe.LRem(context.Background(), rediskey, 0, "delete")
			pipe.Del(context.Background(), lockid) //删除锁
			_, err = pipe.Exec(context.Background())
			dxsvalue.FreeValue(jsovalue)
			//log.Debug("移除正在执行的任务 ,err=",err)
			return deljob, nextJobGroup, isrunningjob, willBack, err
		}
	}
	redis.SimpleUnLock(lockid)
	dxsvalue.FreeValue(jsovalue)
	deljob.JobId = jobId
	deljob.Jobgroup = groupId
	return deljob, nil, false, false, fmt.Errorf("未发现指定的任务组:%s下的任务ID=%s的相关任务", groupId, jobId)
}

func RemoveJobInfo(officeId string, jobinf *JobInfo) (deljob BaseJobInfo, nextRungroup []BaseJobInfo, isrunningjob, willBack bool, err error) {
	if jobinf == nil || officeId == "" || jobinf.JobId == "" || jobinf.Jobgroup == "" || jobinf.RobotID == "" {
		return deljob, nil, false, false, errors.New("未指定机器人信息以及任务信息")
	}
	return RemoveJob(officeId, jobinf.JobId, jobinf.Jobgroup, jobinf.RobotID)
}

func RemoveJobGroup(officeId, groupId, robotId string, robotPos *RobotPosInfo) (currentGroupJobs, nextJobGroup []BaseJobInfo, err error) {
	if robotId == "" || groupId == "" {
		return nil, nil, errors.New("必须指定机器人信息和任务组信息")
	}
	rediskey := strings.Join([]string{"jobs", officeId, robotId}, ":")
	lockid := rediskey + ":lock"
	if err = redis.SimpleLock(lockid); err != nil {
		log.WithError(err).Errorf("删除任务组失败，锁定资源%s失败", lockid)
		return nil, nil, errors.New("锁定任务资源列表发生错误")
	}
	values, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil {
		redis.SimpleUnLock(lockid)
		if redis.IsRedisNil(err) {
			err = nil
		}
		return nil, nil, err
	}
	jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	maxcount := len(values)
	pipe := redis.Client.Pipeline()
	var nextRunning BaseJobInfo
	hasDel := false
	currentGroupJobs = make([]BaseJobInfo, 0, 8)
	willReturnNextGroup := robotPos != nil
	isfirst := true
	for i := 0; i < maxcount; i++ {
		err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[i]), true)
		if err != nil {
			hasDel = true
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			continue
		}
		json2JobInfo(jsovalue, &nextRunning)
		isSameGroup := nextRunning.Jobgroup == groupId
		if isSameGroup {
			currentGroupJobs = append(currentGroupJobs, nextRunning)
		}
		if nextRunning.JobId == "" || isSameGroup {
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			hasDel = true
			continue
		}
		if willReturnNextGroup { //如果是删除的中间的，不返回
			if isfirst {
				isfirst = false
				nextRunning.StartBuildName = robotPos.BuildName
				nextRunning.StartBuild = robotPos.BuildId
				nextRunning.StartFloor = robotPos.Floor
				nextRunning.StartPosName = robotPos.PosName
				nextRunning.StartPosId = robotPos.PosId
				nextRunning.StartTime.Time = time.Now()
			}
			nextJobGroup = append(nextJobGroup, nextRunning)
			nextGroup := nextRunning.Jobgroup
			for k := i + 1; k < maxcount; k++ {
				err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(values[k]), true)
				if err != nil {
					hasDel = true
					pipe.LSet(context.Background(), rediskey, int64(k), "delete")
					continue
				}
				json2JobInfo(jsovalue, &nextRunning)
				if nextGroup != nextRunning.Jobgroup {
					break
				}
				nextJobGroup = append(nextJobGroup, nextRunning)
			}
			break
		}
		willReturnNextGroup = false
	}
	dxsvalue.FreeValue(jsovalue)
	if hasDel {
		pipe.LRem(context.Background(), rediskey, 0, "delete")
	}
	pipe.Del(context.Background(), lockid) //删除锁
	_, err = pipe.Exec(context.Background())
	if err != nil {
		err = fmt.Errorf("删除机器人%s的任务组%s失败：%w", robotId, groupId, err)
	}
	return currentGroupJobs, nextJobGroup, err
}

//是否是一个正在执行的任务组，如果已经下发过到机器人那边
//返回结果,-2表示发生错误，-1，表示没有任务，0表示有任务，但是任务是在正在执行的任务组中，需要提交给机器人去删除，不直接删除，否则表示直接从缓存删除任务不用提交给机器人
func DeleteIfJobGroupUnRunning(officeId, groupId, jobId, robotId string) (int, error) {
	if officeId == "" || groupId == "" || robotId == "" {
		return -2, nil
	}
	rediskey := strings.Join([]string{"jobs", officeId, robotId}, ":")
	lockid := rediskey + ":lock"
	if err := redis.SimpleLock(lockid); err != nil {
		log.WithError(err).Errorf("DeleteIfJobGroupUnRunning锁定资源%s失败", lockid)
		return -2, errors.New("锁定任务资源列表发生错误")
	}
	jsonValues, err := redis.LRange(context.Background(), rediskey, 0, -1).Result()
	if err != nil {
		if redis.IsRedisNil(err) {
			err = nil
		}
		redis.SimpleUnLock(lockid)
		return -2, err
	}
	jsovalue := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	defer dxsvalue.FreeValue(jsovalue)
	var pipe redis2.Pipeliner
	var jobinfo BaseJobInfo
	result := 0
	runningGroup := ""
	findJobIndexs := make([]int, 0, 10) //查找到的任务索引
	//删除其中之一的任务
	for i := 0; i < len(jsonValues); i++ {
		err = jsovalue.LoadFromJson(DxCommonLib.FastString2Byte(jsonValues[i]), true)
		if err != nil {
			if pipe == nil {
				pipe = redis.Client.TxPipeline()
			}
			pipe.LSet(context.Background(), rediskey, int64(i), "delete")
			continue
		}
		json2JobInfo(jsovalue, &jobinfo)
		if jobinfo.JobState > enum.JsQueue {
			//正在执行的组
			runningGroup = jobinfo.Jobgroup
		}
		//查找到删除的任务索引
		if jobinfo.Jobgroup == groupId && (jobId == "" || jobId == jobinfo.JobId) {
			findJobIndexs = append(findJobIndexs, i)
		}
	}
	if len(findJobIndexs) > 0 {
		if runningGroup != groupId {
			//是没有下发的任务组中的任务，直接删除掉
			result = 1
			if pipe == nil {
				pipe = redis.Client.TxPipeline()
			}
			for _, idx := range findJobIndexs {
				pipe.LSet(context.Background(), rediskey, int64(idx), "delete")
			}
		}
	} else {
		//未发现任务
		result = -1
	}

	if pipe != nil {
		pipe.LRem(context.Background(), rediskey, 0, "delete")
		pipe.Del(context.Background(), lockid) //删除锁
		_, err = pipe.Exec(context.Background())
		return result, err
	}
	redis.SimpleUnLock(lockid)
	return result, nil
}
