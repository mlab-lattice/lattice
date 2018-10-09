package time

import (
	"time"
)

func New(t time.Time) *Time {
	return &Time{t}
}

type Time struct {
	time.Time
}

func (t *Time) DeepCopyInto(out *Time) {
	*out = *t
}
