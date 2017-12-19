package cli

type OutputFormat int

const (
	JSON_OUTPUT  OutputFormat = iota
	TABLE_OUTPUT OutputFormat = iota
)

func GetTypeFromString(typeString string) OutputFormat {
	switch typeString {
	case "json":
		return JSON_OUTPUT
	case "table":
		return TABLE_OUTPUT
	}
	panic(typeString + " is invalid output type")
}
