package main

import (
	"fmt"

	"github.com/tonygilkerson/astroeq/pkg/driver"
	"github.com/tonygilkerson/astroeq/pkg/msg"

	"machine"
	"math"
	"time"
)

// See wire.md for wiring details and pin assignments

func main() {

	// run light
	runLight()

	/////////////////////////////////////////////////////////////////////////////
	// Broker
	/////////////////////////////////////////////////////////////////////////////

	fmt.Println("[main] Create new broker")

	//
	// UART0
	//
	machine.UART0.Configure(machine.UARTConfig{
		TX: machine.UART0_TX_PIN,
		RX: machine.UART0_RX_PIN,
	})

	var uartUp msg.UART
	var uartUpTxPin machine.Pin
	var uartUpRxPin machine.Pin

	uartUp = machine.UART0
	uartUpTxPin = machine.UART0_TX_PIN
	uartUpRxPin = machine.UART0_RX_PIN

	//
	// UART1
	//
	machine.UART1.Configure(machine.UARTConfig{
		TX: machine.GP4,
		RX: machine.GP5,
	})

	var uartDn msg.UART
	var uartDnTxPin machine.Pin
	var uartDnRxPin machine.Pin

	uartDn = machine.UART1
	uartDnTxPin = machine.GP4
	uartDnRxPin = machine.GP5

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
	//
	// Create subscription channels and
	// Register the them with the broker
	//
	fooCh := make(chan msg.FooMsg)
	mb.SetFooCh(fooCh)

	// DEVTODO - delete me soon if not needed, I dont think the ra-driver needs to consume this message only publish it
	// raDriverCh := make(chan msg.RADriverMsg)
	// mb.SetRADriverCh(raDriverCh)

	raDriverCmdCh := make(chan msg.RADriverCmdMsg)
	mb.SetRADriverCmdCh(raDriverCmdCh)
	//
	// Start the subscription reader, it will read from the the UARTS
	// and then dispatch message to the proper channels
	//
	go mb.SubscriptionReaderRoutine()

	/////////////////////////////////////////////////////////////////////////////
	// RA-Drive
	/////////////////////////////////////////////////////////////////////////////

	//
	// Configure SPI bus
	//
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 115200,
		LSBFirst:  false,
		Mode:      0,
		DataBits:  8,
		SCK:       machine.SPI0_SCK_PIN, // GP18
		SDO:       machine.SPI0_SDO_PIN, // GP19
		SDI:       machine.SPI0_SDI_PIN, // GP16
	})

	//
	// motor
	//

	// Select the hardware PWM for the RA Driver
	var raPWM driver.PWM
	raPWM = machine.PWM4

	// Direction North or South
	raDirectionPin := machine.GP8

	// Enable motor
	raEnableMotorPin := machine.GP13

	raStep := machine.GP9
	var raStepsPerRevolution int32 = 400
	var raMaxHz int32 = 1000
	var raMaxMicroStepSetting driver.MicroStep = 16
	var raWormRatio int32 = 144
	var raGearRatio int32 = 3
	raMicroStep1 := machine.GP12
	raMicroStep2 := machine.GP11
	raEncoderSPI := *machine.SPI0
	raEncoderCS := machine.GP20
	ra, _ := driver.NewRADriver(
		raStep,
		raPWM,
		raDirectionPin,
		raStepsPerRevolution,
		raMaxHz,
		raMicroStep1,
		raMicroStep2,
		raMaxMicroStepSetting,
		raEnableMotorPin,
		raWormRatio,
		raGearRatio,
		raEncoderSPI,
		raEncoderCS,
	)
	ra.Configure()
	//ra.RunAtHz(700.0)
	//ra.RunAtHz(300.0)
	//ra.RunAtHz(200.0)
	// ra.RunAtHz(50.0)
	//ra.RunAtHz(40.0)
	ra.RunAtSiderealRate()

	//
	// Start the message consumers
	//
	go fooConsumerRoutine(fooCh, &mb)
	go raCmdConsumeRoutine(raDriverCmdCh, &mb, &ra)
	go raPublishInfoRoutine(&ra, &mb)

	var position uint32 = 0
	var lastPosition int = 0

	//
	// Track by the second
	//

	for i := 0; i < 6000; i++ {

		position = ra.GetPosition()

		pos := int(position)
		perSec := math.Abs(float64(pos - lastPosition))

		fmt.Printf("[main] position: %v, per sec: %.2f (81.92 expected))\n", position, perSec)
		lastPosition = pos
		time.Sleep(time.Millisecond * 1000)

		//
		// Testing to see if I can count one RA rotation
		//
		// The motor and encoder rotate together so one full turn of the motor is one full turn of the encoder
		// The encoder positions are from 0 to 2^14 (16_384)
		// So we should be able to just multiple by the gear ratios:
		// 16_384 (1 motor turn) * 3 (main gear) * 144 (worm gear) = 7_077_888
		// if position >= 7_077_888 {
		// 	break
		// }

	}

	fmt.Println("[main] Reset RA and track by min...")
	ra.ZeroRA() // DEVTODO does not seem to work, make sure I am clearing the rotation count as well
	time.Sleep(time.Millisecond * 5000)

	//
	// Track for a few min
	for i := 0; i < 15; i++ {

		position = ra.GetPosition()

		pos := int(position)
		perMin := math.Abs(float64(pos - lastPosition))

		fmt.Printf("[main] position: %v, per min: %.2f (4915.2 expected))\n", position, perMin)
		lastPosition = pos
		time.Sleep(time.Millisecond * 60000)

	}

	// Done
	println("[main] Done!")

}

func runLight() {

	// run light
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// blink run light for a bit seconds so I can tell it is starting
	for i := 0; i < 25; i++ {
		led.High()
		time.Sleep(time.Millisecond * 100)
		led.Low()
		time.Sleep(time.Millisecond * 100)
	}
	led.High()
}

func fooConsumerRoutine(ch chan msg.FooMsg, mb *msg.MsgBroker) {

	for foo := range ch {
		fmt.Printf("[ra-driver.fooConsumerRoutine] - Kind: [%s], Name: [%s]\n", foo.Kind, foo.Name)
	}
}

func raCmdConsumeRoutine(ch chan msg.RADriverCmdMsg, mb *msg.MsgBroker, ra *driver.RADriver) {

	for raCmdMsg := range ch {
		fmt.Printf("[raCmdConsumeRoutine] - raCmdMsg: [%v]\n", raCmdMsg)
		raDriverCtl(raCmdMsg, ra)
	}

}

func raDriverCtl(cmdMsg msg.RADriverCmdMsg, ra *driver.RADriver) {

	switch cmdMsg.Cmd {

	case msg.RA_CMD_SET_DIRECTION:
		// The first argument is the direction "North" or "South"
		ra.SetDirection(driver.RaValue(cmdMsg.Args[0]))

	case msg.RA_CMD_SET_TRACKING:
		// The first argument is the tracking "On" or "Off"
		ra.SetTracking(driver.RaValue(cmdMsg.Args[0]))

	}
}

func raPublishInfoRoutine(ra *driver.RADriver, mb *msg.MsgBroker) {

	for {
		var raMsg msg.RADriverMsg
		raMsg.Kind = msg.MSG_RADRIVER
		raMsg.Tracking = ra.GetTracking()
		raMsg.Direction = ra.GetDirection()
		raMsg.Position = ra.GetPosition()

		mb.PublishRADriver(raMsg)

		time.Sleep(time.Second * 2)
	}
}
