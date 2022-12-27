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
	Filter string // Foo | Handset | RADriver | RADriverCmd
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

	df.Key0.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key0")
		df.Filter = "RADriverCmd"
	})

	df.Key1.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key1")
		df.Filter = "RADriver"
	})

	df.Key2.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key2")
		df.Filter = "Handset"
	})

	df.Key3.SetInterrupt(machine.PinFalling, func(p machine.Pin) {
		fmt.Println("key3")
		df.Filter = ""
	})

	/////////////////////////////////////////////////////////////////////////////
	// Console Display
	/////////////////////////////////////////////////////////////////////////////

	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 9_000_000,
		LSBFirst:  false,
		Mode:      0,
		DataBits:  0,
		SCK:       machine.GP10,
		SDO:       machine.GP11,
		SDI:       machine.GP28, // I don't think this is actually used for LCD, just assign to any open pin
	})

	// machine.SPI1.SetBaudRate(60_000_000)

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
		// FrameRate:    st7789.FRAMERATE_111,
		// VSyncLines:   st7789.MAX_VSYNC_SCANLINES,

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
	raDriverCmdCh := make(chan msg.RADriverCmdMsg)

	//
	// Register the channels with the broker
	//
	mb.SetFooCh(fooCh)
	mb.SetHandsetCh(handsetCh)
	mb.SetRADriverCh(raDriverCh)
	mb.SetRADriverCmdCh(raDriverCmdCh)

	//
	// Start the message consumers
	//
	go fooConsumerRoutine(fooCh, mb, consoleCh)
	go handsetConsumerRoutine(handsetCh, mb, consoleCh)
	go raDriverConsumerRoutine(raDriverCh, mb, consoleCh)
	go raDriverCmdConsumerRoutine(raDriverCmdCh, mb, consoleCh)

	//
	// Start the subscription reader, it will read from the the UARTS
	//
	go mb.SubscriptionReaderRoutine()

	/////////////////////////////////////////////////////////////////////////////
	// writeConsole
	/////////////////////////////////////////////////////////////////////////////

	go consoleRoutine(display, consoleCh, &df)
	go checkFilterRoutine(consoleCh, &df)

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
		txt := fmt.Sprintf("Kind: %s\nName: %s", msg.Kind, msg.Name)
		consoleCh <- txt
	}
}

// Read from handsetCh and write to consoleCh
func handsetConsumerRoutine(handsetCh chan msg.HandsetMsg, mb msg.MsgBroker, consoleCh chan string) {

	for msg := range handsetCh {
		txt := fmt.Sprintf("Kind: %v\nKeys: %v", msg.Kind, msg.Keys)
		consoleCh <- txt
	}
}

// Read from raDriverCh and write to consoleCh
func raDriverConsumerRoutine(raDriverCh chan msg.RADriverMsg, mb msg.MsgBroker, consoleCh chan string) {

	for msg := range raDriverCh {
		txt := fmt.Sprintf("Kind: %v\nTracking: %v\nDirection: %v\nPosition: %v", msg.Kind, msg.Tracking, msg.Direction, msg.Position)
		consoleCh <- txt
	}
}

// Read from raDriverCmdCh and write to consoleCh
func raDriverCmdConsumerRoutine(raDriverCmdCh chan msg.RADriverCmdMsg, mb msg.MsgBroker, consoleCh chan string) {

	for msg := range raDriverCmdCh {
		txt := fmt.Sprintf("Kind: %v\nCmd: %v\nArgs: %v", msg.Kind, msg.Cmd, msg.Args)
		consoleCh <- txt
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

	var lastMsg string = ""
	var lastMsgTime time.Time = time.Now()

	for msg := range ch {

		//
		// If no filter or if filter text if found
		//
		if len(df.Filter) == 0 || strings.Contains(msg, df.Filter) {

			//
			// Don't redraw screen if the message is the same
			// Also the messages come in faster than the screen can redraw so only redraw every so often
			// This means that not all messages get displayed
			//
			if msg != lastMsg && (time.Since(lastMsgTime) > time.Duration(time.Second*3)) {
				lastMsgTime = time.Now()

				cls(&display)

				// txt := fmt.Sprintf("Filter: [%v]\n%v",df.Filter,lastMsg)
				// tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 20, txt, black)

				txt := fmt.Sprintf("Filter: [%v]\n%v", df.Filter, msg)
				tinyfont.WriteLine(&display, &freemono.Regular12pt7b, 5, 20, txt, green)

				// pause for bit so I can see the screen before it refreshes
				time.Sleep(time.Millisecond * 1000)
			}

		}

		lastMsg = msg

	}

}

func checkFilterRoutine(ch chan string, df *DisplayFilter) {

	var lastFilter string = ""

	for {
		if df.Filter != lastFilter {
			ch <- "Filter changed to:\n" + df.Filter
		}

		lastFilter = df.Filter
		time.Sleep(time.Millisecond * 500)
	}
}
