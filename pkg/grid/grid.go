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
	cells  [][]Cell
	rows   int
	cols   int
	width  int
	height int
}

func (grid *Grid) Configure(rows int, cols int, width int, height int) {

	// Set once
	grid.cols = cols
	grid.rows = rows
	grid.width = width
	grid.height = height

	// Create a 2D Grid
	grid.cells = make([][]Cell, rows)
	for i := range grid.cells {
		grid.cells[i] = make([]Cell, cols)
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
			if c >= grid.cols {
				continue
			}

			// if at the bottom of the screen then we need to stop
			if r >= grid.rows {
				break out
			}
			grid.cells[r][c].SetChar(char)
	

		}

		// Clear out to the end of the line
		for i := len(row); i < grid.cols; i++ {
			grid.cells[r][i].SetChar(' ')
		}

	}

}

func (grid *Grid) GetWidth() int {
	return grid.width
}
func (grid *Grid) GetHeight() int {
	return grid.height
}

func (cell *Cell) SetChar(char rune){
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
func (cell *Cell) SetIsDirty(dirty bool) {
	cell.isDirty = dirty
}
