package service

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/model"
	"epshealth-airobot-monitor/result"
	"epshealth-airobot-monitor/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"micro-common1/biz/dto"
	"micro-common1/biz/enum"
	"micro-common1/elasticsearch"
	"micro-common1/redis"
	"micro-common1/util"
	"strings"
	"time"
)

const (
	// ES的索引
	robotStatusIndex          = "robot_status"
	sourceRobotStatusIndex    = "source_robot_status"
	RobotPushMessageIndex     = "robot_push_message"
	robotJobStatusChangeIndex = "robot_job_status_change"
	// ES的索引日期格式
	DataFormat = "2006.01"
)

type ElasticRobotStatusPageQuery struct {
	model.BasePageQuery
	Status    enum.RobotStatusEnum `json:"status"`    // 机器人状态
	NetStatus enum.NetStatusEnum   `json:"netStatus"` // 网络状态
}

type ElasticRobotPushMessagePageQuery struct {
	model.BasePageQuery
	Path      string              `json:"path"`      // 消息路径
	Status    constant.PushStatus `json:"status"`    // 消息推送是否成功  1:推送成功  2:执行成功
	SendCount int                 `json:"sendCount"` // 发送次数
}

// 获取机器人状态
func FindPageRobotStatus(c *gin.Context, vo ElasticRobotStatusPageQuery) model.PageResult {
	if vo.StartDate.IsZero() || vo.EndDate.IsZero() {
		t := time.Now()
		vo.StartDate = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		vo.EndDate = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
	}
	// 获取索引
	indices := getIndex(vo.StartDate, vo.EndDate, robotStatusIndex)
	dataList := make([]model.ElasticRobotStatus, 0)
	var total int64
	if len(indices) > 0 {
		query := getRobotStatusDsl(vo)
		hits := elasticsearch.PageByBody(c, query, vo.PageIndex, vo.PageSize, "lastUploadTime:desc", indices...)
		total = hits.Total.Value
		for _, source := range hits.Hits {
			jsonData, _ := json.Marshal(source.Source)
			var data model.ElasticRobotStatus
			_ = json.Unmarshal(jsonData, &data)
			dataList = append(dataList, data)
		}
	}
	return model.PageResult{
		PageIndex: vo.PageIndex,
		PageSize:  vo.PageSize,
		Total:     total,
		Data:      dataList,
	}
}

func getRobotStatusDsl(vo ElasticRobotStatusPageQuery) string {
	dsl := `{"query": {"bool": {"must": [%s]}}}`

	var mustQuery []string
	mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"officeId": "%s"}}`, vo.OfficeId))
	mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"robotId": "%s"}}`, vo.RobotId))

	status := vo.Status
	if status != 0 {
		if status == enum.RsStop {
			// eStopStatus： 是否急停  0-正常，1-急停状态
			mustQuery = append(mustQuery, fmt.Sprintf(
				`{"bool": {"should":[{"term": {"status": %d}},{"term": {"eStopStatus": 1}}]}}`, status))
		} else {
			mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"status": %d}}`, status))
		}
	}

	if vo.NetStatus == enum.NsOnline || vo.NetStatus == enum.NsOffline {
		mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"netStatus": %d}}`, vo.NetStatus))
	}

	mustQuery = append(mustQuery, fmt.Sprintf(
		`{"range":{"lastUploadTime":{"gte":"%s","lte":"%s","format":"yyyy-MM-dd HH:mm:ss","time_zone":"+08:00"}}}`,
		vo.StartDate.Format(constant.DateTimeFormat), vo.EndDate.Format(constant.DateTimeFormat)))

	return fmt.Sprintf(dsl, strings.Join(mustQuery, ","))
}

// 获取机器人推送信息
func FindPageRobotPushMessage(c *gin.Context, vo ElasticRobotPushMessagePageQuery) model.PageResult {
	if vo.StartDate.IsZero() || vo.EndDate.IsZero() {
		t := time.Now()
		vo.StartDate = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		vo.EndDate = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
	}
	// 获取索引
	indices := getIndex(vo.StartDate, vo.EndDate, RobotPushMessageIndex)
	dataList := make([]model.ElasticRobotPushMessage, 0)
	var total int64
	if len(indices) > 0 {
		query := getRobotPushMessageDsl(vo)
		hits := elasticsearch.PageByBody(c, query, vo.PageIndex, vo.PageSize, "timestamp:desc", indices...)
		total = hits.Total.Value
		for _, source := range hits.Hits {
			jsonData, _ := json.Marshal(source.Source)
			var data model.ElasticRobotPushMessage
			_ = json.Unmarshal(jsonData, &data)
			dataList = append(dataList, data)
		}
	}
	return model.PageResult{
		PageIndex: vo.PageIndex,
		PageSize:  vo.PageSize,
		Total:     total,
		Data:      dataList,
	}
}

func getRobotPushMessageDsl(vo ElasticRobotPushMessagePageQuery) string {
	dsl := `{"query": {"bool": {"must": [%s]}}}`

	var mustQuery []string
	mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"officeId": "%s"}}`, vo.OfficeId))
	mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"robotId": "%s"}}`, vo.RobotId))

	if vo.Path != "" {
		mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"path.keyword": "%s"}}`, vo.Path))
	}

	if vo.SendCount != 0 {
		mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"sendCount": %d}}`, vo.SendCount))
	}

	if vo.Status != 0 {
		mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"status": %d}}`, vo.Status))
	}

	mustQuery = append(mustQuery, fmt.Sprintf(
		`{"range":{"timestamp":{"gte":"%s","lte":"%s","format":"yyyy-MM-dd HH:mm:ss","time_zone":"+08:00"}}}`,
		vo.StartDate.Format(constant.DateTimeFormat), vo.EndDate.Format(constant.DateTimeFormat)))

	return fmt.Sprintf(dsl, strings.Join(mustQuery, ","))
}

// 添加机器人状态文档
func AddRobotStatusDocument(robotVo dto.RobotStatus, sourceData interface{}) {
	index := fmt.Sprintf("%s-%s", robotStatusIndex, time.Now().Format(DataFormat))
	elasticRobotStatus := toElasticRobotStatus(robotVo)
	err := elasticsearch.CreateDocument(index, elasticRobotStatus.DocumentId, elasticRobotStatus)
	if err == nil {
		jsonData, _ := json.Marshal(sourceData)
		AddSourceRobotStatusDocument(model.ElasticSourceRobotStatus{
			DocumentId:     elasticRobotStatus.DocumentId,
			SourceMsg:      string(jsonData),
			RobotId:        elasticRobotStatus.RobotId,
			OfficeId:       elasticRobotStatus.OfficeId,
			Status:         elasticRobotStatus.StatusText,
			LastUploadTime: elasticRobotStatus.LastUploadTime,
		})
		ctx := context.Background()
		redisKey := fmt.Sprintf("%s:%s", constant.LastRobotStatusName, robotVo.OfficeId)
		_ = redis.HSetJson(ctx, redisKey, robotVo.RobotId, robotVo)
		socketData, _ := json.Marshal(elasticRobotStatus)
		redis.Publish(ctx, constant.WebsocketQueues[1], socketData)
	}
}

func toElasticRobotStatus(robot dto.RobotStatus) model.ElasticRobotStatus {
	return model.ElasticRobotStatus{
		DocumentId:     util.CreateUUID(),
		RobotId:        robot.RobotId,
		OfficeId:       robot.OfficeId,
		RobotModel:     robot.RobotModel,
		BuildingName:   robot.BuildingName,
		BuildingId:     robot.BuildingId,
		Status:         robot.RobotStatus,
		StatusText:     robot.RobotStatus.Description(),
		JobId:          robot.JobId,
		LastUploadTime: robot.Time,
		X:              robot.X,
		Y:              robot.Y,
		SpotId:         robot.LastPositionId,
		SpotName:       robot.LastPositionName,
		Target:         robot.TargetPositionId,
		TargetName:     robot.TargetPositionName,
		NextSpot:       robot.NextPositionId,
		Floor:          robot.Floor,
		Electric:       robot.Electric,
		NetStatus:      robot.NetStatus,
		NetStatusText:  robot.NetStatus.Description(),
		PauseType:      robot.PauseType,
		EStopStatus:    robot.EstopStatus,
	}
}

// 添加机器人推送信息文档
func AddRobotPushMessageDocument(message model.ElasticRobotPushMessage, index string) {
	_ = elasticsearch.CreateDocument(index, message.DocumentId, message)
}

// ES查询的时候，处理 jobId = queryJobId || (dispatchMode = 1 && finalJobId = queryJobId)
func FindRobotJobExecRecordList(c *gin.Context, query model.RobotJobStatusChangeQuery) (robotStatuses []model.ElasticRobotJobExec) {
	dsl := getRobotJobExecRecordDsl(query)
	index := getRobotJobExecRecordIndex(query.Day)
	hits := elasticsearch.PageByBody(c, dsl, 1, 100, "lastUploadTime:asc", index)
	for _, data := range hits.Hits {
		jsonData, _ := json.Marshal(data.Source)
		var status model.ElasticRobotJobExec
		_ = json.Unmarshal(jsonData, &status)
		handlerInfo(&status)
		robotStatuses = append(robotStatuses, status)
	}
	return
}

// 处理描述信息
func handlerInfo(status *model.ElasticRobotJobExec) {
	// 设置状态描述
	status.StatusText = status.Status.Description()
	// 如果为配送且机器人状态为空闲中,根据接收状态修改状态描述信息
	if status.AcceptState == enum.ASCompleted {
		status.StatusText = enum.ASCompleted.String()
	}
	if status.AcceptState == enum.ASTimeout {
		status.StatusText = enum.ASTimeout.String()
	}
	if status.ExecState != 0 {
		status.ExecStateText = status.ExecState.String()
	}
	// 计算时长
	if status.TimeConsume == 0 {
		status.TimeConsume = utils.GetTimeDif(status.StatusStartTime, status.StatusEndTime)
	}
	// 急停加暂停 1是 0否
	if status.EstopStatus == 1 && status.PauseType == 1 {
		status.StopInfo = fmt.Sprintf("%s/%s", enum.RsStop.Description(), enum.RsPause.Description())
		// 急停
	} else if status.EstopStatus == 1 {
		status.StopInfo = enum.RsStop.Description()
		// 暂停
	} else if status.PauseType == 1 {
		status.StopInfo = enum.RsPause.Description()
	}
	if status.Status == enum.RsFailed {
		status.StopInfo = enum.EsFailed.String()
	}
	if status.StopInfo == "" {
		status.StopInfo = enum.EsNormal.String()
	}
}

func getRobotJobExecRecordDsl(vo model.RobotJobStatusChangeQuery) string {
	dsl := `{"query": {"bool": {"must": [%s]}}}`
	var mustQuery []string
	mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"officeId": "%s"}}`, vo.OfficeId))
	if vo.RobotId != "" {
		mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"robotId": "%s"}}`, vo.RobotId))
	}
	mustQuery = append(mustQuery,
		fmt.Sprintf(`{"bool":{"should":[{"term":{"jobId":"%s"}},{"bool":{"must":[{"term":{"dispatchMode":{"value":"1"}}},{"term":{"finalJobId":{"value":"%s"}}}]}}]}}`, vo.JobId, vo.JobId))
	return fmt.Sprintf(dsl, strings.Join(mustQuery, ","))
}

func getRobotJobExecRecordIndex(indexDate time.Time) string {
	if indexDate.IsZero() {
		return fmt.Sprintf("%s-%s", robotJobStatusChangeIndex, time.Now().Format(DataFormat))
	}
	return fmt.Sprintf("%s-%s", robotJobStatusChangeIndex, indexDate.Format(DataFormat))
}

// 查找存入的源数据
func GetSourceRobotStatus(c *gin.Context) {
	index := fmt.Sprintf("%s-*", sourceRobotStatusIndex)
	documentId := c.Param("documentId")
	query := "documentId:" + documentId
	hits := elasticsearch.PageByQuery(c, query, 1, 1, "", index)
	if len(hits.Hits) > 0 {
		result.Success(c, hits.Hits[0].Source)
		return
	}
	result.Fail(c, "没有该文档的源数据")
}

// 添加机器人状态原始数据
func AddSourceRobotStatusDocument(message model.ElasticSourceRobotStatus) {
	index := fmt.Sprintf("%s-%s", sourceRobotStatusIndex, time.Now().Format(DataFormat))
	_ = elasticsearch.CreateDocument(index, message.DocumentId, message)
}

// 获取索引
func getIndex(startDate, endDate time.Time, indexName string) []string {
	now := time.Now()
	indexList := make([]string, 0)
	localStartData, _ := time.ParseInLocation(constant.DateTimeFormat, startDate.Format(constant.DateTimeFormat), time.Local)
	if localStartData.After(now) {
		return indexList
	}
	monthIndex := now.Format(DataFormat)
	lastMonthIndex := now.AddDate(0, -1, 0).Format(DataFormat)
	startIndex := startDate.Format(DataFormat)
	endIndex := endDate.Format(DataFormat)
	if startIndex != monthIndex && endIndex >= lastMonthIndex {
		index := fmt.Sprintf("%s-%s", indexName, lastMonthIndex)
		indexList = append(indexList, index)
	}
	if endIndex >= monthIndex {
		index := fmt.Sprintf("%s-%s", indexName, monthIndex)
		indexList = append(indexList, index)
	}
	return indexList
}

func UpdatePushMessageSendCount(documentId, index string, status constant.PushStatus, sendCount int, now time.Time) error {
	dsl := fmt.Sprintf(`{"doc":{"sendCount":%d,"timestamp":"%s","status":%d,"statusText":"%s"}}`,
		sendCount, now.Format(time.RFC3339Nano), status, status.String())
	return elasticsearch.UpdateDocument(index, documentId, dsl)
}
