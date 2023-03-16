package internal

import (
	"fmt"
	"unicode"

	"github.com/logrusorgru/aurora"
)

func Escape(str string) string {
	runes := []rune(str)

	for i := 0; i < len(runes); {
		r := runes[i]

		if unicode.IsPrint(r) {
			i++
			continue
		}

		var tmp string

		switch r {
		case '\a':
			tmp = "\\a"
		case '\b':
			tmp = "\\b"
		case '\f':
			tmp = "\\f"
		case '\n':
			tmp = "\\n"
		case '\r':
			tmp = "\\r"
		case '\t':
			tmp = "\\t"
		case '\v':
			tmp = "\\v"
		default:
			tmp = fmt.Sprintf("%U", r)
		}

		tmp = aurora.Black(tmp).BgYellow().String()

		runes = append(runes[:i], append([]rune(tmp), runes[i+1:]...)...)

		i += len([]rune(tmp))
	}

	return string(runes)
}
