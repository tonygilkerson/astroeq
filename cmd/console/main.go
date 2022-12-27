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

type DisplayFilter struct {
	Filter string // Foo | Handset | RADriver
	Key0   machine.Pin
	Key1   machine.Pin
	Key2   machine.Pin
	Key3   machine.Pin
}

func main() {

	// run light
	runLight()

	///////////////////////////////////////////////
	// DEBUG -
	///////////////////////////////////////////////
	// key0 := machine.GP15
	// key0.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	// key1 := machine.GP17
	// key1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	// for {

	// 	fmt.Printf("DEBUG test key0: %v  key1: %v\n",key0.Get(), key1.Get())
	// 	time.Sleep(time.Millisecond * 1000)
	// }
	///////////////////////////////////////////////
	// DEBUG -
	///////////////////////////////////////////////

	//
	// Configure the filter
	//
	var df DisplayFilter
	df.Key0 = machine.GP15
	df.Key1 = machine.GP17
	df.Key2 = machine.GP2
	df.Key3 = machine.GP3

	df.Key0.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	df.Key1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	df.Key2.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	df.Key3.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	df.Key0.SetInterrupt(machine.PinFalling, func(p machine.Pin) { fmt.Println("key0"); df.Filter = "Handset" })
	df.Key1.SetInterrupt(machine.PinFalling, func(p machine.Pin) { fmt.Println("key1"); df.Filter = "RADriver" })
	df.Key2.SetInterrupt(machine.PinFalling, func(p machine.Pin) { fmt.Println("key2"); df.Filter = "XXXXX" })
	df.Key3.SetInterrupt(machine.PinFalling, func(p machine.Pin) { fmt.Println("key3"); df.Filter = "XXXXX" })

	/////////////////////////////////////////////////////////////////////////////
	// Console Display
	/////////////////////////////////////////////////////////////////////////////

	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 8_000_000,
		LSBFirst:  false,
		Mode:      0,
		DataBits:  0,
		SCK:       machine.GP10,
		SDO:       machine.GP11,
		SDI:       machine.GP28, // I don't think this is actually used for LCD, just assign to any open pin
	})

	display := st7789.New(machine.SPI1,
		machine.GP12, // TFT_RESET
		machine.GP8,  // TFT_DC
		machine.GP9,  // TFT_CS
		machine.GP13) // TFT_LITE

	display.Configure(st7789.Config{
		// With the display in portrait and the usb socket on the left and in the back
		// the actual width and height are switched width=320 and height=240
		Width:        240,
		Height:       320,
		Rotation:     st7789.ROTATION_90,
		RowOffset:    0,
		ColumnOffset: 0,
		FrameRate:    st7789.FRAMERATE_111,
		VSyncLines:   st7789.MAX_VSYNC_SCANLINES,
	})

	consoleCh := make(chan string)

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

	//
	// Register the channels with the broker
	//
	mb.SetFooCh(fooCh)
	mb.SetHandsetCh(handsetCh)
	mb.SetRADriverCh(raDriverCh)

	//
	// Start the message consumers
	//
	go fooConsumerRoutine(fooCh, mb, consoleCh)
	go handsetConsumerRoutine(handsetCh, mb, consoleCh)
	go raDriverConsumerRoutine(raDriverCh, mb, consoleCh)

	//
	// Start the subscription reader, it will read from the the UARTS
	//
	go mb.SubscriptionReaderRoutine()

	/////////////////////////////////////////////////////////////////////////////
	// writeConsole
	/////////////////////////////////////////////////////////////////////////////

	go consoleRoutine(display, consoleCh, &df)

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

func paintScreen(c color.RGBA, d *st7789.Device, s int16) {
	var x, y int16
	for y = 0; y < 240; y = y + s {
		for x = 0; x < 320; x = x + s {
			d.FillRectangle(x, y, s, s, c)
		}
	}
}

func cls(d *st7789.Device) {
	// green := color.RGBA{0, 255, 0, 255}
	black := color.RGBA{0, 0, 0, 255}
	d.FillScreen(black)
}

// Read from fooCh and write to consoleCh
func fooConsumerRoutine(fooCh chan msg.FooMsg, mb msg.MsgBroker, consoleCh chan string) {

	for msg := range fooCh {
		s := fmt.Sprintf("%s: %s", msg.Kind, msg.Name)
		consoleCh <- s
	}
}

// Read from handsetCh and write to consoleCh
func handsetConsumerRoutine(handsetCh chan msg.HandsetMsg, mb msg.MsgBroker, consoleCh chan string) {

	for msg := range handsetCh {
		s := fmt.Sprintf("%v\n", msg)
		consoleCh <- s
	}
}

// Read from raDriverCh and write to consoleCh
func raDriverConsumerRoutine(raDriverCh chan msg.RADriverMsg, mb msg.MsgBroker, consoleCh chan string) {

	for msg := range raDriverCh {
		var s string = string(msg.Kind) + "\n"
		s += string(msg.Tracking) + "\n"
		s += string(msg.Direction) + "\n"
		s += string(msg.Position)

		consoleCh <- s
	}
}

func consoleRoutine(display st7789.Device, ch chan string, df *DisplayFilter) {
	width, height := display.Size()
	fmt.Printf("width: %v, height: %v\n", width, height)

	//red := color.RGBA{126, 0, 0, 255}
	red := color.RGBA{255, 0, 0, 255}
	// black := color.RGBA{0, 0, 0, 255}
	// white := color.RGBA{255, 255, 255, 255}
	// blue := color.RGBA{0, 0, 255, 255}
	green := color.RGBA{0, 255, 0, 255}
	// greenDim := color.RGBA{0, 126, 0, 255}

	cls(&display)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 20, "123456789-123456789-x", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 40, "Ready...2", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 60, "Ready...3", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 80, "Ready...4", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 100, "Ready...5", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 120, "Ready...6", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 140, "Ready...7", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 160, "Ready...8", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 180, "Ready...9", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 200, "Ready...10", red)
	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 220, "Ready...11", red)

	time.Sleep(time.Millisecond * 3000)
	fmt.Printf("[consoleRoutine] - startup...\n")
	// var l int16 = 1
	// var first bool = true
	// var toggle bool = true
	// var vline int16 = 0

	// for msg := range ch {
	// 	if first {
	// 		cls(&display)
	// 		first = false
	// 	}

	// 	if l > 9 {
	// 		display.DrawFastHLine(0, 300, vline+7, black) // erase the last line
	// 		l = 1
	// 	}

	// 	vline = int16(l * 25)
	// 	if toggle {
	// 		toggle = false
	// 		display.FillRectangle(0, vline-20, 320, 25, black)
	// 	} else {
	// 		toggle = true
	// 		display.FillRectangle(0, vline-20, 320, 25, black)
	// 	}

	// 	tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, vline, msg, red)
	// 	fmt.Println("vline: ", vline)
	// 	display.DrawFastHLine(0, 300, vline+7, greenDim)

	// 	l++
	// }

	var lastMsg string = ""

	for msg := range ch {
		if strings.Contains(msg, df.Filter) {
			if msg != lastMsg {
				cls(&display)
				tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 20, msg, green)
			}
			fmt.Printf("[consoleRoutine] - msg:\n%v\n%v\n", msg, df.Filter)
		} else {
			fmt.Printf("[consoleRoutine] - msg:\n%v\nDOES NOT MATCH FILTER\n", msg, df.Filter)
		}

		lastMsg = msg
	}

}
