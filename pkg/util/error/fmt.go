package error

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/util/runtime"
)

func Errorf(format string, vals ...interface{}) error {
	var location string

	function, file, line, err := runtime.TraceToFrame(3)

	if err != nil {
		location = "unable to retrieve function/file/line"
	} else {
		location = fmt.Sprintf("%v:%v:%v", function, file, line)
	}

	return fmt.Errorf("[%s] %s", location, fmt.Sprintf(format, vals...))
}
