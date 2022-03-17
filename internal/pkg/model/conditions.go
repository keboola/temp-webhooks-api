package model

import (
	"strconv"
	"time"
)

const (
	MaxCount     uint          = 10000             // 10k rows
	MaxTime      time.Duration = 30 * time.Minute  // 30 min
	MaxSize      uint64        = 100 * 1024 * 1024 // 100MB
	DefaultCount uint          = 1000
)

type Conditions struct {
	Count *uint
	Time  *time.Duration
	Size  *uint64
}

func NewConditions() Conditions {
	return Conditions{
		Count: nil,
		Time:  nil,
		Size:  nil,
	}
}

func (c *Conditions) SetCount(count *uint) error {
	c.Count = count
	return nil
}

func (c *Conditions) SetTime(str *string) error {
	if str == nil {
		c.Time = nil
		return nil
	}

	duration, err := time.ParseDuration(*str)
	if err != nil {
		return err
	}
	*c.Time = duration
	return nil
}

func (c *Conditions) SetSize(str *string) error {
	if str == nil {
		c.Size = nil
		return nil
	}

	parsed, err := strconv.ParseUint(*str, 0, 64)
	*c.Size = parsed
	return err
}

func (c *Conditions) ReachCondition(count uint, time time.Duration, size uint64) bool {
	if c.Count == nil && c.Time == nil && c.Size == nil {
		*c.Count = DefaultCount
	}
	if c.Count != nil && count > *c.Count {
		return true
	}
	if c.Time != nil && time > *c.Time {
		return true
	}
	if c.Size != nil && size > *c.Size {
		return true
	}
	return false
}
