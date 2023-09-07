package enum

//消息来源
type MsgOriginEnum int8

const (
	MoUndefind      MsgOriginEnum = iota //未定义
	MoAPP                                //APP
	MoPhone                              //手机客户端
	MoPC                                 //PC端
	MoApplets                            //小程序
	MoPAD                                //PAD
	MoDemoModelPad                       //演示模式PAD
	MoDispatch                           //调度系统
	MoSimulator                          //模拟器
	MoServerChecker                      //定时任务检测
	MoAdminSystem                        //后台管理系统
	MoMonitor                            //监控系统
	MoMonitorSystemChecker               //监控系统定时任务检测
	MoHospitalSystem                     //医院系统
	MoQrCode                             //扫码
	MoWarn                               //预警系统
)

func (m MsgOriginEnum) Code() int {
	return int(m)
}

func (m MsgOriginEnum) String() string {
	switch m {
	case MoAPP:
		return "APP"
	case MoPhone:
		return "手机客户端"
	case MoPC:
		return "PC端"
	case MoApplets:
		return "小程序"
	case MoPAD:
		return "PAD"
	case MoDemoModelPad:
		return "演示模式PAD"
	case MoDispatch:
		return "调度系统"
	case MoSimulator:
		return "模拟器"
	case MoServerChecker:
		return "定时任务检测"
	case MoAdminSystem:
		return "后台管理系统"
	case MoMonitor:
		return "监控系统"
	case MoMonitorSystemChecker:
		return "监控系统定时任务检测"
	case MoHospitalSystem:
		return "医院系统"
	case MoQrCode:
		return "扫码"
	case MoWarn:
		return "预警系统"
	default:
		return "未定义"
	}
}
