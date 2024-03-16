package utils_test

import (
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
	"gotest.tools/v3/assert"
)

func TestCircularBuffer1(t *testing.T) {
	cb := utils.MakeCircularBuffer[int](3)

	assert.Equal(t, cb.Size(), 0)
	assert.Equal(t, cb.Empty(), true)
	assert.Equal(t, cb.Full(), false)

	cb.Push(1)
	assert.Equal(t, cb.Size(), 1)
	assert.Equal(t, cb.Back(), 1)
	assert.Equal(t, cb.Front(), 1)
	cb.Push(2)
	assert.Equal(t, cb.Size(), 2)
	assert.Equal(t, cb.Back(), 2)
	assert.Equal(t, cb.Front(), 1)
	cb.Push(3)
	assert.Equal(t, cb.Size(), 3)
	assert.Equal(t, cb.Back(), 3)
	assert.Equal(t, cb.Front(), 1)
	cb.Push(4)
	assert.Equal(t, cb.Size(), 3)
	assert.Equal(t, cb.Back(), 4)
	assert.Equal(t, cb.Front(), 2)
	cb.Push(5)
	assert.Equal(t, cb.Size(), 3)
	assert.Equal(t, cb.Back(), 5)
	assert.Equal(t, cb.Front(), 3)
}
