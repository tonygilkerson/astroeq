package main

import (
	"fmt"
	// "image/color"
	"github.com/tonygilkerson/astroeq/pkg/grid"
)

func main() {

	var grid grid.Grid
	grid.ConfigureGrid(3, 4)

	grid.LoadGrid("AAAA\nBBBB\nCCCC")
	printGrid(grid)
	// grid.LoadGrid("aaaa\nbbbb\ncccc")
	// printGrid(grid)
	grid.LoadGrid("1111\n")
	printGrid(grid)

}

func printGrid(grid grid.Grid) {

	for _, row := range grid.GetCells() {
		// fmt.Printf("row: %v\n", r)

		for _, cell := range row {

			fmt.Printf("[%c|%c] \t", cell.GetChar(), cell.GetPrevChar())
		}
		fmt.Printf("\n")
	}
	fmt.Printf("-----------------------------------------------\n")
}
