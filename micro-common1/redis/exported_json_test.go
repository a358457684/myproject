package redis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type TestDTO struct {
	Name string
	Age  int64
	Time time.Time
}

func TestLPushJson(t *testing.T) {
	LPushJson(context.TODO(), "test", TestDTO{
		Name: "name1",
		Age:  11,
		Time: time.Now(),
	}, TestDTO{
		Name: "name2",
		Age:  12,
		Time: time.Now(),
	}, TestDTO{
		Name: "name3",
		Age:  13,
		Time: time.Now(),
	})
}

func TestLRangeJson(t *testing.T) {
	var result []TestDTO
	_ = LRangeJson(context.TODO(), &result, "test", 0, 20)
	fmt.Println("Slice after appending data:", result)
}
