package domain

import (
	"bytes"
	"fmt"
)

const (
	boardAlphabet = "abcdefghij"
)

type CellType = rune

const (
	Null  CellType = 0
	Empty CellType = ' '
	Dead  CellType = '□'
	Ship  CellType = '■'
	Miss  CellType = '∙'
)

type Cell = rune

type Board [len(boardAlphabet)][len(boardAlphabet)]Cell

func (b *Board) IsCellDead(cellX, cellY byte) bool {
	cellType := b.GetCellType(cellX, cellY)
	if cellType == Null {
		return false
	}
	return cellType == Dead
}

func (b *Board) IsCellEmpty(cellX, cellY byte) bool {
	cellType := b.GetCellType(cellX, cellY)
	if cellType == Null {
		return true
	}
	return cellType == Empty
}

func (b *Board) GetCellType(cellX, cellY byte) CellType {
	if b.checkBounds(cellX, cellY) {
		return b[cellY][cellX]
	}
	return Null
}

func (b *Board) SetCell(cellX, cellY byte, cellType CellType) {
	if b.checkBounds(cellX, cellY) {
		b[cellY][cellX] = cellType
	}
}

func (b *Board) CellString(cellX, cellY byte) string {
	if b.checkBounds(cellX, cellY) {
		return fmt.Sprintf("%c%d", boardAlphabet[cellX], cellY+1)
	}
	return ""
}

func (b *Board) Size() int {
	return len(b)
}

func (b *Board) Alphabet() []rune {
	return []rune(boardAlphabet)
}

func (b *Board) Lines() []string {
	result := make([]string, b.Size())

	for i := 0; i < b.Size(); i++ {
		result[i] = b.renderRow(i)
	}
	return result
}

func (b *Board) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Grow(b.Size() * 2)

	for i := 0; i < b.Size(); i++ {
		for j := 0; j < b.Size(); j++ {
			if _, err := buf.WriteRune(b[i][j]); err != nil {
				return nil, err
			}
		}
	}
	return buf.Bytes(), nil
}

func (b *Board) UnmarshalBinary(buf []byte) error {
	buffer := bytes.NewBuffer(buf)

	var board Board
	for i := 0; i < board.Size(); i++ {
		for j := 0; j < board.Size(); j++ {
			r, _, err := buffer.ReadRune()
			if err != nil {
				return err
			}
			board[i][j] = r
		}
	}
	return nil
}

func (b *Board) checkBounds(cellX, cellY byte) bool {
	return cellX < byte(b.Size()) && cellY < byte(b.Size())
}

func (b *Board) renderRow(rowIdx int) string {
	if rowIdx >= b.Size() {
		return ""
	}

	row := make([]rune, 0, b.Size())
	for colIdx := 0; colIdx < b.Size(); colIdx++ {
		cell := b[rowIdx][colIdx]
		if cell == 0 {
			cell = Empty
		}
		row = append(row, rune(cell))

		if colIdx < b.Size()-1 {
			row = append(row, ' ')
		}
	}

	return string(row)
}
