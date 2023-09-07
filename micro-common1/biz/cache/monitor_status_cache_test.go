package cache

import (
	"common/biz/enum"
	"common/log"
	"common/redis"
	"context"
	"fmt"
	redis2 "github.com/go-redis/redis/v8"
	"reflect"
	"testing"
)

func Test_getMonitorStatusKey(t *testing.T) {
	ctx := context.TODO()
	log.Info(redis.GetSet(ctx, "b", "b").Result())
	log.Info(redis.GetSet(ctx, "a", "b").Result())
}

func TestFindMonitorStatusAll(t *testing.T) {
	tests := []struct {
		name                string
		wantOfficeRobotKeys []OfficeRobotKey
		wantData            []MonitorStatusVo
		wantErr             bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOfficeRobotKeys, gotData, err := FindMonitorStatusAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindMonitorStatusAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOfficeRobotKeys, tt.wantOfficeRobotKeys) {
				t.Errorf("FindMonitorStatusAll() gotOfficeRobotKeys = %v, want %v", gotOfficeRobotKeys, tt.wantOfficeRobotKeys)
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("FindMonitorStatusAll() gotData = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}

func TestGetMonitorStatus(t *testing.T) {
	type args struct {
		officeId string
		robotId  string
	}
	tests := []struct {
		name     string
		args     args
		wantData MonitorStatusVo
		wantErr  bool
	}{
		{
			name: "one",
			args: args{
				officeId: "asdf",
				robotId:  "asdf",
			},
			wantData: MonitorStatusVo{},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, err := GetMonitorStatus(tt.args.officeId, tt.args.robotId)
			if err == redis2.Nil {
				fmt.Println(true)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMonitorStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("GetMonitorStatus() gotData = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}

func TestSaveMonitorStatus(t *testing.T) {
	type args struct {
		officeId string
		robotId  string
		data     MonitorStatusVo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "one",
			args: args{
				officeId: "adsf",
				robotId:  "qwer",
				data:     MonitorStatusVo{Status: enum.RsCharging},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SaveMonitorStatus(tt.args.officeId, tt.args.robotId, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SaveMonitorStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_findMonitorStatusKey(t *testing.T) {
	type args struct {
		ctx      context.Context
		officeId string
		robotId  string
	}
	tests := []struct {
		name     string
		args     args
		wantKeys []OfficeRobotKey
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKeys, err := findMonitorStatusKey(tt.args.ctx, tt.args.officeId, tt.args.robotId)
			if (err != nil) != tt.wantErr {
				t.Errorf("findMonitorStatusKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("findMonitorStatusKey() gotKeys = %v, want %v", gotKeys, tt.wantKeys)
			}
		})
	}
}
