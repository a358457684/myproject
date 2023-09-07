package redis

import (
	"common/log"
	"context"
	"errors"
	"sync"
	"time"
)

const LockPrefix = "lock:"

type Mutex struct {
	Key             string        //锁主键
	LeaseTimeOut    time.Duration //锁过期时间
	TryTimeOut      time.Duration //尝试加锁超时时间
	TryIntervalTime time.Duration //尝试加锁间隔时间
	TryErrMaxTimes  int           //尝试加锁最大失败次数
	l               sync.Mutex
}

func NewMutex(key string) *Mutex {
	return &Mutex{
		Key:             LockPrefix + key,
		LeaseTimeOut:    time.Minute,
		TryTimeOut:      time.Second * 30,
		TryIntervalTime: time.Millisecond * 300,
		TryErrMaxTimes:  10,
	}
}

func (m *Mutex) Lock() error {
	m.l.Lock()
	defer m.l.Unlock()
	errTimes := 0
	for {
		select {
		case <-time.After(m.TryTimeOut):
			return errors.New("加锁超时")
		default:
			log.Info("-----------")
			result, err := SetNX(context.Background(), m.Key, time.Now(), m.LeaseTimeOut).Result()
			if err != nil {
				errTimes++
			}
			if errTimes >= m.TryErrMaxTimes {
				return err
			}
			if result {
				log.Warn("加锁")
				return nil
			}
		}
		time.Sleep(m.TryIntervalTime)
	}
}

func (m *Mutex) TryLock() error {
	m.l.Lock()
	defer m.l.Unlock()
	result, err := SetNX(context.Background(), m.Key, time.Now(), m.LeaseTimeOut).Result()
	if err != nil {
		return err
	}
	if result {
		return nil
	}
	return errors.New("尝试加锁失败，已被他人占用")
}

func (m *Mutex) Unlock() (bool, error) {
	result, err := Del(context.Background(), m.Key).Result()
	if err != nil {
		log.WithError(err).Errorf("解锁失败 %s", m.Key)
	} else {
		log.Debugf("解锁成功 %s", m.Key)
	}
	return result == 1, err
}
