// Code generated by "stringer -type=ErrorType"; DO NOT EDIT.

package zanzibar

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TChannelError-1]
	_ = x[ClientException-2]
	_ = x[BadResponse-3]
}

const _ErrorType_name = "TChannelErrorClientExceptionBadResponse"

var _ErrorType_index = [...]uint8{0, 13, 28, 39}

func (i ErrorType) String() string {
	i -= 1
	if i < 0 || i >= ErrorType(len(_ErrorType_index)-1) {
		return "ErrorType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _ErrorType_name[_ErrorType_index[i]:_ErrorType_index[i+1]]
}
