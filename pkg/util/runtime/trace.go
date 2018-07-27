package runtime

import (
	"fmt"
	"runtime"
)

// Return the function name, file name, and line number of the frame denoted by the frameNum argument
func TraceToFrame(frameNum int) (string, string, int, error) {
	var callers []uintptr

	// NOTE: we do not use the skip argument here since it can skip inlined functions
	_ = runtime.Callers(0, callers)
	frames := runtime.CallersFrames(callers)

	for i := 0; i < frameNum; i++ {
		_, more := frames.Next()
		if !more {
			return "", "", 0, fmt.Errorf(
				"runtime.CallersFrames returned %d frames when looking for frame %d (indexed from 0)",
				i+1, frameNum)
		}
	}

	frame, _ := frames.Next()

	return frame.Function, frame.File, frame.Line, nil
}

// Return the function name, file name, and line number of the caller
func Trace() (string, string, int, error) {
	return TraceToFrame(3)
}
