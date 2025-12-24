package model

import (
	"fmt"
	"strings"
)

const (
	boardAlphabet = "abcdefghjk"

	verticalLine     = '│'
	horizontalLine   = '─'
	upperLeftCorner  = '┌'
	lowerLeftCorner  = '└'
	upperRightCorner = '┐'
	lowerRightCorner = '┘'
)

type CellType = rune

const (
	Empty CellType = ' '
	Dead  CellType = 'X'
	Alive CellType = 'O'
	Miss  CellType = '*'
)

type Cell = rune

type Board [len(boardAlphabet)][len(boardAlphabet)]Cell

func (b Board) IsCellDead(rowIdx, colIdx byte) bool {
	cellType := b.GetCellType(rowIdx, colIdx)
	if cellType == 0 {
		return false
	}
	return cellType == Dead
}

func (b Board) IsCellEmpty(rowIdx, colIdx byte) bool {
	cellType := b.GetCellType(rowIdx, colIdx)
	if cellType == 0 {
		return true
	}
	return cellType == Empty || cellType == 0
}

func (b Board) GetCellType(rowIdx, colIdx byte) CellType {
	if rowIdx >= byte(b.size()) || colIdx >= byte(b.size()) {
		return 0
	}
	return b[rowIdx][colIdx]
}

func (b Board) Lines() []string {
	result := make([]string, 0, b.size()+4)

	result = append(result, b.renderAlphabet())                                // a b c d e f
	result = append(result, b.renderBorder(upperLeftCorner, upperRightCorner)) // ┌---------┐

	for i := 0; i < b.size(); i++ {
		var builder strings.Builder
		if (i+1)/b.size() == 0 {
			builder.WriteString(b.alignLeft())
		}
		builder.WriteString(b.renderRow(i))

		if (i+1)%b.size() != 0 {
			builder.WriteString(b.alignLeft())
		}
		result = append(result, builder.String()) // Example: 				10|**X O       *  X|10
	}

	result = append(result, b.renderBorder(lowerLeftCorner, lowerRightCorner)) // └---------┘
	result = append(result, b.renderAlphabet())                                // a b c d e f

	return result
}

func (b Board) renderRow(rowIdx int) string {
	if rowIdx >= b.size() {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprint(rowIdx + 1))
	builder.WriteRune(verticalLine)

	for colIdx := 0; colIdx < b.size(); colIdx++ {
		cell := b[rowIdx][colIdx]
		if cell == 0 {
			cell = Empty
		}
		builder.WriteRune(rune(cell))

		if colIdx < b.size()-1 {
			builder.WriteRune(' ')
		}
	}

	builder.WriteRune(verticalLine)
	builder.WriteString(fmt.Sprint(rowIdx + 1))
	return builder.String()
}

func (b Board) renderAlphabet() string {
	var builder strings.Builder
	builder.WriteString(b.alignLeft())
	builder.WriteString("  ")

	for i := 0; i < b.size(); i++ {
		builder.WriteRune(rune(boardAlphabet[i]))
		if i < b.size()-1 {
			builder.WriteRune(' ')
		}
	}

	builder.WriteString("  ")
	builder.WriteString(b.alignLeft())
	return builder.String()
}

func (b Board) renderBorder(leftCorner, rightCorner rune) string {
	var builder strings.Builder
	builder.WriteString(b.alignLeft())
	builder.WriteRune(' ')
	builder.WriteRune(leftCorner)

	for range b.size()*2 - 1 {
		builder.WriteRune(horizontalLine)
	}
	builder.WriteRune(rightCorner)
	builder.WriteRune(' ')
	builder.WriteString(b.alignLeft())
	return builder.String()
}

func (b Board) alignLeft() string {
	buf := make([]byte, 0, 3)
	for range b.size() / 10 {
		buf = append(buf, ' ')
	}
	return string(buf)
}

func (b Board) size() int {
	return len(b)
}
