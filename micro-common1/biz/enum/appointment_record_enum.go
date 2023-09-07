package enum

type AppointmentRecordEnum int8

const (
	ARCreate AppointmentRecordEnum = iota + 1
	ARStart
	ARCancel
	ARException
)

func (ad AppointmentRecordEnum) String() string {
	switch ad {
	case ARCreate:
		return "创建预约"
	case ARStart:
		return "启动配送"
	case ARCancel:
		return "取消配送"
	case ARException:
		return "异常取消配送"
	}
	return "未知"
}
