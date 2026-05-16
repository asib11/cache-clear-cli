package ui

import (
	"fmt"
	"strings"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	white  = "\033[37m"
	gray   = "\033[90m"
)

func Banner() {
	fmt.Println(cyan + bold + `
╔╦╗╔═╗╔╦╗╦ ╦╔═╗╔═╗   ╔═╗╔═╗╔═╗╦ ╦╔═╗   ╔═╗╦  ╦
║║║║╣  ║║║ ║╚═╗╠═╣───║  ╠═╣║  ╠═╣║╣ ───║  ║  ║
╩ ╩╚═╝═╩╝╚═╝╚═╝╩ ╩   ╚═╝╩ ╩╚═╝╩ ╩╚═╝   ╚═╝╩═╝╩
` + reset + gray + "  Medusa v2 + Next.js Cache Manager  |  golang edition" + reset + "\n")
}

func Section(title string) {
	fmt.Printf("\n%s%s── %s %s──%s\n", bold, cyan, title, gray, reset)
}

func Success(msg string) {
	fmt.Printf("  %s✓%s  %s\n", green+bold, reset, msg)
}

func Error(msg string) {
	fmt.Printf("  %s✗%s  %s%s%s\n", red+bold, reset, red, msg, reset)
}

func Warn(msg string) {
	fmt.Printf("  %s⚠%s  %s%s%s\n", yellow+bold, reset, yellow, msg, reset)
}

func Info(label, value string) {
	fmt.Printf("  %s%-22s%s %s%s%s\n", gray, label, reset, white+bold, value, reset)
}

func Stat(label, value, unit string) {
	fmt.Printf("  %s%-22s%s %s%s%s %s%s%s\n",
		gray, label, reset,
		cyan+bold, value, reset,
		gray, unit, reset,
	)
}

func Divider() {
	fmt.Println(gray + "  " + strings.Repeat("─", 44) + reset)
}

func Prompt(msg string) {
	fmt.Printf("\n%s%s?%s %s%s%s ", bold, yellow, reset, bold, msg, reset)
}

func Step(n int, total int, msg string) {
	fmt.Printf("\n  %s[%d/%d]%s %s\n", gray, n, total, reset, msg)
}
