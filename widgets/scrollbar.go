package widgets

import (
	"image"

	termui "github.com/gizak/termui/v3"
)

func DrawScrollbar(buf *termui.Buffer, box image.Rectangle, right int, first int, last int, max int) {
	if max < 0 {
		return
	}
	height := box.Dy() - 3
	diff := max - last + first
	if diff <= 0 {
		diff = 1
	}
	pos := height * first / diff

	// draw UP_ARROW if needed
	if first > 0 {
		buf.SetCell(
			termui.NewCell(termui.UP_ARROW, termui.NewStyle(termui.ColorWhite)),
			image.Pt(box.Max.X - 1 + right, box.Min.Y),
		)
	}

	// draw DOWN_ARROW if needed
	if max > last {
		buf.SetCell(
			termui.NewCell(termui.DOWN_ARROW, termui.NewStyle(termui.ColorWhite)),
			image.Pt(box.Max.X - 1 + right, box.Max.Y - 1),
		)
	}

	if first > 0 || max > last {
		buf.SetCell(
			termui.NewCell('┃', termui.NewStyle(termui.ColorWhite)),
			image.Pt(box.Max.X - 1 + right, box.Min.Y + 1 + pos),
		)
	}
}