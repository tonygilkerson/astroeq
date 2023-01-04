package main

import (
	"fmt"
	// "image/color"
	"github.com/tonygilkerson/astroeq/pkg/grid"
)

func main() {

	var grid grid.Grid
	grid.Configure(3, 4, 1, 1)

	grid.LoadGrid("AAAA\nBBBB\nCCCC")
	printGrid(grid)
	grid.LoadGrid("aaa\nbbb\ncccc")
	printGrid(grid)

}

func printGrid(grid grid.Grid) {

	for _, row := range grid.GetCells() {
		// fmt.Printf("row: %v\n", r)

		for _, cell := range row {

			fmt.Printf("[%c|%c] \t", cell.GetChar(), cell.GetPrevChar())
		}
		fmt.Println("--")
	}
}
