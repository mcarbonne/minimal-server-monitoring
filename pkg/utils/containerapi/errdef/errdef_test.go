package errdef_test

import (
	"errors"
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/containerapi/errdef"
	"gotest.tools/v3/assert"
)

func TestErrors1(t *testing.T) {
	baseError := errors.New("container not found")
	notFoundError := errdef.ErrorNotFound(baseError)
	assert.Equal(t, errdef.IsErrNotFound(baseError), false)
	assert.Equal(t, errdef.IsErrNotFound(notFoundError), true)
	assert.Equal(t, errdef.IsErrNotFound(nil), false)
}
