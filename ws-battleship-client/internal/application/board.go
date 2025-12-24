package application

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

	empty = ' '
	dead  = 'X'
	alive = 'O'
)

type Cell rune

type Board [len(boardAlphabet)][len(boardAlphabet)]Cell

func (b Board) Render() {
	clearTerminal()

	fmt.Println(b.renderAlphabetHorizontal()) // a b c d e f ...
	fmt.Println(b.renderBorder(upperLeftCorner, upperRightCorner))

	for i := 0; i < b.size(); i++ {
		if (i+1)/b.size() == 0 {
			fmt.Print(b.alignLeft())
		}

		fmt.Println(b.renderRow(i))
	}

	fmt.Println(b.renderBorder(lowerLeftCorner, lowerRightCorner))
	fmt.Println(b.renderAlphabetHorizontal()) // a b c d e f ...
}

func (b Board) renderRow(rowIdx int) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprint(rowIdx + 1))
	builder.WriteRune(verticalLine)

	for columnIdx := 0; columnIdx < b.size(); columnIdx++ {
		cell := b[rowIdx][columnIdx]
		if cell == 0 {
			cell = empty
		}
		builder.WriteRune(rune(cell))

		if columnIdx < b.size()-1 {
			builder.WriteRune(' ')
		}
	}

	builder.WriteRune(verticalLine)
	builder.WriteRune(' ')
	builder.WriteString(fmt.Sprint(rowIdx + 1))
	return builder.String()
}

func (b Board) renderAlphabetHorizontal() string {
	var builder strings.Builder
	builder.WriteString(b.alignLeft())
	builder.WriteString("  ")

	for i := 0; i < b.size(); i++ {
		builder.WriteRune(rune(boardAlphabet[i]))
		builder.WriteRune(' ')
	}
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
	return builder.String()
}

func (b Board) alignLeft() string {
	buf := make([]byte, 0, 5)
	for range b.size() / 10 {
		buf = append(buf, ' ')
	}
	return string(buf)
}

func (b Board) size() int {
	return len(b)
}
