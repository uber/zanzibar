// TODO: (lu) to be generated

package baz

import (
	"runtime"
	"github.com/uber/zanzibar/runtime"
)

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)
	return zanzibar.GetDirnameFromRuntimeCaller(file)
}
