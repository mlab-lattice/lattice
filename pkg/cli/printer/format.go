package printer

type Format string

const (
	FormatDefault Format = "default"
	FormatJSON    Format = "json"
	FormatTable   Format = "table"
)
