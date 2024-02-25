package widgets

import (
	. "github.com/gizak/termui/v3"
)

func ParseRawStyles(s string, defaultStyle Style) []Cell {
	cells := []Cell{}
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
						style.Modifier = ModifierClear
					case 1:
						style.Modifier = ModifierBold
					case 4:
						style.Modifier = ModifierUnderline
					}
					i += 2
					if runes[i+1] == 'm' {
						break
					}
				} else if runes[i+2] == ';' || runes[i+2] == 'm' {
					color := runes[i+1] - '0'
					switch  (runes[i] - '0') {
					case 3:
						if color == 9 {
							style.Fg = defaultStyle.Fg
						} else {
							style.Fg = Color(color)
						}
					case 4:
						if color == 9 {
							style.Bg = defaultStyle.Bg
						} else {
							style.Bg = Color(color)
						}
					}
					i += 3
					if runes[i+2] == 'm' {
						break
					}
				} else {
					break
				}
			}
			_rune = runes[i]
		}
		cells = append(cells, Cell{Rune: _rune, Style: style})
	}

	return cells
}