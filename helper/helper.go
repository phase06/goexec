package helper

import (
	"fmt"
	"strconv"
	"strings"
)

func parseAddr(raw string) (host, port string) {
	if i := strings.LastIndex(raw, ":"); i != -1 {
		return raw[:i], raw[i+1:]
	}
	return raw, ""
}

// StartupMessage print startup message
func StartupMessage(addr string, version string) {

	var logo string
	logo += "%s"
	logo += " ┌───────────────────────────────────────────────────┐\n"
	logo += " │ %s │\n"
	logo += " │ %s │\n"
	logo += " └───────────────────────────────────────────────────┘"
	logo += "%s"

	const (
		cBlack = "\u001b[90m"
		// cRed   = "\u001b[91m"
		cCyan = "\u001b[96m"
		// cGreen = "\u001b[92m"
		// cYellow  = "\u001b[93m"
		// cBlue    = "\u001b[94m"
		// cMagenta = "\u001b[95m"
		// cWhite   = "\u001b[97m"
		cReset = "\u001b[0m"
	)

	center := func(s string, width int) string {
		pad := strconv.Itoa((width - len(s)) / 2)
		str := fmt.Sprintf("%"+pad+"s", " ")
		str += s
		str += fmt.Sprintf("%"+pad+"s", " ")
		if len(str) < width {
			str += " "
		}
		return str
	}

	centerValue := func(s string, width int) string {
		pad := strconv.Itoa((width - len(s)) / 2)
		str := fmt.Sprintf("%"+pad+"s", " ")
		str += fmt.Sprintf("%s%s%s", cCyan, s, cBlack)
		str += fmt.Sprintf("%"+pad+"s", " ")
		if len(str)-10 < width {
			str += " "
		}
		return str
	}

	host, port := parseAddr(addr)
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	addr = "http://" + host + ":" + port

	mainLogo := fmt.Sprintf(logo,
		cBlack,
		centerValue(" goexec v"+version, 49),
		center(addr, 49),
		cReset,
	)

	fmt.Println(mainLogo)
}
