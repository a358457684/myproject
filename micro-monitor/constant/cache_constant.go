package constant

const (

	// 调度相关
	DispatchRobotJobJobs = "dispatch_robot_job_jobs"
	CurrentSentJobKey    = "current_sent"
	LastSentJobKey       = "last_sent"
	NeedLiftStatusRobots = "need_lift_status_robots"

	// 校验监控系统是否要存入新的状态到ES
	LastRobotStatusName = "last_robot_status"

	// 机器人是否在线、在线时长
	LogisticsRobotOnline     = "logistics_robot_online:"
	LogisticsRobotOnlineTime = "logistics_robot_onlinetime:"

	// 监控系统预警相关
	MonitorJobConfigKey        = "monitor_job_config_key"
	MonitorElectricConfigKey   = "monitor_electric_config_key"
	MonitorNetConnectStatusKey = "monitor_net_connect_status_key"
	MonitorStatusConfigKey     = "monitor_status_config_key"
	MonitorScopeConfigKey      = "monitor_scope_point_config_key"

	// 工作配置推送状态
	RobotWorkPushStatusInfo = "robot_work_push_status_info"

	// 机器人信息变化相关
	LogisticsRobotsInfo           = "logistics_robots_info"
	LogisticsRobotsStatusUpload   = "logistics_robots_status_upload"
	LogisticsRobotsPositionUpload = "logistics_robots_position_upload"

	// 分布式锁
	MonitorStatusLock      = "monitor_status_lock:"
	MonitorPushMessageLock = "monitor_push_message_lock:"
	MonitorOnlineLock      = "monitor_online_lock"

	// 推送消息相关
	PushMessageCallback = "push_message_callback:"
	PushMessageRecord   = "push_message_record:"

	// websocket相关
	MonitorMap = "monitor_map"

	// 代理服务
	ProxyStatus = "proxy_status"
)
