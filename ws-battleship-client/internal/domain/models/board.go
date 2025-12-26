package models

const (
	boardAlphabet = "abcdefghjk"
)

type CellType = rune

const (
	Empty CellType = ' '
	Dead  CellType = 'X'
	Alive CellType = 'O'
	Miss  CellType = '∙'
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
	if rowIdx >= byte(b.Size()) || colIdx >= byte(b.Size()) {
		return 0
	}
	return b[rowIdx][colIdx]
}

func (b Board) Size() int {
	return len(b)
}

func (b Board) Alphabet() []rune {
	return []rune(boardAlphabet)
}

func (b Board) Lines() []string {
	result := make([]string, b.Size())

	for i := 0; i < b.Size(); i++ {
		result[i] = b.renderRow(i)
	}
	return result
}

func (b Board) renderRow(rowIdx int) string {
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
