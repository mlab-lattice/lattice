package color

import (
	"github.com/fatih/color"
)

type Color func(format string, a ...interface{}) string

var (
	Success       Color = color.GreenString
	BoldHiSuccess Color = color.New(color.Bold).Add(color.FgHiGreen).SprintfFunc()
	Failure       Color = color.RedString
	BoldHiFailure Color = color.New(color.Bold).Add(color.FgHiRed).SprintfFunc()
	Warning       Color = color.YellowString
	BoldHiWarning Color = color.New(color.Bold).Add(color.FgHiYellow).SprintfFunc()
	ID            Color = color.HiCyanString
	Bold          Color = color.New(color.Bold).SprintfFunc()

	Black Color = color.New(color.FgHiBlack).SprintfFunc()
)
