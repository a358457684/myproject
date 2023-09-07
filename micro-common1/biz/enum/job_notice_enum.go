package enum

//任务变更通知
type JobNoticeEnum string

const (
	JnStart       JobNoticeEnum = "jobStart"           //任务开始
	JnRefresh     JobNoticeEnum = "refreshJobRank"     //任务队列刷新
	JnArrived     JobNoticeEnum = "jobArrive"          //任务到达目的地
	JnFinished    JobNoticeEnum = "jobEnd"             //任务结束
	JnCancel      JobNoticeEnum = "jobCancel"          //任务被取消
	JnFailed      JobNoticeEnum = "jobFailed"          //任务失败
	JnStatus      JobNoticeEnum = "robotStatus"        //机器人状态刷新
	JnRfidArticle JobNoticeEnum = "refreshRfidArticle" //RFID物品刷新
)
