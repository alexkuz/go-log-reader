package widgets

import (
	"strings"

	termui "github.com/gizak/termui/v3"
)

func ParseRawStyles(s string, defaultStyle termui.Style) []termui.Cell {
	cells := []termui.Cell{}
	runes := []rune(s)

	style := defaultStyle
	for i := 0; i < len(runes); i++ {
		_rune := runes[i]
		if _rune == 27 && runes[i+1] == '[' {
			style = defaultStyle
			i += 2
			for {
				if runes[i+1] == ';' || runes[i+1] == 'm' {
					switch (runes[i] - '0') {
					case 0:
						style.Modifier = termui.ModifierClear
					case 1:
						style.Modifier = termui.ModifierBold
					case 4:
						style.Modifier = termui.ModifierUnderline
					}
					if runes[i+1] == 'm' {
						i += 2
						break
					}
					i += 2
				} else if runes[i+2] == ';' || runes[i+2] == 'm' {
					color := runes[i+1] - '0'
					switch  (runes[i] - '0') {
					case 3:
						if color == 9 {
							style.Fg = defaultStyle.Fg
						} else {
							style.Fg = termui.Color(color)
						}
					case 4:
						if color == 9 {
							style.Bg = defaultStyle.Bg
						} else {
							style.Bg = termui.Color(color)
						}
					}
					if runes[i+2] == 'm' {
						i += 3
						break
					}
					i += 3
				} else {
					break
				}
			}
			_rune = runes[i]
		}
		cells = append(cells, termui.Cell{Rune: _rune, Style: style})
	}

	return cells
}

func StripAsciiCodes(str string) string {
	runes := []rune(str)
	stripped := []rune{}

	for i := 0; i < len(runes); i++ {
		_rune := runes[i]
		if _rune == 27 && runes[i+1] == '[' {
			idx := strings.IndexRune(str[i:], 'm')
			if idx > -1 {
				i += idx
				continue
			}
		}
		stripped = append(stripped, _rune)
	}

	return string(stripped)
}