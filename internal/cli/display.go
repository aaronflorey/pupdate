package cli

import "strconv"

func safeDisplay(value string) string {
	for index := 0; index < len(value); index++ {
		if value[index] < 0x20 || value[index] == 0x7f {
			return strconv.QuoteToASCII(value)
		}
	}
	return value
}
