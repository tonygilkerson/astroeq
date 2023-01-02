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
	Char     rune
	PrevChar rune
	IsDirty  bool
}

type Grid struct {
	Cells  [][]Cell
	Rows   int
	Cols   int
	Width  int
	Height int
}

func (grid *Grid) Configure(rows int, cols int, width int, height int) {

	// Set once
	grid.Cols = cols
	grid.Rows = rows
	grid.Width = width
	grid.Height = height

	// Create a 2D Grid
	grid.Cells = make([][]Cell, rows)
	for i := range grid.Cells {
		grid.Cells[i] = make([]Cell, cols)
	}

}

func (grid *Grid) LoadGrid(str string) {

	rows := strings.Split(str, "\n")

	out:
		for r, row := range rows {

			for c, char := range row {

				// if the row is to big then it will be truncated
				if c >= grid.Cols {
					continue
				}

				// if at the bottom of the screen then we need to stop
				if r >= grid.Rows {
					break out
				}

				grid.SetCell(&grid.Cells[r][c], char)

			}

			// Clear out to the end of the line
			for i := len(row); i < grid.Cols; i++ {
				grid.SetCell(&grid.Cells[r][i], ' ')
			}

		}

}

func (grid *Grid) SetCell(cell *Cell, char rune) {
	cell.PrevChar = cell.Char
	cell.Char = char

	if cell.Char != cell.PrevChar {
		cell.IsDirty = true
	} else {
		cell.IsDirty = false
	}

}
