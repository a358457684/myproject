package cache

import (
	"common/redis"
	"context"
	"fmt"
)

//耗材缓存

const (
	consumableKey = "consumable:%s"
)

func SaveConsumable(officeID, robotID string, consumableCodeNumber map[string]int) error {
	codeRobotIDNumMap, err := GetConsumable(officeID)
	if err != nil {
		return err
	}
	for code, robotIDNum := range codeRobotIDNumMap {
		robotIDNum[robotID] = consumableCodeNumber[code]
	}
	for code, number := range consumableCodeNumber {
		if codeRobotIDNumMap[code] != nil {
			continue
		}
		robotIDNum := make(map[string]int)
		robotIDNum[robotID] = number
		codeRobotIDNumMap[code] = robotIDNum
	}
	return redis.SetJson(context.Background(), getConsumableKey(officeID), codeRobotIDNumMap, 0)
}

func GetConsumable(officeID string) (map[string]map[string]int, error) {
	codeRobotIDNumMap := make(map[string]map[string]int)
	err := redis.GetJson(context.Background(), &codeRobotIDNumMap, getConsumableKey(officeID))
	return codeRobotIDNumMap, err
}

func DelConsumable(officeID, robotID string) error {
	return SaveConsumable(officeID, robotID, make(map[string]int))
}

func getConsumableKey(officeID string) string {
	return fmt.Sprintf(consumableKey, officeID)
}
