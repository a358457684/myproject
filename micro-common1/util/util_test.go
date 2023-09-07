package util

import (
	"common/log"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestRemove(t *testing.T) {
	RemoveDir("D:/filePath/mapPath")
}

type Position struct {
	X float64
	Y float64
}

type One struct {
	Name     string
	Age      string
	Position Position
}

type Two struct {
	Name     string
	Age      string
	Position interface{}
}

func TestNormal(t *testing.T) {
	one := One{
		Name: "张三",
		Age:  "38",
		Position: Position{
			X: 12,
			Y: 37.8,
		},
	}
	bytes, _ := json.Marshal(one)
	fmt.Println(string(bytes))
	two := Two{}
	err := json.Unmarshal(bytes, &two)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(two)
	marshal, _ := json.Marshal(two.Position)
	p := Position{}
	json.Unmarshal(marshal, &p)
	fmt.Println(p)
}

func TestSocket(t *testing.T) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:9010")
	if err != nil {
		log.WithError(err).Error("创建连接失败")
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.WithError(err).Error("创建连接失败2")
	}
	_, err = conn.Write([]byte{0x77, 0x77, 0x77, 0x00, 0x00, 0x00, 0x00, 0x09})
	for i := 0; i < 1; i++ {
		time.Sleep(time.Second * 4)
		_, err = conn.Write([]byte{0x61, 0x62, 0x63})
		time.Sleep(time.Second * 4)
		_, err = conn.Write([]byte{90, 8, 31, 141, 16, 0, 0, 30})
		log.Info("send")
	}
	//time.Sleep(60*time.Second)
	//result, err := ioutil.ReadAll(conn)
	//fmt.Println(string(result))
}
