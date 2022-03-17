package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetSize(t *testing.T) {
	t.Parallel()
	cond := NewConditions()
	size := "10MB"
	err := cond.SetSize(&size)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10485760), *cond.Size)
}

func TestTooBigSize(t *testing.T) {
	t.Parallel()
	cond := NewConditions()
	size := "200MB"
	err := cond.SetSize(&size)
	assert.Error(t, err, "size is too big")
}
