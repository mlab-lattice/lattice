package testingsystem

import (
	"time"
)

type Version interface {
	Test() error
	Poll(time.Duration, time.Duration) error
}
