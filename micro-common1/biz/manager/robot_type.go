package manager

import (
	"common/biz/enum"
	"common/biz/mq/sys_res_mq"
	"common/log"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
	"sync"
	"unsafe"
)

//底盘类型信息
var (
	chassisArray []Chassis
	chassisMutex sync.RWMutex
)

func init() {
	chassisArray = make([]Chassis, 0, 5)
}

//注册一个底盘
func RegisterChassis(chassisModel Chassis) {
	chassisMutex.Lock()
	for i := 0; i < len(chassisArray); i++ {
		if chassisArray[i].Name == chassisModel.Name {
			chassisArray[i] = chassisModel
			chassisMutex.Unlock()
			return
		}
	}
	chassisArray = append(chassisArray, chassisModel)
	chassisMutex.Unlock()
}

//反注册
func UnRegisterRobotChassis(chassisModel string) {
	chassisMutex.Lock()
	l := len(chassisArray)
	for i := 0; i < l; i++ {
		if chassisArray[i].Name == chassisModel {
			copy(chassisArray[i:l-1], chassisArray[i+1:l])
			chassisArray = chassisArray[:l-1]
			break
		}
	}
	chassisMutex.Unlock()
}

func RobotChassisByModel(chassisModel string) *Chassis {
	chassisMutex.Lock()
	for i := 0; i < len(chassisArray); i++ {
		if chassisArray[i].Name == chassisModel {
			chassisMutex.Unlock()
			return &chassisArray[i]
		}
	}
	chassisMutex.Unlock()
	return nil
}

func BindRobotType(chassisModel, tp, tpName string, typeFunction RobotFunction) {
	//绑定一个机器人类型到本底盘上
	chassisMutex.Lock()
	for i := 0; i < len(chassisArray); i++ {
		if chassisArray[i].Name == chassisModel {
			for j := 0; j < len(chassisArray[i].robotTypes); j++ {
				if chassisArray[i].robotTypes[j].Name == tp {
					chassisArray[i].robotTypes[j].Function = typeFunction
					chassisArray[i].robotTypes[j].Code = tpName
					chassisMutex.Unlock()
					return
				}
			}
			chassisArray[i].robotTypes = append(chassisArray[i].robotTypes, RobotTypeStruct{
				Name:     tp,
				Code:     tpName,
				Chassis:  &chassisArray[i],
				Function: typeFunction,
			})
		}
	}
	chassisMutex.Unlock()
}

func UnBindRobotType(chassisModel, robotType string) {
	chassisMutex.Lock()
	for i := 0; i < len(chassisArray); i++ {
		if chassisArray[i].Name != chassisModel {
			continue
		}
		for j := 0; j < len(chassisArray[i].robotTypes); j++ {
			if chassisArray[i].robotTypes[j].Name == robotType || chassisArray[i].robotTypes[j].Code == robotType {
				chassisArray[i].robotTypes = append(chassisArray[i].robotTypes[:j], chassisArray[i].robotTypes[j+1:]...)
				chassisMutex.Unlock()
				return
			}
		}
	}
	chassisMutex.Unlock()
}

//底盘型号
type Chassis struct {
	Supplier           string            //供应商
	Name               string            //底盘类型
	DefaultResolution  float64           //默认分辨率（米/像素）
	DefaultRealOriginX float64           //默认地图图片的左下角在真实物里环境下的坐标信息，主要用来作为计算相对于坐标原点的地图偏移
	DefaultRealOriginY float64           //默认地图图片的左下角在真实物里环境下的坐标信息，主要用来作为计算相对于坐标原点的地图偏移
	robotTypes         []RobotTypeStruct //机器人类型
}

func (chassis *Chassis) BindRobotType(tp, tpName string, typeFunction RobotFunction) {
	//绑定一个机器人类型到本底盘上
	chassisMutex.Lock()
	for i := 0; i < len(chassis.robotTypes); i++ {
		if chassis.robotTypes[i].Name == tp {
			chassis.robotTypes[i].Function = typeFunction
			chassis.robotTypes[i].Code = tpName
			chassisMutex.Unlock()
			return
		}
	}
	chassis.robotTypes = append(chassis.robotTypes, RobotTypeStruct{
		Name:     tp,
		Code:     tpName,
		Chassis:  chassis,
		Function: typeFunction,
	})
	chassisMutex.Unlock()
}

type RobotType string

//获取机器人类型的真实数据
func (t RobotType) Type() *RobotTypeStruct {
	var result *RobotTypeStruct
	chassisMutex.RLock()
	for i := 0; i < len(chassisArray); i++ {
		for j := 0; j < len(chassisArray[i].robotTypes); j++ {
			if chassisArray[i].robotTypes[j].Name == string(t) {
				result = &chassisArray[i].robotTypes[j]
				break
			}
		}
	}
	chassisMutex.RUnlock()
	return result
}

//机器人类型
type RobotTypeStruct struct {
	Name     string
	Code     string
	Function RobotFunction //功能列表，目前每个类型可以有自己独立的64个功能位开关
	Chassis  *Chassis      //机器人底盘
}

//功能号,0-63
type RobotFunction uint64

//判定某个功能号是否开放，0-63
func (f RobotFunction) IsOpen(function enum.RobotFunctionEnum) bool {
	if function < 64 {
		var bf [8]byte
		*(*uint64)(unsafe.Pointer(&bf[0])) = uint64(f)
		btIndex := function / 8
		offset := function % 8
		value := byte(1 << uint(offset))
		return bf[btIndex]&value == value
	}
	return false
}

//打开某个功能号
func (f *RobotFunction) Open(function enum.RobotFunctionEnum) {
	if function < 64 {
		var bf [8]byte
		*(*uint64)(unsafe.Pointer(&bf[0])) = uint64(*f)
		btIndex := function / 8
		offset := function % 8
		value := byte(1 << uint(offset))
		bf[btIndex] = bf[btIndex] | value
		*f = *(*RobotFunction)(unsafe.Pointer(&bf[0]))
	}
}

//关闭某个功能号
func (f *RobotFunction) Close(function enum.RobotFunctionEnum) {
	if function < 64 {
		var bf [8]byte
		*(*uint64)(unsafe.Pointer(&bf[0])) = uint64(*f)
		btIndex := function / 8
		offset := function % 8
		value := byte(1 << uint(offset))
		value = ^value
		bf[btIndex] = bf[btIndex] & value
		*f = *(*RobotFunction)(unsafe.Pointer(&bf[0]))
	}
}

func appendRobotChassis() *Chassis {
	chassisArray = append(chassisArray, Chassis{
		robotTypes: make([]RobotTypeStruct, 0, 16),
	})
	return &chassisArray[len(chassisArray)-1]
}

//从数据库中加载
func InitRobotTypeFromDB(RDB *sqlx.DB) error {
	rows, err := RDB.Queryx(`select ifnull(a.supplier,''),a.chassis_model,ifnull(b.model_name,''),b.model_value,ifnull(b.function_num,0) 
from device_robot_chassis a inner join device_robot_model b on a.id=b.chassis_id where a.del_flag=0 and b.del_flag=0 order by a.id`)
	if err != nil {
		return err
	}
	supplier, chassisName, robotTypeCode, robotTypeName := "", "", "", ""
	lastChassisName := ""
	functionNum := uint64(0)
	chassisMutex.Lock()
	var chassis *Chassis
	for rows.Next() {
		err = rows.Scan(&supplier, &chassisName, &robotTypeCode, &robotTypeName, &functionNum)
		if err != nil {
			fmt.Println("执行初始化RobotType发生错误：", err)
			continue
		}
		if lastChassisName != chassisName {
			if lastChassisName != "" || chassis == nil {
				chassis = appendRobotChassis()
			}
			lastChassisName = chassisName
			//一个新的
			chassis.Name = chassisName
			chassis.Supplier = supplier
			if chassisName == "MIR" {
				chassis.DefaultRealOriginX = 0
				chassis.DefaultRealOriginY = 0
				chassis.DefaultResolution = 0.05
			} else {
				chassis.DefaultRealOriginX = 0
				chassis.DefaultRealOriginY = 0
				chassis.DefaultResolution = 0
			}
		}
		//注册机器人类型
		chassis.robotTypes = append(chassis.robotTypes, RobotTypeStruct{
			Name:     robotTypeName,
			Code:     robotTypeCode,
			Function: RobotFunction(functionNum),
			Chassis:  chassis,
		})
	}
	chassisMutex.Unlock()
	rows.Close()
	sqlxRDB = RDB
	sys_res_mq.SubRobotTypeChange(robotTypeChange)
	return nil
}

func InitRobotTypeFromOrm(RDB *gorm.DB) error {
	rows, err := RDB.Raw(`select ifnull(a.supplier,''),a.chassis_model,ifnull(b.model_name,''),b.model_value,ifnull(b.function_num,0) 
from device_robot_chassis a inner join device_robot_model b on a.id=b.chassis_id where a.del_flag=0 and b.del_flag=0 order by a.id`).Rows()
	if err != nil {
		return err
	}
	supplier, chassisName, robotTypeCode, robotTypeName := "", "", "", ""
	lastChassisName := ""
	functionNum := uint64(0)
	chassisMutex.Lock()
	var chassis *Chassis
	for rows.Next() {
		err = rows.Scan(&supplier, &chassisName, &robotTypeCode, &robotTypeName, &functionNum)
		if err != nil {
			fmt.Println("执行初始化RobotType发生错误：", err)
			continue
		}
		if lastChassisName != chassisName {
			if lastChassisName != "" || chassis == nil {
				chassis = appendRobotChassis()
			}
			lastChassisName = chassisName
			//一个新的
			chassis.Name = chassisName
			chassis.Supplier = supplier
			if chassisName == "MIR" {
				chassis.DefaultRealOriginX = 0.05
				chassis.DefaultRealOriginY = 0
				chassis.DefaultResolution = 0
			} else {
				chassis.DefaultRealOriginX = 0
				chassis.DefaultRealOriginY = 0
				chassis.DefaultResolution = 0
			}
		}
		//注册机器人类型
		chassis.robotTypes = append(chassis.robotTypes, RobotTypeStruct{
			Name:     robotTypeName,
			Code:     robotTypeCode,
			Function: RobotFunction(functionNum),
			Chassis:  chassis,
		})
	}
	chassisMutex.Unlock()
	rows.Close()
	//执行订阅变动处理
	ormRdb = RDB
	sys_res_mq.SubRobotTypeChange(robotTypeChange)
	return nil
}

var (
	ormRdb  *gorm.DB
	sqlxRDB *sqlx.DB
)

func modifyChassisModelFromSqlx(ChassisModel, robotType string) {
	var rows *sqlx.Rows
	var err error
	if robotType == "" {
		rows, err = sqlxRDB.Queryx(`select ifnull(a.supplier,''),a.chassis_model,ifnull(b.model_name,''),b.model_value,ifnull(b.function_num,0) 
from device_robot_chassis a left join device_robot_model b on a.id=b.chassis_id where a.del_flag=0 and b.del_flag=0 and a.chassis_model=?`, ChassisModel)
	} else {
		rows, err = sqlxRDB.Queryx(`select ifnull(a.supplier,''),a.chassis_model,ifnull(b.model_name,''),b.model_value,ifnull(b.function_num,0) 
from device_robot_chassis a left join device_robot_model b on a.id=b.chassis_id where a.del_flag=0 and b.del_flag=0 and a.chassis_model=? and b.model_value=?`, ChassisModel, robotType)
	}
	if err != nil {
		log.WithError(err).Error("修改底盘信息发生错误")
		return
	}
	var supplier, chassisName, robotTypeCode, robotTypeName string
	functionNum := uint64(0)
	var chassis *Chassis
	chassisMutex.Lock()
	for rows.Next() {
		err = rows.Scan(&supplier, &chassisName, &robotTypeCode, &robotTypeName, &functionNum)
		if err != nil {
			fmt.Println("执行初始化RobotType发生错误：", err)
			continue
		}
		if chassis == nil {
			for i := 0; i < len(chassisArray); i++ {
				if chassisArray[i].Name == chassisName {
					chassis = &chassisArray[i]
					break
				}
			}
			if chassis == nil {
				chassis = appendRobotChassis()
				chassis.Name = chassisName
			}
			chassis.Supplier = supplier
			if chassisName == "MIR" {
				chassis.DefaultRealOriginX = 0.05
				chassis.DefaultRealOriginY = 0
				chassis.DefaultResolution = 0
			} else {
				chassis.DefaultRealOriginX = 0
				chassis.DefaultRealOriginY = 0
				chassis.DefaultResolution = 0
			}
		}
		if robotTypeName == "" {
			continue
		}
		for i := 0; i < len(chassis.robotTypes); i++ {
			if chassis.robotTypes[i].Name == robotTypeName {
				chassis.robotTypes[i].Function = RobotFunction(functionNum)
				chassis.robotTypes[i].Code = robotTypeCode
				break
			}
		}
	}
	chassisMutex.Unlock()
	rows.Close()
}

func modifyChassisModelFromOrm(ChassisModel, robotType string) {
	var rows *sql.Rows
	var err error
	if robotType == "" {
		rows, err = ormRdb.Raw(`select ifnull(a.supplier,''),a.chassis_model,ifnull(b.model_name,''),b.model_value,ifnull(b.function_num,0) 
from device_robot_chassis a left join device_robot_model b on a.id=b.chassis_id where a.del_flag=0 and b.del_flag=0 and a.chassis_model=?`, ChassisModel).Rows()
	} else {
		rows, err = ormRdb.Raw(`select ifnull(a.supplier,''),a.chassis_model,ifnull(b.model_name,''),b.model_value,ifnull(b.function_num,0) 
from device_robot_chassis a left join device_robot_model b on a.id=b.chassis_id where a.del_flag=0 and b.del_flag=0 and a.chassis_model=? and b.model_value=?`, ChassisModel, robotType).Rows()
	}
	if err != nil {
		log.WithError(err).Error("修改底盘信息发生错误")
		return
	}
	var supplier, chassisName, robotTypeCode, robotTypeName string
	functionNum := uint64(0)
	var chassis *Chassis
	chassisMutex.Lock()
	for rows.Next() {
		err = rows.Scan(&supplier, &chassisName, &robotTypeCode, &robotTypeName, &functionNum)
		if err != nil {
			fmt.Println("执行初始化RobotType发生错误：", err)
			continue
		}
		if chassis == nil {
			for i := 0; i < len(chassisArray); i++ {
				if chassisArray[i].Name == chassisName {
					chassis = &chassisArray[i]
					break
				}
			}
			if chassis == nil {
				chassis = appendRobotChassis()
				chassis.Name = chassisName
			}
			chassis.Supplier = supplier
			if chassisName == "MIR" {
				chassis.DefaultRealOriginX = 0.05
				chassis.DefaultRealOriginY = 0
				chassis.DefaultResolution = 0
			} else {
				chassis.DefaultRealOriginX = 0
				chassis.DefaultRealOriginY = 0
				chassis.DefaultResolution = 0
			}
		}
		if robotTypeName == "" {
			continue
		}
		HasSet := false
		for i := 0; i < len(chassis.robotTypes); i++ {
			if chassis.robotTypes[i].Name == robotTypeName {
				chassis.robotTypes[i].Function = RobotFunction(functionNum)
				chassis.robotTypes[i].Code = robotTypeCode
				HasSet = true
				break
			}
		}
		if !HasSet { //增加新类型
			chassis.robotTypes = append(chassis.robotTypes, RobotTypeStruct{
				Name:     robotTypeName,
				Code:     robotTypeCode,
				Function: RobotFunction(functionNum),
				Chassis:  chassis,
			})
		}
	}
	chassisMutex.Unlock()
	rows.Close()
}

//机器人类型变动
func robotTypeChange(robotType, ChassisModel string, changeType sys_res_mq.ChangeType) {
	if ChassisModel == "" {
		return
	}
	if changeType == sys_res_mq.TypeDel {
		if robotType == "" {
			UnRegisterRobotChassis(ChassisModel)
			return
		}
		UnBindRobotType(ChassisModel, robotType)
		return
	}
	if sqlxRDB != nil {
		modifyChassisModelFromSqlx(ChassisModel, robotType)
		return
	}
	modifyChassisModelFromOrm(ChassisModel, robotType)
}
