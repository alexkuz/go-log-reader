// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package widgets

import (
	"image"

	. "github.com/gizak/termui/v3"
)

type RawParagraph struct {
	Block
	Text      string
	WrapText  bool
}

func NewRawParagraph() *RawParagraph {
	return &RawParagraph{
		Block:     *NewBlock(),
		WrapText:  true,
	}
}

func (self *RawParagraph) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	cells := ParseRawStyles(self.Text, Theme.Table.Text)
	if self.WrapText {
		cells = WrapCells(cells, uint(self.Inner.Dx()))
	}

	rows := SplitCells(cells, '\n')

	for y, row := range rows {
		if y+self.Inner.Min.Y >= self.Inner.Max.Y {
			break
		}
		row = TrimCells(row, self.Inner.Dx())
		for _, cx := range BuildCellWithXArray(row) {
			x, cell := cx.X, cx.Cell
			buf.SetCell(cell, image.Pt(x, y).Add(self.Inner.Min))
		}
	}
}