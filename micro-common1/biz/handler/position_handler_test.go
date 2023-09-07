package handler

import (
	"common/biz/manager"
	"common/log"
	"math"
	"testing"
)

func TestConvertPosition(t *testing.T) {
	type args struct {
		buildingId string
		floor      int
		srcType    manager.RobotType
		desType    manager.RobotType
		point      Point
	}
	tests := []struct {
		name       string
		args       args
		wantResult Point
		wantErr    bool
	}{
		{
			name: "相同类型",
			args: args{
				buildingId: "3beab1d5f8c24f29845d3f38d84921c1__A",
				floor:      9,
				srcType:    manager.RobotType("E2"),
				desType:    manager.RobotType("E2"),
				point: Point{
					X: 10,
					Y: 20,
				},
			},
			wantResult: Point{
				X: 10,
				Y: 20,
			},
			wantErr: false,
		}, {
			name: "Y2到E2",
			args: args{
				buildingId: "3beab1d5f8c24f29845d3f38d84921c1__A",
				floor:      9,
				srcType:    manager.RobotType("Y2"),
				desType:    manager.RobotType("E2"),
				point: Point{
					X: 10,
					Y: 20,
				},
			},
			wantResult: Point{
				X: 32,
				Y: 38,
			},
			wantErr: false,
		}, {
			name: "E2到Y2",
			args: args{
				buildingId: "3beab1d5f8c24f29845d3f38d84921c1__A",
				floor:      9,
				srcType:    manager.RobotType("E2"),
				desType:    manager.RobotType("Y2"),
				point: Point{
					X: 32,
					Y: 38,
				},
			},
			wantResult: Point{
				X: 10,
				Y: 20,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := ConvertPoint(tt.args.buildingId, tt.args.floor, tt.args.srcType, tt.args.desType, tt.args.point)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			log.Infof("ConvertPoint() gotResult = %v, want %v", gotResult, tt.wantResult)
			if math.Abs(tt.wantResult.X-gotResult.X) > 1 {
				t.Errorf("ConvertPoint() gotResult.X = %v, want %v", gotResult.X, tt.wantResult.X)
				return
			}
			if math.Abs(tt.wantResult.Y-gotResult.Y) > 1 {
				t.Errorf("ConvertPoint() gotResult.Y = %v, want %v", gotResult.Y, tt.wantResult.Y)
			}
		})
	}
}
