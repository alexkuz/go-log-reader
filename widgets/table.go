// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package widgets

import (
	"image"

	termui "github.com/gizak/termui/v3"
)

/*Table is like:
┌ Awesome Table ───────────────────────────────────────────────┐
│  Col0          | Col1 | Col2 | Col3  | Col4  | Col5  | Col6  |
│──────────────────────────────────────────────────────────────│
│  Some Item #1  | AAA  | 123  | CCCCC | EEEEE | GGGGG | IIIII |
│──────────────────────────────────────────────────────────────│
│  Some Item #2  | BBB  | 456  | DDDDD | FFFFF | HHHHH | JJJJJ |
└──────────────────────────────────────────────────────────────┘
*/
type RawTable struct {
	termui.Block
	Rows          [][]string
	ColumnWidths  []int
	TextStyle     termui.Style
	RowSeparator  bool
	TextAlignment termui.Alignment
	RowStyles     map[int]termui.Style
	FillRow       bool
	ActiveRowIndex int
	SeparatorStyle termui.Style
	ScrollTop int

	// ColumnResizer is called on each Draw. Can be used for custom column sizing.
	ColumnResizer func()
}

func NewRawTable() *RawTable {
	return &RawTable{
		Block:         *termui.NewBlock(),
		TextStyle:     termui.Theme.Table.Text,
		RowSeparator:  true,
		RowStyles:     make(map[int]termui.Style),
		ColumnResizer: func() {},
		ActiveRowIndex: -1,
		SeparatorStyle: termui.Theme.Block.Border,
		ScrollTop: 0,
	}
}

func (self *RawTable) Draw(buf *termui.Buffer) {
	self.Block.Draw(buf)

	self.ColumnResizer()

	columnWidths := self.ColumnWidths
	if len(columnWidths) == 0 {
		columnCount := len(self.Rows[0])
		columnWidth := self.Inner.Dx() / columnCount
		for i := 0; i < columnCount; i++ {
			columnWidths = append(columnWidths, columnWidth)
		}
	}

	yCoordinate := self.Inner.Min.Y

	maxIndex := self.Inner.Dy() - 1
	if self.RowSeparator {
		maxIndex = maxIndex / 2
	}
	if self.ActiveRowIndex > self.ScrollTop + maxIndex {
		self.ScrollTop = self.ActiveRowIndex - maxIndex
	}
	if self.ScrollTop > self.ActiveRowIndex {
		self.ScrollTop = self.ActiveRowIndex
	}
	if self.ActiveRowIndex == -1 {
		self.ScrollTop = 0
	}

	// draw rows
	for i := self.ScrollTop; i < len(self.Rows) && yCoordinate < self.Inner.Max.Y; i++ {
		row := self.Rows[i]
		colXCoordinate := self.Inner.Min.X

		rowStyle := self.TextStyle
		// get the row style if one exists
		if style, ok := self.RowStyles[i]; ok {
			rowStyle = style
		}

		if self.FillRow {
			blankCell := termui.NewCell(' ', rowStyle)
			buf.Fill(blankCell, image.Rect(self.Inner.Min.X, yCoordinate, self.Inner.Max.X, yCoordinate+1))
		}

		// draw row cells
		for j := 0; j < len(row); j++ {
			col := ParseRawStyles(row[j], rowStyle)
			// draw row cell
			if len(col) > columnWidths[j] || self.TextAlignment == termui.AlignLeft {
				for _, cx := range termui.BuildCellWithXArray(col) {
					k, cell := cx.X, cx.Cell
					if k == columnWidths[j] || colXCoordinate+k == self.Inner.Max.X {
						cell.Rune = termui.ELLIPSES
						buf.SetCell(cell, image.Pt(colXCoordinate+k-1, yCoordinate))
						break
					} else {
						buf.SetCell(cell, image.Pt(colXCoordinate+k, yCoordinate))
					}
				}
			} else if self.TextAlignment == termui.AlignCenter {
				xCoordinateOffset := (columnWidths[j] - len(col)) / 2
				stringXCoordinate := xCoordinateOffset + colXCoordinate
				for _, cx := range termui.BuildCellWithXArray(col) {
					k, cell := cx.X, cx.Cell
					buf.SetCell(cell, image.Pt(stringXCoordinate+k, yCoordinate))
				}
			} else if self.TextAlignment == termui.AlignRight {
				stringXCoordinate := termui.MinInt(colXCoordinate+columnWidths[j], self.Inner.Max.X) - len(col)
				for _, cx := range termui.BuildCellWithXArray(col) {
					k, cell := cx.X, cx.Cell
					buf.SetCell(cell, image.Pt(stringXCoordinate+k, yCoordinate))
				}
			}
			colXCoordinate += columnWidths[j] + 1
		}

		// draw vertical separators
		separatorStyle := self.SeparatorStyle

		separatorXCoordinate := self.Inner.Min.X
		verticalCell := termui.NewCell(termui.VERTICAL_LINE, separatorStyle)
		for i, width := range columnWidths {
			if self.FillRow && i < len(columnWidths)-1 {
				verticalCell.Style.Bg = rowStyle.Bg
			} else {
				verticalCell.Style.Bg = self.Block.BorderStyle.Bg
			}

			separatorXCoordinate += width
			buf.SetCell(verticalCell, image.Pt(separatorXCoordinate, yCoordinate))
			separatorXCoordinate++
		}

		yCoordinate++

		// draw horizontal separator
		horizontalCell := termui.NewCell(termui.HORIZONTAL_LINE, separatorStyle)
		if self.RowSeparator && yCoordinate < self.Inner.Max.Y && i != len(self.Rows)-1 {
			buf.Fill(horizontalCell, image.Rect(self.Inner.Min.X, yCoordinate, self.Inner.Max.X, yCoordinate+1))
			yCoordinate++
		}
	}

	// draw UP_ARROW if needed
	if self.ScrollTop > 0 {
		buf.SetCell(
			termui.NewCell(termui.UP_ARROW, termui.NewStyle(termui.ColorWhite)),
			image.Pt(self.Inner.Max.X-1+self.PaddingRight, self.Inner.Min.Y),
		)
	}

	// draw DOWN_ARROW if needed
	if len(self.Rows) > int(self.ScrollTop)+self.Inner.Dy() {
		buf.SetCell(
			termui.NewCell(termui.DOWN_ARROW, termui.NewStyle(termui.ColorWhite)),
			image.Pt(self.Inner.Max.X-1+self.PaddingRight, self.Inner.Max.Y-1),
		)
	}
}
