package utils_test

import (
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
	"gotest.tools/v3/assert"
)

func TestIsNameValid(t *testing.T) {
	assert.Equal(t, utils.IsNameValid("abc_123_ABC"), true)
	assert.Equal(t, utils.IsNameValid("abc_123_ABC."), false)
	assert.Equal(t, utils.IsNameValid("abc_123_ABC/"), false)
}
