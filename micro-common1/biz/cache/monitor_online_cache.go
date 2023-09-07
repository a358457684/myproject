package cache

//const (
//	monitorOnlineKey     = "monitor_online:%s:%s"
//	monitorOnlineTimeOUt = time.Minute
//)
//
////刷新机器人最后通讯时间
//func SaveMonitorOnline(officeId, robotId string) error {
//	return redis.Set(context.TODO(), getMonitorOnlineKey(officeId, robotId), time.Now().Unix(), monitorOnlineTimeOUt).Err()
//}
//
////机器人是否在线
//func HasMonitorOnline(officeId, robotId string) (bool, error) {
//	cmd := redis.Exists(context.TODO(), getMonitorOnlineKey(officeId, robotId))
//	return cmd.Val() > 0, cmd.Err()
//}
//
////获取所有在线机器人id
//func FindMonitorOnlineAll() (result []OfficeRobotKey, err error) {
//	return findRobotOnline("*", "*")
//}
//
////获取所有在线机器人id
//func FindMonitorOnlineByOfficeId(officeId string) (result []OfficeRobotKey, err error) {
//	return findRobotOnline(officeId, "*")
//}
//
//func findRobotOnline(officeId, robotId string) (result []OfficeRobotKey, err error) {
//	cmd := redis.Keys(context.TODO(), getMonitorOnlineKey(officeId, robotId))
//	err = cmd.Err()
//	for _, key := range cmd.Val() {
//		result = append(result, OfficeRobotKey{Data: key})
//	}
//	return
//}
//
//func getMonitorOnlineKey(officeId, robotId string) string {
//	return fmt.Sprintf(monitorOnlineKey, officeId, robotId)
//}
