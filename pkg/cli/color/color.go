package color

import (
	"github.com/fatih/color"
)

type Color func(format string, a ...interface{}) string

var (
	Success Color = color.GreenString
	Failure Color = color.RedString
	Warning Color = color.YellowString
)
