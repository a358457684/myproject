package redis

import (
	"common/log"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	Name     string
	Age      int
	Birthday time.Time
}

func TestSetAndGet(t *testing.T) {
	user := User{
		Name:     "张三2",
		Age:      272,
		Birthday: time.Now(),
	}
	ctx := context.TODO()
	err := HSetJson(ctx, "user", "2", user)
	assert.Nil(t, err)
	result := User{}
	err = HGetJson(ctx, &result, "user", "2")
	assert.Nil(t, err)
	log.Info(result)
	userMap := make(map[string]User)
	err = HGetALLJson(ctx, userMap, "user")
	log.Warn(err)
	assert.Nil(t, err)
	//log.Warn(users)
}

type TelContet struct {
	Addr     string `redisDB:"addr,0"`
	Telphone string `redisDB:"tel"`
	Mobile   string
}

type RedisOffice struct {
	gg         string
	OfficeName string `redisDB:"Name,1"`
	OfficeID   string `redisDB:"ID,0"`
	OfficeCode string `redisDB:"Code"`
	Secret     string
	TelContet  `redisDB:",2"`
}

func TestSaveTable(t *testing.T) {
	offices := make([]RedisOffice, 0, 3)
	var of RedisOffice
	of.Addr = "湖北武汉"
	of.Secret = "234243Key"
	of.OfficeName = "测试医院23"
	of.OfficeID = "Test01"
	of.OfficeCode = "Code_0001"
	offices = append(offices, of)
	SaveTable(offices)
}

func TestFind(t *testing.T) {
	offices := make([]RedisOffice, 0, 3)
	Find(&offices, map[string]string{
		"Name": "测试医院23",
		"ID":   "Test01",
	})
	fmt.Println(offices)
}

func TestFirst(t *testing.T) {
	office := RedisOffice{}
	office.OfficeID = "Test01"
	office.OfficeName = "测试医院23"
	//查找 ID=Test01,名称为 测试医院23的第一个返回
	First(&office)
	fmt.Println(office)
	office.Mobile = "13129903278"
	Update(&office)
	office.Mobile = ""
	office.Addr = ""
	First(&office)
	fmt.Println(office)
	Delete("RedisOffice")
	//Delete(office)

}
