package cache

import (
	"common/biz/enum"
	"common/log"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"
)

func TestMake(t *testing.T) {
	str := "http://minio-dev-api.epshealth.com:7070//robotLog/goe201/2021-04-06-14-58-32.log?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=minio%2F20210406%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20210406T065832Z&X-Amz-Expires=36000&X-Amz-SignedHeaders=host&X-Amz-Signature=cf2037d0deea476c0b4b62c086646d8e6a1c9268e2ab737bd18b383ed8813da5"
	log.Info(url.PathEscape(str))

}

func TestJobInfo_MarshalJSON(t *testing.T) {
	jbinfo := JobInfo{
		BaseJobInfo: BaseJobInfo{
			JobType:  enum.JtCall,
			JobState: enum.JsStated,
			//OfficeID: "OfficeID",
			JobCreator: "调度系统",
			RobotID:    "",

			EndPositionID: "EndPositionID",
			EndBuildId:    "EndBuildId",
			Endfloor:      9,
			Description:   "任务描述",
			JobId:         "job01231231230",
			Jobgroup:      "groupID234234",

			ReplayTime: JsonTime{time.Now()},
			StartTime:  JsonTime{time.Now()},
			EndTime:    JsonTime{time.Now()},

			CreateTime: JsonTime{time.Now()},
		},
		RobotStatus: enum.RsWaitForDelivery,
	}
	b, err := json.Marshal(&jbinfo)
	if err != nil {
		fmt.Println("发生错误", err)
		return
	}
	fmt.Println(string(b))

	json.Unmarshal(b, &jbinfo)
	fmt.Println(jbinfo)
}

func TestPushJobInfo(t *testing.T) {
	jbinfo := JobInfo{
		BaseJobInfo: BaseJobInfo{
			JobType:    enum.JtCall,
			JobState:   enum.JsStated,
			JobCreator: "调度系统",
			RobotID:    "001",

			EndPositionID: "EndPositionID",
			EndBuildId:    "EndBuildId",
			Endfloor:      9,
			Description:   "任务描述",
			JobId:         "job001",
			Jobgroup:      "groupID001",

			ReplayTime: JsonTime{time.Now()},
			StartTime:  JsonTime{time.Now()},
			EndTime:    JsonTime{time.Now()},

			CreateTime: JsonTime{time.Now()},
		},
		RobotStatus: enum.RsWaitForDelivery,
	}
	err := SaveJobInfo("testOffice", jbinfo.BaseJobInfo, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	jbinfo.RobotID = "001"
	jbinfo.JobId = "job002"
	jbinfo.Jobgroup = "groupID002"
	SaveJobInfo("testOffice", jbinfo.BaseJobInfo, true)

	jobs, err := GetRobotJobInfo("testOffice", "001")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(jobs[0])
	}
	RemoveJob("testOffice", jbinfo.JobId, jbinfo.Jobgroup, jbinfo.RobotID)
}
