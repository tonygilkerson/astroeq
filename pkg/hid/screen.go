package hid

import (
	"image/color"

	"tinygo.org/x/drivers/ssd1351"
	"tinygo.org/x/tinyfont"

	"github.com/tonygilkerson/astroeq/pkg/driver"
)

// The Screen properties are used to determine what is written to the display
type Screen struct {
	Grid
	displayDevice ssd1351.Device
	font          tinyfont.Font
	fontColor     color.RGBA
	bodyText      string
	// RA Data
	tracking  driver.RaValue
	direction driver.RaValue
	position  uint32
}

func (screen *Screen) GetPosition() uint32 {
	return screen.position
}
func (screen *Screen) SetPosition(position uint32) {
	screen.position = position
}

func (screen *Screen) GetBodyText() string {
	return screen.bodyText
}
func (screen *Screen) SetBodyText(text string) {
	screen.bodyText = text
}

func (screen *Screen) GetTracking() driver.RaValue {
	return screen.tracking
}
func (screen *Screen) SetTracking(tracking driver.RaValue) {
	screen.tracking = tracking
}

func (screen *Screen) GetDirection() driver.RaValue {
	return screen.tracking
}
func (screen *Screen) SetDirection(direction driver.RaValue) {
	screen.tracking = direction
}

func (screen *Screen) WriteLines(text string) {

	screen.LoadGrid(text)

	var x, y int16
	black := color.RGBA{0, 0, 0, 255}

	for r, row := range screen.GetCells() {
		for c, col := range row {
			cell := col
			// DEVTODO might need to add width and height back, also the x and y seem backward do I have the screen rotated
			//         undo the hard code when you figure it out
			x = int16(10*c) + 10
			y = int16(15*r) + 15

			// x = int16(screen.grid.GetWidth()*c) + 5
			// y = int16(screen.grid.GetHeight()*r) + 20
			if cell.IsDirty() {
				cells := screen.GetCells()
				tinyfont.WriteLine(&screen.displayDevice, &screen.font, x, y, string(cells[r][c].GetPrevChar()), black) // erase the previous character
				tinyfont.WriteLine(&screen.displayDevice, &screen.font, x, y, string(cells[r][c].GetChar()), screen.fontColor)
			}
		}
	}
}
