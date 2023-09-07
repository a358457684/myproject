package redis

import (
	"common/log"
	"sync"
	"testing"
	"time"
)

func TestLocal(t *testing.T) {
	mutex := NewMutex("a")
	mutex.Unlock()
	group := sync.WaitGroup{}
	group.Add(3)
	go func() {
		defer group.Done()
		err := mutex.Lock()
		log.Warn(err)
		if err == nil {
			defer mutex.Unlock()
			log.Info("开始")
			time.Sleep(time.Second)
			log.Info("结束")
		}
	}()

	go func() {
		defer group.Done()
		err := mutex.TryLock()
		if err == nil {
			defer mutex.Unlock()
			log.Info("开始1")
		}
	}()

	go func() {
		defer group.Done()
		err := mutex.Lock()
		log.Error(err)
		if err == nil {
			defer mutex.Unlock()
			log.Info("开始2")
			time.Sleep(time.Second)
			log.Info("结束2")
		}
	}()

	group.Wait()
	log.Info("end")
}
