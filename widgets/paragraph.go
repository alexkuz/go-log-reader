// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package widgets

import (
	"image"

	termui "github.com/gizak/termui/v3"
	"github.com/mitchellh/go-wordwrap"
)

type RawParagraph struct {
	termui.Block
	Text      string
	WrapText  bool
}

func NewRawParagraph() *RawParagraph {
	return &RawParagraph{
		Block:     *termui.NewBlock(),
		WrapText:  true,
	}
}

func (self *RawParagraph) Draw(buf *termui.Buffer) {
	self.Block.Draw(buf)

	cells := ParseRawStyles(self.Text, termui.Theme.Table.Text)
	if self.WrapText {
		cells,_ = WrapCells(cells, uint(self.Inner.Dx()))
	}

	rows := termui.SplitCells(cells, '\n')

	for y, row := range rows {
		if y+self.Inner.Min.Y >= self.Inner.Max.Y {
			break
		}
		row = termui.TrimCells(row, self.Inner.Dx())
		for _, cx := range termui.BuildCellWithXArray(row) {
			x, cell := cx.X, cx.Cell
			buf.SetCell(cell, image.Pt(x, y).Add(self.Inner.Min))
		}
	}
}

func WrapCells(cells []termui.Cell, width uint) ([]termui.Cell, int) {
	str := termui.CellsToString(cells)
	wrapped := wordwrap.WrapString(str, width)
	
	return ForceWrap(wrapped, width, cells)
}

func ForceWrap(str string, width uint, cells []termui.Cell) ([]termui.Cell, int) {
	wrappedCells := []termui.Cell{}
	lineCount := 1

	var col uint = 0
	for i, char := range str {
		if char == '\n' {
			col = 0
			wrappedCells = append(wrappedCells, termui.Cell{Rune: '\n', Style: termui.StyleClear})
			lineCount++
		} else if col == width - 3 && len(str) > i + 3 && str[i+1] != '\n' && str[i+2] != '\n' && str[i+3] != '\n' {
			col = 0
			wrappedCells = append(wrappedCells,
				termui.Cell{Rune: char, Style: cells[i].Style},
				termui.Cell{Rune: ' ', Style: termui.StyleClear},
				termui.Cell{Rune: '⏎', Style: termui.NewStyle(termui.ColorYellow)},
				termui.Cell{Rune: '\n', Style: termui.StyleClear},
			)
			lineCount++
		} else {
			col++
			style := termui.StyleClear
			if len(cells) > i {
				// TODO: shouldn't get here
				style = cells[i].Style
			}
			wrappedCells = append(wrappedCells, termui.Cell{Rune: char, Style: style})
		}
	}

	return wrappedCells, lineCount
}