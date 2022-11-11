package main

import (
	"fmt"
	"machine"
	"time"

	"image/color"

	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"

	"github.com/tonygilkerson/astroeq/pkg/msg"
)

func main() {

	

	// run light
	runLight()

	
	/////////////////////////////////////////////////////////////////////////////
	// Console Display
	/////////////////////////////////////////////////////////////////////////////

	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 8000000,
		LSBFirst:  false,
		Mode:      0,
		DataBits:  0,
		SCK:       machine.GP10,
		SDO:       machine.GP11,
		SDI:       machine.GP28, // I don't think this is actually used for LCD, just assign to any open pin
	})

	display := st7789.New(machine.SPI1,
		machine.GP12, // TFT_RESET
		machine.GP8, // TFT_DC
		machine.GP9, // TFT_CS
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
	

	//
	// Register the channels with the broker
	//
	mb.SetFooCh(fooCh)


	//
	// Start the message consumers
	//
	go fooConsumer(fooCh, mb, consoleCh)

	//
	// Start the subscription reader, it will read from the the UARTS
	//
	go mb.SubscriptionReader()


	/////////////////////////////////////////////////////////////////////////////
	// writeConsole
	/////////////////////////////////////////////////////////////////////////////
	
	writeConsole(display,consoleCh)
}


func runLight() {

	// run light
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// blink run light for a bit seconds so I can tell it is starting
	for i := 0; i < 15; i++ {
		led.High()
		time.Sleep(time.Millisecond * 100)
		led.Low()
		time.Sleep(time.Millisecond * 100)
	}
	led.High()
}

func paintScreen(c color.RGBA, d *st7789.Device, s int16) {
	var x,y int16
	for y = 0; y < 240; y=y+s {
		for x = 0; x < 320; x=x+s {
			d.FillRectangle(x, y, s, s, c)
		}
	}
}

func cls (d *st7789.Device){
	black := color.RGBA{0, 0, 0, 255}
	d.FillScreen(black)
	fmt.Printf("FillScreen(black)\n")
}

//
// Read from fooCh and write to consoleCh
//
func fooConsumer(fooCh chan msg.FooMsg, mb msg.MsgBroker, consoleCh chan string) {

	for msg := range fooCh {
		s := fmt.Sprintf("%s: %s", msg.Kind, msg.Name)
		consoleCh <- s
	}
}

func writeConsole(display st7789.Device, ch chan string) {

	width, height := display.Size()
	fmt.Printf("width: %v, height: %v\n",width, height)

	red := color.RGBA{126, 0, 0, 255}
	// red := color.RGBA{255, 0, 0, 255}
	// black := color.RGBA{0, 0, 0, 255}
	// white := color.RGBA{255, 255, 255, 255}
	// blue := color.RGBA{0, 0, 255, 255}
	// green := color.RGBA{0, 255, 0, 255}

	cls(&display)
	tinyfont.WriteLine(&display,&freemono.Regular12pt7b,10,20,"Ready...",red)

	for msg := range ch {
		
		cls(&display)
		// tinyfont.WriteLine(&display,&freemono.Regular12pt7b,10,20,"123456789-123456789-x",red)
		tinyfont.WriteLine(&display,&freemono.Regular12pt7b,10,20,msg,red)
		
	}

}