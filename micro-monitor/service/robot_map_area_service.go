package service

import (
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/model"
	"micro-common1/biz/dto"
	"time"
)

func SaveAreaJobRelation(s dto.RobotStatus, finalJobId string, areaId string) string {
	areaJobRelation := model.AreaJobRelation{
		OfficeId:      s.OfficeId,
		StartPosition: s.LastPositionId,
		EndPosition:   s.TargetPositionId,
		FinalJobId:    finalJobId,
		JobId:         s.JobId,
		StartTime:     time.Now(),
		AreaId:        areaId,
	}
	dao.SaveAreaJobRelation(areaJobRelation)
	return areaJobRelation.Id
}

func SaveRobotJobArea(s dto.RobotStatus, finalJobId string, areaId string, finalEndSpotId string) {
	areaJobRelationId := SaveAreaJobRelation(s, finalJobId, areaId)
	robotJobArea := model.RobotJobArea{
		OfficeId:      s.OfficeId,
		BuildingId:    s.BuildingId,
		Floor:         s.Floor,
		StartPosition: s.LastPositionId,
		EndPosition:   finalEndSpotId,
		FinalJobId:    finalJobId,
		JobId:         s.JobId,
		RobotModel:    string(s.RobotModel),
		// 设置最后关联的关系Id
		AreaJobId: areaJobRelationId,
	}
	dao.SaveRobotJobArea(robotJobArea)
}
