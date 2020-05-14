package client

func SafeSlice(str string, start int, end int) string {
	if len(str) == 0 {
		return str
	}

	s := start
	if s < 0 {
		s = 0
	}

	e := end
	if e > len(str) {
		e = len(str)
	}

	return str[s:e]
}
