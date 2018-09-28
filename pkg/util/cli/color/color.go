package color

import (
	"fmt"
	"github.com/fatih/color"
)

type Color int

const (
	Success Color = iota
	BoldHiSuccess
	Failure
	BoldHiFailure
	Warning
	BoldHiWarning
	ID
	Bold
	Black
)

type Formatter func(format string, a ...interface{}) string

var (
	SuccessString       Formatter = color.GreenString
	BoldHiSuccessString           = color.New(color.Bold).Add(color.FgHiGreen).SprintfFunc()
	FailureString                 = color.RedString
	BoldHiFailureString           = color.New(color.Bold).Add(color.FgHiRed).SprintfFunc()
	WarningString                 = color.YellowString
	BoldHiWarningString           = color.New(color.Bold).Add(color.FgHiYellow).SprintfFunc()
	IDString                      = color.HiCyanString
	BoldString                    = color.New(color.Bold).SprintfFunc()
	BlackString                   = color.New(color.FgHiBlack).SprintfFunc()
	WhiteString                   = color.New(color.FgHiWhite).SprintfFunc()
	NormalString                  = fmt.Sprintf
)
