package main

func formatRed(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func formatYellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func formatCyan(s string) string {
	return "\033[36m" + s + "\033[0m"
}

func formatGreen(s string) string {
	return "\033[92m" + s + "\033[0m"
}

func formatAllYellow(ss []string) []string {
	formatted := make([]string, 0)

	for _, s := range ss {
		formatted = append(formatted, formatYellow(s))
	}

	return formatted
}
