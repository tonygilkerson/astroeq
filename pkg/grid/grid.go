package grid

import (
	"image/color"
	"strings"
)

var COLOR_BLACK color.Color = color.RGBA{0, 0, 0, 255}
var COLOR_RED color.Color = color.RGBA{255, 0, 0, 255}
var COLOR_GREEN color.Color = color.RGBA{0, 255, 0, 255}
var COLOR_BLUE color.Color = color.RGBA{0, 0, 255, 255}

type Cell struct {
	char     rune
	prevChar rune
	isDirty  bool
}

type Grid struct {
	cells    [][]Cell
	rowCount int
	colCount int
}

func (grid *Grid) Configure(rowCount int, colCount int) {

	// Set once
	grid.colCount = colCount
	grid.rowCount = rowCount

	// Create a 2D Grid
	grid.cells = make([][]Cell, rowCount)
	for i := range grid.cells {
		grid.cells[i] = make([]Cell, colCount)
	}

}

func (grid *Grid) GetCells() [][]Cell {
	return grid.cells
}

func (grid *Grid) LoadGrid(str string) {

	rows := strings.Split(str, "\n")


out:
	for r, row := range rows {

		for c, char := range row {

			// if the row is to big then it will be truncated
			if c >= grid.colCount {
				continue
			}

			// if at the bottom of the screen then we need to stop
			if r >= grid.rowCount {
				break out
			}
			grid.cells[r][c].SetChar(char)

		}

		// Clear out to the end of the line
		for cc := len(row); cc < grid.colCount; cc++ {
			grid.cells[r][cc].SetChar(' ')
		}

	}

	
	// Clear out remaining lines
	for rr := len(rows); rr < grid.rowCount; rr++ {
		for cc := 0; cc < grid.colCount; cc++ {
			grid.cells[rr][cc].SetChar(' ')
			// fmt.Printf("DEBUG rr: %v  cc: %v\n",rr,cc)
		}
	}

}

func (grid *Grid) GetWidth() int {
	return grid.colCount
}
func (grid *Grid) GetHeight() int {
	return grid.rowCount
}

func (cell *Cell) SetIsDirty(dirty bool) {
	cell.isDirty = dirty
}

func (cell *Cell) SetChar(char rune) {
	cell.prevChar = cell.char
	cell.char = char

	if cell.GetChar() != cell.GetPrevChar() {
		cell.SetIsDirty(true)
	} else {
		cell.SetIsDirty(false)
	}
}

func (cell *Cell) GetChar() rune {
	return cell.char
}
func (cell *Cell) GetPrevChar() rune {
	return cell.prevChar
}
func (cell *Cell) IsDirty() bool {
	return cell.isDirty
}
