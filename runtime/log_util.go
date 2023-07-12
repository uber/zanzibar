package zanzibar

import "go.uber.org/zap"

var LogFieldErrTypeClientException = LogFieldErrorType("client_exception")

func LogFieldErrorType(errType string) zap.Field {
	return zap.String(logFieldErrorType, errType)
}

func LogFieldErrorLocation(loc string) zap.Field {
	return zap.String(logFieldErrorLocation, loc)
}
