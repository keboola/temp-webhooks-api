package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetSize(t *testing.T) {
	t.Parallel()
	cond := NewConditions()
	size := "10MB"
	err := cond.SetSize(&size)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), cond.Size.MBytes())
}

func TestInvalidUnit(t *testing.T) {
	t.Parallel()
	cond := NewConditions()
	size := "dveste kilo"
	err := cond.SetSize(&size)
	assert.Contains(t, err.Error(), "invalid size value. use format X MB|KB")
}

func TestTooBigSize(t *testing.T) {
	t.Parallel()
	cond := NewConditions()
	size := "200 MB"
	err := cond.SetSize(&size)
	assert.Contains(t, err.Error(), "au, size is too big")
}

func TestSetTime(t *testing.T) {
	t.Parallel()
	cond := NewConditions()
	size := "20s"
	err := cond.SetTime(&size)
	assert.NoError(t, err)
	assert.Equal(t, 20*time.Second, *cond.Time)
}

func TestTooLong(t *testing.T) {
	t.Parallel()
	cond := NewConditions()
	testedTime := "50m"
	err := cond.SetTime(&testedTime)
	assert.Contains(t, err.Error(), "time is too high")
}

func TestInvalidTime(t *testing.T) {
	t.Parallel()
	cond := NewConditions()
	testedTime := "fsf"
	err := cond.SetTime(&testedTime)
	assert.Contains(t, err.Error(), "invalid time value. use format Xs|m")
}
