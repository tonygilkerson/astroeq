package main

import (
	"fmt"
	"machine"
	"strings"

	"time"

	"image/color"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"

	"github.com/tonygilkerson/astroeq/pkg/msg"
)

const SCREEN_MAX_LINES int = 14

type DisplayFilter struct {
	Txt     string // Foo | Handset | RADriver | RADriverCmd
	PrevTxt string
	Key0    machine.Pin
	Key1    machine.Pin
	Key2    machine.Pin
	Key3    machine.Pin
}

type Screen struct {
	Lines     [SCREEN_MAX_LINES]string
	PrevLines [SCREEN_MAX_LINES]string
	LinePtr   int // Index used when adding a new line
	Filter    DisplayFilter
	display   st7789.Device
	ch        chan Screen
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
	// Create the screen
	//
	var screen Screen
	screen.ch = make(chan Screen)
	screen.display = display

	//
	// Configure the filter
	//
	screen.Filter.Key0 = machine.GP15
	screen.Filter.Key1 = machine.GP17
	screen.Filter.Key2 = machine.GP2
	screen.Filter.Key3 = machine.GP3

	screen.Filter.Key0.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	screen.Filter.Key1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	screen.Filter.Key2.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	screen.Filter.Key3.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	screen.Filter.Key0.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key0 - RADriverCmd")
		screen.Filter.Txt = "RADriverCmd"
		screen.ClearLines()
		screen.WriteStatus()
	})

	screen.Filter.Key1.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key1 - RADriver")
		screen.Filter.Txt = "RADriver"
		screen.ClearLines()
		screen.WriteStatus()
	})

	screen.Filter.Key2.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key2 - Handset")
		screen.Filter.Txt = "Handset"
		screen.ClearLines()
		screen.WriteStatus()
	})

	screen.Filter.Key3.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key3 - ANY")
		screen.Filter.Txt = "*ANY"
		screen.ClearLines()
		screen.WriteStatus()
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
	// maine Console routine
	/////////////////////////////////////////////////////////////////////////////
	go screen.consoleRoutine()

	//
	// Keep main live
	//
	for {
		time.Sleep(time.Millisecond * 5000)
		fmt.Println("[console.main] heart beat...")
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

// Read from fooCh and write to screenCh
func (screen *Screen) fooConsumerRoutine(fooCh chan msg.FooMsg, mb msg.MsgBroker) {

	for msg := range fooCh {

		screen.StartLines()
		screen.SetLine(fmt.Sprintf("Kind: %v", msg.Kind))
		screen.SetLine(fmt.Sprintf("Name: %v", msg.Name))
		screen.EndLines()
		screen.ch <- *screen

	}
}

// Read from handsetCh and write to screenCh
func (screen *Screen) handsetConsumerRoutine(handsetCh chan msg.HandsetMsg, mb msg.MsgBroker) {

	for msg := range handsetCh {

		screen.StartLines()
		screen.SetLine(fmt.Sprintf("Kind: %v", msg.Kind))
		screen.SetLine(fmt.Sprintf("Keys: %v", msg.Keys))
		screen.EndLines()
		screen.ch <- *screen

	}
}

// Read from raDriverCh and write to screenCh
func (screen *Screen) raDriverConsumerRoutine(raDriverCh chan msg.RADriverMsg, mb msg.MsgBroker) {

	for msg := range raDriverCh {

		screen.StartLines()
		screen.SetLine(fmt.Sprintf("Kind: %v", msg.Kind))
		screen.SetLine(fmt.Sprintf("Tracking: %v", msg.Tracking))
		screen.SetLine(fmt.Sprintf("Direction: %v", msg.Direction))
		screen.SetLine(fmt.Sprintf("Position: %v", msg.Position))
		screen.EndLines()
		screen.ch <- *screen

	}
}

// Read from raDriverCmdCh and write to screenCh
func (screen *Screen) raDriverCmdConsumerRoutine(raDriverCmdCh chan msg.RADriverCmdMsg, mb msg.MsgBroker) {

	for msg := range raDriverCmdCh {
		screen.StartLines()
		screen.SetLine(fmt.Sprintf("Kind: %v", msg.Kind))
		screen.SetLine(fmt.Sprintf("Cmd: %v", msg.Cmd))
		screen.SetLine(fmt.Sprintf("Args: %v", msg.Args))
		screen.EndLines()
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

	screen.display.FillScreen(black)

	for screen := range screen.ch {

		//
		// If no filter or filter text is found
		//
		if screen.Filter.Txt == "*ANY" || screen.SearchLines(screen.Filter.Txt) {

			screen.WriteLines()

		}

	}

}

func (screen *Screen) SetLine(txt string) {

	screen.PrevLines[screen.LinePtr] = screen.Lines[screen.LinePtr]
	screen.Lines[screen.LinePtr] = txt
	screen.LinePtr++

}

func (screen *Screen) WriteStatus() {
	yellow := color.RGBA{255, 255, 0, 255}
	black := color.RGBA{0, 0, 0, 255}
	var x, y int16 = 5, 20
	screen.display.FillRectangle(x, y-15, 234, 20, black)

	var status string
	status = fmt.Sprintf("F: %v ch: %v", screen.Filter.Txt, len(screen.ch))
	tinyfont.WriteLine(&screen.display, &freemono.Regular9pt7b, x, y, status, yellow)

}

func (screen *Screen) SearchLines(txt string) bool {

	for i := 0; i < SCREEN_MAX_LINES; i++ {

		if len(screen.Lines[i]) > 0 && len(txt) > 0 && strings.Contains(screen.Lines[i], txt) {
			return true
		}
	}

	return false
}

func (screen *Screen) WriteLines() {

	green := color.RGBA{0, 255, 0, 255}
	//red := color.RGBA{255, 0, 0, 20}
	black := color.RGBA{0, 0, 0, 255}

	var x, y int16
	// tinyfont.WriteLine(display, &freemono.Regular9pt7b, 5, 20, screen.Lines[0], green)
	// display.FillRectangle(x,y-15,234,20,red)

	for i := 0; i < SCREEN_MAX_LINES; i++ {
		if screen.Lines[i] != screen.PrevLines[i] {
			x = 5
			y = int16((20 * i) + 40)
			screen.display.FillRectangle(x, y-15, 234, 20, black)
			tinyfont.WriteLine(&screen.display, &freemono.Regular9pt7b, x, y, screen.Lines[i], green)
		}
	}

}

func (screen *Screen) ClearLines() {
	screen.LinePtr = 0
	screen.EndLines()
}

func (screen *Screen) StartLines() {
	screen.LinePtr = 0
}

func (screen *Screen) EndLines() {

	for i := screen.LinePtr; i < SCREEN_MAX_LINES; i++ {
		screen.SetLine("")
	}
}
