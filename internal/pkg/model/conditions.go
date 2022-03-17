package model

import (
	"strconv"
	"time"
)

type Conditions struct {
	Count *int
	Time  *time.Duration
	Size  *int64
}

func NewConditions() Conditions {
	return Conditions{
		Count: nil,
		Time:  nil,
		Size:  nil,
	}
}

const MAX_COUNT int = 10000                    // 10k rows
const MAX_TIME int64 = int64(30 * time.Minute) // 30 min
const MAX_SIZE int64 = 100 * 1024 * 1024       // 100MB

const DEFAULT_SIZE int = 1000 // 100MB

func (c Conditions) SetCount(count int) error {
	*c.Count = count
	return nil
}

func (c Conditions) SetTime(str string) error {

	seconds, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*c.Time = time.Second * seconds
	return nil
}

func (c Conditions) SetSize(str string) error {

	parsed, err := strconv.ParseInt(str, 0, 64)
	*c.Size = parsed
	return err
}

func (c Conditions) ReachCondition(count int, time time.Duration, size int64) bool {
	if c.Count == nil && c.Time == nil && c.Size == nil {
		*c.Count = DEFAULT_SIZE
	}

	if count > *c.Count {
		return true
	}
	if time > *c.Time {
		return true
	}
	if size > *c.Size {
		return true
	}

	return false
}
