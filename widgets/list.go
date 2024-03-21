// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package widgets

import (
	"image"

	rw "github.com/mattn/go-runewidth"

	termui "github.com/gizak/termui/v3"
)

type List struct {
	termui.Block
	Rows             []string
	WrapText         bool
	TextStyle        termui.Style
	SelectedRow      int
	topRow           int
	SelectedRowStyle termui.Style
}

func NewList() *List {
	return &List{
		Block:            *termui.NewBlock(),
		TextStyle:        termui.Theme.List.Text,
		SelectedRowStyle: termui.Theme.List.Text,
	}
}

func (self *List) Draw(buf *termui.Buffer) {
	self.Block.Draw(buf)

	point := self.Inner.Min

	// adjusts view into widget
	if self.SelectedRow >= self.Inner.Dy()+self.topRow {
		self.topRow = self.SelectedRow - self.Inner.Dy() + 1
	} else if self.SelectedRow < self.topRow {
		self.topRow = self.SelectedRow
	}

	if self.WrapText && len(self.Rows) > 0 {
		extraLines := 0

		for row := self.topRow; row < len(self.Rows) && point.Y < self.Inner.Max.Y; row++ {
			cells := ParseRawStyles(self.Rows[row], self.TextStyle)
			_, lineCount := WrapCells(cells, uint(self.Inner.Dx()))
			if lineCount > 1 {
				extraLines += lineCount - 1
			}
			point = image.Pt(self.Inner.Min.X, point.Y+lineCount)
		}

		if self.SelectedRow > 0 && self.SelectedRow + extraLines >= self.Inner.Dy()+self.topRow {
			self.topRow = self.SelectedRow + extraLines - self.Inner.Dy() + 1
		}
		point = self.Inner.Min
	}

	var row int

	// draw rows
	for row = self.topRow; row < len(self.Rows) && point.Y < self.Inner.Max.Y; row++ {
		cells := ParseRawStyles(self.Rows[row], self.TextStyle)
		if self.WrapText {
			cells, _ = WrapCells(cells, uint(self.Inner.Dx()))
		}
		for j := 0; j < len(cells) && point.Y < self.Inner.Max.Y; j++ {
			style := cells[j].Style
			if row == self.SelectedRow {
				if style.Fg == self.TextStyle.Fg {
					style.Fg = self.SelectedRowStyle.Fg
				}
				if style.Bg == self.TextStyle.Bg {
					style.Bg = self.SelectedRowStyle.Bg
				}
				if style.Modifier == self.TextStyle.Modifier {
					style.Modifier = self.SelectedRowStyle.Modifier
				}
			}
			if cells[j].Rune == '\n' {
				point = image.Pt(self.Inner.Min.X, point.Y+1)
			} else {
				if point.X+1 == self.Inner.Max.X+1 && len(cells) > self.Inner.Dx() {
					buf.SetCell(termui.NewCell(termui.ELLIPSES, style), point.Add(image.Pt(-1, 0)))
					break
				} else {
					buf.SetCell(termui.NewCell(cells[j].Rune, style), point)
					point = point.Add(image.Pt(rw.RuneWidth(cells[j].Rune), 0))
				}
			}
		}
		point = image.Pt(self.Inner.Min.X, point.Y+1)
	}

	DrawScrollbar(buf, self.Inner, self.PaddingRight, self.topRow, row, len(self.Rows))
}

// ScrollAmount scrolls by amount given. If amount is < 0, then scroll up.
// There is no need to set self.topRow, as this will be set automatically when drawn,
// since if the selected item is off screen then the topRow variable will change accordingly.
func (self *List) ScrollAmount(amount int) {
	if len(self.Rows)-int(self.SelectedRow) <= amount {
		self.SelectedRow = len(self.Rows) - 1
	} else if int(self.SelectedRow)+amount < 0 {
		self.SelectedRow = 0
	} else {
		self.SelectedRow += amount
	}
}

func (self *List) ScrollUp() {
	self.ScrollAmount(-1)
}

func (self *List) ScrollDown() {
	self.ScrollAmount(1)
}

func (self *List) ScrollPageUp() {
	// If an item is selected below top row, then go to the top row.
	if self.SelectedRow > self.topRow {
		self.SelectedRow = self.topRow
	} else {
		self.ScrollAmount(-self.Inner.Dy())
	}
}

func (self *List) ScrollPageDown() {
	self.ScrollAmount(self.Inner.Dy())
}

func (self *List) ScrollHalfPageUp() {
	self.ScrollAmount(-int(termui.FloorFloat64(float64(self.Inner.Dy()) / 2)))
}

func (self *List) ScrollHalfPageDown() {
	self.ScrollAmount(int(termui.FloorFloat64(float64(self.Inner.Dy()) / 2)))
}

func (self *List) ScrollTop() {
	self.SelectedRow = 0
}

func (self *List) ScrollBottom() {
	self.SelectedRow = len(self.Rows) - 1
}
