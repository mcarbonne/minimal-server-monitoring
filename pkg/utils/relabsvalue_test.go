package utils_test

import (
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
	"gotest.tools/v3/assert"
)

func TestRelAbsValueParsingRelative(t *testing.T) {

	val, err := utils.RelativeAbsoluteValueFromString("5%")

	assert.NilError(t, err)
	assert.Equal(t, val.GetValue(100), uint64(5))
	assert.Equal(t, val.GetValue(1000), uint64(50))

	val, err = utils.RelativeAbsoluteValueFromString("-5 %")

	assert.ErrorContains(t, err, "illegal relative value")
	assert.Equal(t, val.GetValue(100), uint64(0))
	assert.Equal(t, val.GetValue(1000), uint64(0))
}

func TestRelAbsValueParsingAbsolute(t *testing.T) {

	val, err := utils.RelativeAbsoluteValueFromString("5")

	assert.NilError(t, err)
	assert.Equal(t, val.GetValue(100), uint64(5))
	assert.Equal(t, val.GetValue(1000), uint64(5))

	val, err = utils.RelativeAbsoluteValueFromString(" 5 mb")

	assert.NilError(t, err)
	assert.Equal(t, val.GetValue(100), uint64(5000000))

	val, err = utils.RelativeAbsoluteValueFromString(" 5 gi ")

	assert.NilError(t, err)
	assert.Equal(t, val.GetValue(100), uint64(5*1024*1024*1024))

	val, err = utils.RelativeAbsoluteValueFromString("-5")

	assert.ErrorContains(t, err, "invalid syntax")
	assert.Equal(t, val.GetValue(100), uint64(0))
	assert.Equal(t, val.GetValue(1000), uint64(0))
}
