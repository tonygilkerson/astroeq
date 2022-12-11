package main

import (
	"fmt"

	"github.com/tonygilkerson/astroeq/pkg/driver"

	"machine"
	"time"
	"math"
)

/*

  OLED                                            Pico                               ssd1351 parms           AMT223B-V       CAT5
  ----------------------------------------------  ---------------------------------  ----------------------  --------------- -------
                                                  machine.SPI0                       bus
  VCC                                             VBUS 5v                                                    Pin1 VCC  RED   1 ORN-s
  GND                                             GND                                                        Pin4 GND  BLK   3 BLU
  DIN  BLU - data in                              GP19, SPI0_SDO_PIN                                         Pin3 MOSI ORN   5 ORN
                                                  GP16, SPI0_SDI_PIN                                         Pin5 MISO GRN   2 GRN
  CLK  YLW  - clock data in                       GP18, SPI0_SCK_PIN                                         Pin2 SCLK BRN   4 BRN
  CS   ORN - Chip select                          GP17                                csPin
  DC   GRN - Data/Cmd select (high=data,low=cmd)  GP22 (any open pin)                 dcPin
  RST  WHT  - Reset (low=active)                  GP26 (any open pin)                 resetPin
                                                  GP27 (any open pin)                 enPin
                                                  GP28 (any open pin)                 rwPin
                                                  GP20                                                       Pin6 CS   YLW   6 GRN-s


  Motor Driver (A4988)                                                                NIMA17 Stepper motor
  ----------------------------------------------                                      ---------------------
  pin01 GND                                       GND
  pin02 VDD                                       5v                                                         Pin1 5v
  pin03 1B                                                                            Motor 1B (color?)
  pin04 1A                                                                            Motor 1A (color?)
  pin05 2A                                                                            Motor 2A (color?)
  pin06 2B                                                                            Moror 2B (color?)
  pin07 GND                                       GND
  pin08 VMOT (7.2v power supply)
  ----
  pin09 ENABLE
  pin10 MS1                                       GP12
  pin11 MS2                                       GP11
  pin12 MS3                                       GP10
  pin13 RESET (connect to SLEEP)
  pin14 SLEEP (connect to RESET)
  PIN15 STEP                                      GP9
  PIN16 DIR                                       GP8



	Pico
	-------------------
	GP0
	GP1
	
	GP2
	GP3
	GP4
	GP5
	
	GP6
	GP7
	GP8
	GP9

	GP10
	GP11
	GP12
	GP13

	GP14
	GP15
	
	----

	VBUS
	VSS
	
	3v3
	3v3(out)
	ADC_VREF
	GP28

	GP27
	GP26
	RUN
	GP22

	GP21
	GP20
	GP19
	GP18

	GP17
	GP16

*/

func main() {

	// run light
	runLight()

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

	raDirection := false
	raDirectionPin := machine.GP8

	raStep := machine.GP9
	var raStepsPerRevolution int32 = 400
	var raMaxHz int32 = 1000
	var raMaxMicroStepSetting int32 = 16
	var raWormRatio int32 = 144
	var raGearRatio int32 = 3
	raMicroStep1 := machine.GP12
	raMicroStep2 := machine.GP11
	raMicroStep3 := machine.GP10
	raEncoderSPI := *machine.SPI0
	raEncoderCS := machine.GP20

	ra, _ := driver.NewRADriver(
		raStep,
		raPWM,
		raDirection,
		raDirectionPin,
		raStepsPerRevolution,
		raMaxHz,
		raMicroStep1,
		raMicroStep2,
		raMicroStep3,
		raMaxMicroStepSetting,
		raWormRatio,
		raGearRatio,
		raEncoderSPI,
		raEncoderCS,
	)
	ra.Configure()
	// ra.RunAtHz(700.0)
	// ra.RunAtHz(200.0)
	ra.RunAtSiderealRate()
	

	var position uint32 = 0
	var lastPosition int = 0

	// 
	// Track by the second
	//


		for  i := 0; i < 60; i++ {

			position = ra.GetPosition()
			
			pos := int(position)
			perSec := math.Abs(float64(pos-lastPosition))

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
		ra.ZeroRA() // DEVTODO done not seem to work, make sure I am clearing the rotation count as well
		time.Sleep(time.Millisecond * 5000)

		//
		// Track for a few min
		for  i := 0; i < 5; i++ {

			position = ra.GetPosition()
			
			pos := int(position)
			perMin := math.Abs(float64(pos-lastPosition))

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
	for i := 0; i < 50; i++ {
		led.High()
		time.Sleep(time.Millisecond * 100)
		led.Low()
		time.Sleep(time.Millisecond * 100)
	}
	led.High()
}
