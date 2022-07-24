package snaps

import "fmt"

const (
	reset = "\u001b[0m"

	redBG       = "\u001b[48;5;225m"
	greenBG     = "\u001b[48;5;159m"
	boldGreenBG = "\u001b[48;5;23m"
	boldRedBG   = "\u001b[48;5;127m"

	dim    = "\u001b[2m"
	green  = "\u001b[38;5;22m"
	red    = "\u001b[38;5;52m"
	yellow = "\u001b[33;1m"
)

func diffEqualText(prefix, text string, addNewLine bool) string {
	return coloredText(prefix+text+stringTernary(addNewLine, newLine, ""), dim)
}

func diffDeleteText(prefix, text string, addNewLine bool) string {
	return fmt.Sprintf("%s%s%s%s%s%s", red, redBG, prefix, text, reset, stringTernary(addNewLine, newLine, ""))
}

func diffInsertText(prefix, text string, addNewLine bool) string {
	return fmt.Sprintf("%s%s%s%s%s%s", green, greenBG, prefix, text, reset, stringTernary(addNewLine, newLine, ""))
}

func coloredText(text string, code string) string {
	return fmt.Sprintf("%s%s%s", code, text, reset)
}
