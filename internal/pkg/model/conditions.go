package model

import (
	"github.com/oriser/regroup"
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
	if count < 0 || count > MAX_COUNT {
		return Error("invalid count value")
	}
	*c.Count = count
	return nil
}

func (c Conditions) SetTime(str string) error {
	var myExp = regroup.MustCompile(`(?P<value>\d+)(?P<unit>s|m)`)

	match, err := myExp.Groups(str)
	if err != nil {
		panic(err)
	}
	seconds := match["value"]
	if match["unit"] == "m" {
		seconds = seconds * 60
	}

	*c.Time = seconds
	return nil
}

func (c Conditions) SetSize(str string) error {
	var myExp = regroup.MustCompile(`(?P<value>\d+).?(?P<unit>KB|MB)`)

	match, err := myExp.Groups(str)
	if err != nil {
		panic(err)
	}
	multiplyBy := 1024
	if match["unit"] == "MB" {
		multiplyBy = multiplyBy * 1034
	}

	*c.Size = match["value"] * multiplyBy
	return nil
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
