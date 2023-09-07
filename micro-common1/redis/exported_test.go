package redis

import (
	"common/log"
	"context"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	Set(context.Background(), "test", time.Now().Add(time.Hour), time.Hour)
	ti, _ := Get(context.Background(), "test").Time()
	log.Info(ti)

}
