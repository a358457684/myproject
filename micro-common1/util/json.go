package util

import (
	"github.com/suiyunonghen/DxCommonLib"
	"time"
)

type JsonTime struct {
	time.Time
}

func (j *JsonTime) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(`"2006-01-02 15:04:05"`, DxCommonLib.FastByte2String(data))
	(*j).Time = t
	return err
}

func (j *JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(j.Format(`"2006-01-02 15:04:05"`)), nil
}
