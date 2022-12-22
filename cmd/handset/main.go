package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"github.com/tonygilkerson/astroeq/pkg/hid"
	"github.com/tonygilkerson/astroeq/pkg/msg"

	"tinygo.org/x/drivers/ssd1351"
)

/*


	Pico									OLED 																						ssd1351 						keypad						UART
	---------------------	-------------------------------------------			-----------------		----------------	-------------
	3v3										VCC
	GP0 																																																				UART0 TX
	GP1 																																																				UART0 RX
	GP2 																																											scrollDnKey
	GP3 																																											zeroKey
	GP4 																																											scrollKEY_UP
	GP5 																																											sevenKey
	GP6 																																											eightKey
	GP7 																																											nineKey
	GP8 																																											fourKey
	GP9 																																											fiveKey
	GP10 																																											sixKey
	GP11 																																											oneKey
	GP12 																																											twoKey
	GP13 																																											threeKey
	GP14 																																											rightKey
	GP15 																																											leftKey
	GP16 									SPI0_SDI_PIN (used for a hack, see commends in code below)
	GP17 																																											downKey
	GP18 									CLK	- clock input (SPI0_SCK_PIN)
	GP19 									DIN	- data in     (SPI0_SDO_PIN)
	GP20 																																											enterKey
	GP21																																											upKey (move from 16)
	GP22																																											escKey   (move from 18)
	GP26																																											setupKey (move from 19)
	GP27									CS 	- Chip select																csPin
	GP28									DC	- Data/Cmd (high=data,low=cmd)							dcPin
												RST	WHT	- Reset (low=active)										resetPin
																																				enPin
			 																																	rwPin
																																				bus (machine.SPI0)
												https://www.waveshare.com/product/displays/oled/pico-oled-2.23.htm
*/

func main() {

	// run light
	runLight()

	/////////////////////////////////////////////////////////////////////////////
	// Broker
	/////////////////////////////////////////////////////////////////////////////

	fmt.Println("Create new broker")

	machine.UART0.Configure(machine.UARTConfig{
		TX: machine.UART0_TX_PIN,
		RX: machine.UART0_RX_PIN,
	})

	var uartUp msg.UART
	var uartUpTxPin machine.Pin
	var uartUpRxPin machine.Pin

	var uartDn msg.UART
	var uartDnTxPin machine.Pin
	var uartDnRxPin machine.Pin

	uartUp = machine.UART0
	uartUpTxPin = machine.UART0_TX_PIN
	uartUpRxPin = machine.UART0_RX_PIN

	// Note if UART1 was use it would be used here, however
	// the Handset is at the head of the conga line so no UART1 needed

	mb, err := msg.NewBroker(
		uartUp,
		uartUpTxPin,
		uartUpRxPin,
		uartDn,
		uartDnTxPin,
		uartDnRxPin,
	)

	if err != nil {
		fmt.Println(err)
		return
	}
	mb.Configure()

	//
	// Create subscription channels and
	// Register the them with the broker
	//
	fooCh := make(chan msg.FooMsg)
	mb.SetFooCh(fooCh)

	raDriverCh := make(chan msg.RADriverMsg)
	mb.SetRADriverCh(raDriverCh)

	handsetCh := make(chan msg.HandsetMsg)
	mb.SetHandsetCh(handsetCh)

	//
	// Start the subscription reader, it will read from the the UARTS
	// and dispatch to the proper channel
	//
	go mb.SubscriptionReaderRoutine()

	/////////////////////////////////////////////////////////////////////////////
	// Display
	/////////////////////////////////////////////////////////////////////////////
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 5_760_000,
		LSBFirst:  false,
		Mode:      0,
		DataBits:  8,
		SCK:       machine.SPI0_SCK_PIN, // GP18
		SDO:       machine.SPI0_SDO_PIN, // GP19
		SDI:       machine.SPI0_SDI_PIN, // GP16
	})

	var rst machine.Pin // ran out of pins
	dc := machine.Pin(28)
	cs := machine.Pin(27)
	var en machine.Pin // ran out of pins
	var rw machine.Pin // ran out of pins

	// HACK - I ran out of pins and so I used GP16 for rst, en and rw
	//        I am not sure what this does in the ssd1351 driver but the
	//        display functions that I need are working for now but this
	//        might be an issue in the future
	rst = machine.GP16
	en = machine.GP16
	rw = machine.GP16

	display := ssd1351.New(machine.SPI0, rst, dc, cs, en, rw)
	display.Configure(ssd1351.Config{
		Width:        128,
		Height:       128,
		RowOffset:    0,
		ColumnOffset: 0,
	})

	// not sure if this is needed
	display.Command(ssd1351.SET_REMAP_COLORDEPTH)
	display.Data(0x62)

	display.FillScreen(color.RGBA{0, 0, 0, 0})

	//
	// keypad keys
	//
	zeroKey := machine.GP3
	oneKey := machine.GP11
	twoKey := machine.GP12
	threeKey := machine.GP13

	fourKey := machine.GP8
	fiveKey := machine.GP9
	sixKey := machine.GP10
	sevenKey := machine.GP5
	eightKey := machine.GP6
	nineKey := machine.GP7

	scrollKEY_UP := machine.GP4
	scrollDnKey := machine.GP2

	rightKey := machine.GP14
	leftKey := machine.GP15
	upKey := machine.GP21
	downKey := machine.GP17

	escKey := machine.GP22
	setupKey := machine.GP26
	enterKey := machine.GP20

	handset, _ := hid.NewHandset(
		&display,
		&mb,
		zeroKey,
		oneKey,
		twoKey,
		threeKey,
		fourKey,
		fiveKey,
		sixKey,
		sevenKey,
		eightKey,
		nineKey,
		scrollKEY_UP,
		scrollDnKey,
		rightKey,
		leftKey,
		upKey,
		downKey,
		escKey,
		setupKey,
		enterKey)

	keyStrokesCh := handset.Configure()

	//
	// Start the message consumers
	//
	go fooConsumerRoutine(fooCh, &mb)
	go raDriverConsumerRoutine(&handset, raDriverCh, &mb)

	//
	// Start the local key consumer
	//
	go handsetStateMachineRoutine(&handset, keyStrokesCh, &mb)

	//
	// Keep main live
	//
	for {
		time.Sleep(time.Millisecond * 5000)
		fmt.Println("[Handset.main] heart beat...")
	}

}

func runLight() {

	// run light
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// blink run light for a bit seconds so I can tell it is starting
	for i := 0; i < 10; i++ {
		led.High()
		time.Sleep(time.Millisecond * 100)
		led.Low()
		time.Sleep(time.Millisecond * 100)
	}
	led.High()
}

func handsetStateMachineRoutine(hs *hid.Handset, keyStrokesCh chan hid.Key, mb *msg.MsgBroker) {

	// status bar

	// Menu interaction
	var noKey hid.Key

	hs.Screen.BodyText = hs.StateMachine(noKey)
	hs.Screen.PrevBodyText = ""

	hs.RenderScreen()

	for k := range keyStrokesCh {

		// DEVTODO - create a SetBodyText that set the prev for you, hide this crap
		hs.Screen.PrevBodyText = hs.Screen.BodyText
		hs.Screen.BodyText = hs.StateMachine(k)
		hs.RenderScreen()

	}

}

func fooConsumerRoutine(ch chan msg.FooMsg, mb *msg.MsgBroker) {

	for foo := range ch {
		fmt.Printf("[handset.fooConsumerRoutine] - Kind: [%s], name: [%s]\n", foo.Kind, foo.Name)
	}
}

func raDriverConsumerRoutine(hs *hid.Handset, ch chan msg.RADriverMsg, mb *msg.MsgBroker) {

	for raMsg := range ch {
		fmt.Printf("[handset.raDriverConsumerRoutine] - Kind: [%s], Cmd: [%s]\n", raMsg.Kind, raMsg.Cmd)
		fmt.Printf("[handset.raDriverConsumerRoutine] - msg: [%v]\n", raMsg)

		// We are only interested in raDriver info messages
		if raMsg.Cmd != msg.RA_CMD_INFO {
			continue
		}

		hs.Screen.Tracking = raMsg.Tracking
		hs.Screen.Direction = raMsg.Direction
		hs.Screen.Position = raMsg.Position

		// DEVTODO - make this a render status not the entire screen
		hs.RenderScreen()

	}
}
