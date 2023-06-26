package zanzibar

// ErrorType is used for error grouping.
type ErrorType int

const (
	// TChannelError are errors of type tchannel.SystemError
	TChannelError ErrorType = iota + 1
	// ClientException are client exceptions defined in the
	// client IDL.
	ClientException
	// BadResponse are errors reading client response such
	// as undefined exceptions, empty response.
	BadResponse
)

//go:generate stringer -type=ErrorType

// Error extends error interface to set meta fields about
// error that are logged to improve error debugging, as well
// as to facilitate error grouping and filtering in logs.
type Error interface {
	error
	// ErrorLocation is used for logging. It is the module
	// identifier that produced the error. It should be one
	// of client, middleware, endpoint.
	ErrorLocation() string
	// ErrorType is for error grouping.
	ErrorType() ErrorType
	// Unwrap to enable usage of func errors.As
	Unwrap() error
}
