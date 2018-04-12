package markdown

import (
	"fmt"
)

func GetH1(header string) string {
	return fmt.Sprintf("# %s \n", header)
}

func GetH2(header string) string {
	return fmt.Sprintf("## %s \n", header)
}

func GetH3(header string) string {
	return fmt.Sprintf("### %s  \n", header)
}
