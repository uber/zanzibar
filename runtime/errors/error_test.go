package errors_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/runtime/errors"
)

func TestZErrorFactor(t *testing.T) {
	errFactory := errors.NewZErrorFactory("Test", "Foo")
	testErrTypes := []errors.ErrorType{errors.TChannelError, 0}
	for _, errType := range testErrTypes {
		err := errFactory.ZError(fmt.Errorf("test error"), errType)

		assert.Equal(t, "Test::Foo", err.ErrorLocation())
		assert.Equal(t, errType, err.ErrorType())
		assert.Equal(t, "test error", err.Error())
	}
}
