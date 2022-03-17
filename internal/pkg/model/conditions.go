package model

import (
	"errors"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/gen/webhooks"
)

const (
	MaxCount     uint              = 10000             // 10k rows
	MaxTime      time.Duration     = 30 * time.Minute  // 30 min
	MaxSize      datasize.ByteSize = 100 * 1024 * 1024 // 100MB
	DefaultCount uint              = 1000
)

type Conditions struct {
	Count *uint
	Time  *time.Duration
	Size  *datasize.ByteSize
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
	if count != nil && *count > MaxCount {
		return errors.New("count is too high")
	}
	return nil
}

func (c *Conditions) SetSize(str *string) error {
	if str == nil {
		c.Size = nil
		return nil
	}

	var v datasize.ByteSize
	err := v.UnmarshalText([]byte(*str))
	if err != nil {
		return errors.New("invalid size value. use format X MB|KB")
	}

	if v > MaxSize {
		return errors.New("au, size is too big")
	}

	c.Size = &v
	return err
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

	if duration > MaxTime {
		return errors.New("time is too high")
	}
	c.Time = &duration
	return nil
}

func (c *Conditions) ReachCondition(count uint, time time.Duration, size uint64) bool {
	if c.Count == nil && c.Time == nil && c.Size == nil {
		*c.Count = DefaultCount
	}
	if c.Count != nil && count > *c.Count {
		return true
	}
	if c.Size != nil && datasize.ByteSize(size) > *c.Size {
		return true
	}
	if c.Time != nil && time > *c.Time {
		return true
	}
	return false
}

func (c *Conditions) Payload() *webhooks.Conditions {
	timeStr := c.Time.String()
	sizeStr := c.Size.String()
	return &webhooks.Conditions{
		Count: c.Count,
		Size:  &sizeStr,
		Time:  &timeStr,
	}
}
