package zanzibar_test

import (
	"errors"
	"go.uber.org/zap"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
)

func TestErrorBuilder(t *testing.T) {
	eb := zanzibar.NewErrorBuilder("endpoint", "foo")
	err := errors.New("test error")
	zerr := eb.Error(err, zanzibar.TChannelError)

	assert.Equal(t, "endpoint::foo", zerr.ErrorLocation())
	assert.Equal(t, "TChannelError", zerr.ErrorType().String())
	assert.True(t, errors.Is(zerr, err))
}

func TestErrorBuilderLogFields(t *testing.T) {
	eb := zanzibar.NewErrorBuilder("client", "bar")
	testErr := errors.New("test error")
	table := []struct {
		err             error
		wantErrLocation string
		wantErrType     string
	}{
		{
			err:             eb.Error(testErr, zanzibar.ClientException),
			wantErrLocation: "client::bar",
			wantErrType:     "ClientException",
		},
		{
			err:             testErr,
			wantErrLocation: "~client::bar",
			wantErrType:     "ErrorType(0)",
		},
	}
	for i, tt := range table {
		t.Run("test"+strconv.Itoa(i), func(t *testing.T) {
			logFieldErrLocation := eb.LogFieldErrorLocation(tt.err)
			logFieldErrType := eb.LogFieldErrorType(tt.err)

			assert.Equal(t, zap.String("errorLocation", tt.wantErrLocation), logFieldErrLocation)
			assert.Equal(t, zap.String("errorType", tt.wantErrType), logFieldErrType)
		})
	}
}
