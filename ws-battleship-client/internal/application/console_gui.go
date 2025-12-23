package application

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

type Board [len(boardAlphabet)][len(boardAlphabet)]rune

func (b Board) render() {
	clearTerminal()

	renderAlphabetHorizontal()
	renderBorder(upperLeftCorner, upperRightCorner)

	for i := 0; i < len(boardAlphabet); i++ {
		fmt.Printf("%c%c", boardAlphabet[i], verticalLine)

		for j := 0; j < len(boardAlphabet); j++ {
			cell := b[i][j]
			if cell == 0 {
				cell = empty
			}
			fmt.Printf("%c", cell)
			if j < len(boardAlphabet)-1 {
				fmt.Print(" ")
			}
		}

		fmt.Printf("%c\n", verticalLine)
	}
	renderBorder(lowerLeftCorner, lowerRightCorner)
}

func renderAlphabetHorizontal() {
	fmt.Print("  ")
	for i := 0; i < len(boardAlphabet); i++ {
		fmt.Printf("%c ", boardAlphabet[i])
	}
	fmt.Println()
}

func renderBorder(leftCorner, rightCorner rune) {
	fmt.Printf(" %c", leftCorner)
	for range len(boardAlphabet)*2 - 1 {
		fmt.Printf("%c", horizontalLine)
	}
	fmt.Printf("%c\n", rightCorner)
}

func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func clearTerminal() {
	switch runtime.GOOS {
	case "darwin", "linux":
		runCmd("clear")
	case "windows":
		runCmd("cmd", "/c", "cls")
	default:
		runCmd("clear")
	}
}

func renderLoop() {
	var b = Board{
		{0, 0, 0, 0, alive, alive, dead, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, alive, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, alive, 0, 0, dead, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, dead, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, alive, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, alive, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, alive, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, dead, 0, 0, dead, 0, dead, 0, 0},
	}
	b.render()
}
