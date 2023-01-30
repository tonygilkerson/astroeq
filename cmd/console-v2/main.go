package main

// DEVTODO replace console with console-v2, we just need one
import (
	"fmt"
	"machine"
	"strings"

	"time"

	"image/color"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"

	"github.com/tonygilkerson/astroeq/pkg/hid"
	"github.com/tonygilkerson/astroeq/pkg/msg"
)

type Screen struct {
	hid.Grid
	displayDevice st7789.Device
	font          tinyfont.Font
	fontColor     color.RGBA
	StatusText    string
	BodyText      string
	FilterText    string // Foo | Handset | RADriver | RADriverCmd
	Key0          machine.Pin
	Key1          machine.Pin
	Key2          machine.Pin
	Key3          machine.Pin
	ch            chan Screen
}

func main() {

	// run light
	runLight()

	/////////////////////////////////////////////////////////////////////////////
	// Display Device
	/////////////////////////////////////////////////////////////////////////////

	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 8_000_000,
		// LSBFirst:  false,
		Mode: 0,
		// DataBits:  0,
		SCK: machine.GP10,
		SDO: machine.GP11,
		SDI: machine.GP28, // I don't think this is actually used for LCD, just assign to any open pin
	})

	display := st7789.New(machine.SPI1,
		machine.GP12, // TFT_RESET
		machine.GP8,  // TFT_DC
		machine.GP9,  // TFT_CS
		machine.GP13) // TFT_LITE

	display.Configure(st7789.Config{
		// With the display in portrait and the usb socket on the left and in the back
		// the actual width and height are switched width=320 and height=240
		Width:  240,
		Height: 320,
		// Rotation:     st7789.ROTATION_90,
		Rotation:     st7789.NO_ROTATION,
		RowOffset:    0,
		ColumnOffset: 0,
		FrameRate:    st7789.FRAMERATE_111,
		VSyncLines:   st7789.MAX_VSYNC_SCANLINES,
	})

	/////////////////////////////////////////////////////////////////////////////
	// Broker
	/////////////////////////////////////////////////////////////////////////////

	mb, _ := msg.NewBroker(
		machine.UART0,
		machine.UART0_TX_PIN,
		machine.UART0_RX_PIN,
		machine.UART1,
		machine.UART1_RX_PIN,
		machine.UART1_RX_PIN,
	)
	mb.Configure()

	//
	// Create subscription channels
	//
	fooCh := make(chan msg.FooMsg)
	handsetCh := make(chan msg.HandsetMsg)
	raDriverCh := make(chan msg.RADriverMsg)
	raDriverCmdCh := make(chan msg.RADriverCmdMsg)

	//
	// Register the channels with the broker
	//
	mb.SetFooCh(fooCh)
	mb.SetHandsetCh(handsetCh)
	mb.SetRADriverCh(raDriverCh)
	mb.SetRADriverCmdCh(raDriverCmdCh)

	//
	// Create the screen and configure it
	//
	var screen Screen
	screen.ch = make(chan Screen)
	green := color.RGBA{0, 255, 0, 255}
	screen.Configure(10, 20, display, freemono.Regular9pt7b, green)

	// // DEBUG
	// time.Sleep(time.Second * 5)
	// screen.grid.LoadGrid("123\n456\n789")
	// screen.WriteLines()
	// time.Sleep(time.Second * 3)

	//
	// Configure the filter
	//
	screen.Key0 = machine.GP15
	screen.Key1 = machine.GP17
	screen.Key2 = machine.GP2
	screen.Key3 = machine.GP3

	screen.Key0.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	screen.Key1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	screen.Key2.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	screen.Key3.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	screen.Key0.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key0 - RADriverCmd")
		screen.FilterText = "RADriverCmd"
		screen.StatusText = "Filter: RADriverCmd"
		//DEVTODO add a write status method

	})

	screen.Key1.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key1 - RADriver")
		screen.FilterText = "RADriver"
		screen.StatusText = "Filter: RADriver"

	})

	screen.Key2.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key2 - Handset")
		screen.FilterText = "Handset"
		screen.StatusText = "Filter: Handset"

	})

	screen.Key3.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key3 - ANY")
		screen.FilterText = "*ANY"
		screen.StatusText = "Filter: *ANY"

	})

	//
	// Start the routines
	//
	go screen.fooConsumerRoutine(fooCh, mb)
	go screen.handsetConsumerRoutine(handsetCh, mb)
	go screen.raDriverConsumerRoutine(raDriverCh, mb)
	go screen.raDriverCmdConsumerRoutine(raDriverCmdCh, mb)

	//
	// Start the subscription reader, it will read from the the UARTS
	//
	go mb.SubscriptionReaderRoutine()

	/////////////////////////////////////////////////////////////////////////////
	// main Console routine
	/////////////////////////////////////////////////////////////////////////////
	go screen.consoleRoutine()

	//
	// Keep main live
	//
	for {
		time.Sleep(time.Millisecond * 5000)
		fmt.Println("[console-v2.main] heart beat...")
	}
}

func runLight() {

	// run light
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// blink run light for a bit seconds so I can tell it is starting
	for i := 0; i < 5; i++ {
		led.High()
		time.Sleep(time.Millisecond * 100)
		led.Low()
		time.Sleep(time.Millisecond * 100)
	}
	led.High()
}

// DEBUG
func printGrid(grid hid.Grid) {

	for _, row := range grid.GetCells() {
		// fmt.Printf("row: %v\n", r)

		for _, cell := range row {

			fmt.Printf("[%c|%c|%v] \t", cell.GetChar(), cell.GetPrevChar(), cell.IsDirty())
		}
		fmt.Println("--")
	}
}

func (screen *Screen) Configure(rowsCount int, colCount int, displayDevice st7789.Device, font tinyfont.Font, fontColor color.RGBA) {

	screen.ConfigureGrid(rowsCount, colCount)
	screen.displayDevice = displayDevice
	screen.font = font
	screen.fontColor = fontColor

}

func (screen *Screen) WriteLines(text string) {

	screen.LoadGrid(text)

	var x, y int16
	black := color.RGBA{0, 0, 0, 255}

	for r, row := range screen.GetCells() {
		for c, col := range row {
			cell := col
			x = int16(screen.GetWidth()*c) + 5
			y = int16(screen.GetHeight()*r) + 20
			if cell.IsDirty() {
				cells := screen.GetCells()
				tinyfont.WriteLine(&screen.displayDevice, &screen.font, x, y, string(cells[r][c].GetPrevChar()), black) // erase the previous character
				tinyfont.WriteLine(&screen.displayDevice, &screen.font, x, y, string(cells[r][c].GetChar()), screen.fontColor)
			}
		}
	}
}

// Read from fooCh and write to screenCh
func (screen *Screen) fooConsumerRoutine(fooCh chan msg.FooMsg, mb msg.MsgBroker) {

	for msg := range fooCh {

		var bodyText string
		bodyText += fmt.Sprintf("Kind: %v\n", msg.Kind)
		bodyText += fmt.Sprintf("Name: %v", msg.Name)
		screen.BodyText = bodyText
		screen.ch <- *screen

	}
}

// Read from handsetCh and write to screenCh
func (screen *Screen) handsetConsumerRoutine(handsetCh chan msg.HandsetMsg, mb msg.MsgBroker) {

	for msg := range handsetCh {

		var bodyText string
		bodyText += fmt.Sprintf("Kind: %v\n", msg.Kind)
		bodyText += fmt.Sprintf("Keys: %v", msg.Keys)
		screen.BodyText = bodyText
		screen.ch <- *screen

	}
}

// Read from raDriverCh and write to screenCh
func (screen *Screen) raDriverConsumerRoutine(raDriverCh chan msg.RADriverMsg, mb msg.MsgBroker) {

	for msg := range raDriverCh {

		var bodyText string
		bodyText += fmt.Sprintf("Kind: %v\n", msg.Kind)
		bodyText += fmt.Sprintf("Tracking: %v\n", msg.Tracking)
		bodyText += fmt.Sprintf("Direction: %v\n", msg.Direction)
		bodyText += fmt.Sprintf("Position: %v", msg.Position)
		screen.BodyText = bodyText

		fmt.Printf("DEBUG-raDriverConsumerRoutine - %v \n", bodyText)

		screen.ch <- *screen

	}
}

// Read from raDriverCmdCh and write to screenCh
func (screen *Screen) raDriverCmdConsumerRoutine(raDriverCmdCh chan msg.RADriverCmdMsg, mb msg.MsgBroker) {

	for msg := range raDriverCmdCh {
		var bodyText string
		bodyText += fmt.Sprintf("Kind: %v\n", msg.Kind)
		bodyText += fmt.Sprintf("Cmd: %v\n", msg.Cmd)
		bodyText += fmt.Sprintf("Cmd: %v", msg.Cmd)
		screen.BodyText = bodyText
		screen.ch <- *screen
	}

}

func (screen *Screen) consoleRoutine() {

	// red := color.RGBA{255, 0, 0, 255}
	black := color.RGBA{0, 0, 0, 255}
	// white := color.RGBA{255, 255, 255, 255}
	// blue := color.RGBA{0, 0, 255, 255}
	// green := color.RGBA{0, 255, 0, 255}
	// greenDim := color.RGBA{0, 126, 0, 255}

	screen.displayDevice.FillScreen(black)

	for screen := range screen.ch {

		//
		// If no filter or filter text is found
		//
		if screen.FilterText == "*ANY" || strings.Contains(screen.BodyText, screen.FilterText) {

			screen.WriteLines(screen.StatusText + "\n" + screen.BodyText)
			fmt.Printf("DEBUG-1 - %v \t %v\n", screen.StatusText, screen.BodyText)

		}

	}

}
