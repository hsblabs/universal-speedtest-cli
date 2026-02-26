package color

import "os"

// ANSI Color Codes — empty strings when NO_COLOR is set or TERM=dumb.
var (
	Magenta string
	Bold    string
	Yellow  string
	Green   string
	Blue    string
	Red     string
	Reset   string
)

func init() {
	if os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb" {
		Magenta = "\x1b[35m"
		Bold    = "\x1b[1m"
		Yellow  = "\x1b[33m"
		Green   = "\x1b[32m"
		Blue    = "\x1b[34m"
		Red     = "\x1b[31m"
		Reset   = "\x1b[0m"
	}
}
