package enum

type LogTypeEnum int

const (
	LtApp     LogTypeEnum = iota //APP日志
	LtIPC                        //工控机日志
	LtChassis                    //底盘日志
)
