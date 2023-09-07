package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNextID(t *testing.T) {
	idSize := 100000
	var wg sync.WaitGroup
	wg.Add(idSize)
	m := make(map[Long]int)

	startTime := time.Now().UnixNano()
	q := make(chan Long, 0)
	go func(m map[Long]int) {
		for {
			select {
			case id := <-q:
				if _, ok := m[id]; ok {
					m[id]++
				} else {
					m[id] = 1
				}
				wg.Done()
			}
		}
	}(m)

	for i := 0; i < idSize; i++ {
		go func() {
			q <- NextId()
		}()
	}

	wg.Wait()
	endTime := time.Now().UnixNano()
	usedTime := time.Duration(endTime - startTime)
	fmt.Printf("Id size: %d\nTotal seconds:%d\nAvg nanos per id:%d\n", len(m), usedTime/time.Second, int64(usedTime)/int64(idSize))

	badSize := 0
	for _, v := range m {
		if v > 1 {
			badSize++
		}
	}

	if badSize > 0 {
		t.Fail()
	}
}
