// This package is used to control an astronomy equatorial mount
package driver

import (
	"errors"
	"fmt"
	"machine"
	"math"
	"time"

	"github.com/tonygilkerson/astroeq/pkg/encoder"
)

type MicroStep uint16

// Microstep settings
const (
	// MS_FULL      MicroStep = 1     // This seems weird but the TMC2208 does not support a full step
	MS_HALF      MicroStep = 2
	MS_QUARTER   MicroStep = 4
	MS_EIGHTH    MicroStep = 8
	MS_SIXTEENTH MicroStep = 16
)

type RaValue string

const (
	RA_DIRECTION_NORTH RaValue = "North"
	RA_DIRECTION_SOUTH         = "South"
	RA_TRACKING_ON             = "On"
	RA_TRACKING_OFF            = "Off"
)

const SIDEREAL_DAY_IN_SECONDS = 86_164.1

type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
}

// The driver that controls the RA motor
// Based on the A4988 Stepstick Stepper Motor Driver
type RADriver struct {

	// A pulse to this pin will step the motor
	stepPin machine.Pin

	// raPWM
	pwm PWM

	// The pin that controls the direction of the motor rotation
	directionPin machine.Pin

	// The pin that controls the enabling or disabling of the motor
	enableMotorPin machine.Pin

	// The steps need for one full revolution of the motor
	// For example a 1.8° motor takes 200 steps per revolution, a 0.9° motor takes 400 steps per revolution, etc...
	// This is a physical properity of the motor and should NOT account for micro stepping
	stepsPerRevolution int32

	// The maximum PWM cycle in Hz that the motor can accept
	// This is a physical properity of the motor, and for my nima17 0.9° this is around 1_000 Hz
	maxHz int32

	runningHz int32

	// Microstep Pins
	//
	//  ms1  ms2  Steps       Interpolation
	//  ---  ---  ----------- -------------
	//   H    L   1/2         1/256
	//   L    H   1/4         1/256
	//   L    L   1/8         1/256
	//   H    H   1/16        1/256
	//
	microStep1 machine.Pin
	microStep2 machine.Pin

	// The micro stepping setting full, half, quarter, etc...
	// Use 2 for half, 4 for quarter etc...
	microStepSetting MicroStep

	// The limit or "highest" setting, probably 16
	maxMicroStepSetting MicroStep

	// The gear ratios of your RA mount
	// reference: http://www.astrofriend.eu/astronomy/astronomy-calculations/mount-gearbox-ratio/mount-gearbox-ratio.html
	// Common worm drives are 130:1, 135:1, 144:1, 180:1, 435:1; thus use values of 130, 135, 144, 180 or 435 respectively
	wormRatio int32

	// Common primary gear ratios are from 1:1 to 75:1; thus use values 1 to 75 respectively
	// This is the total ratio of all gears combined, for example:
	// if you have a primary gearbox with a ratio of 12:1 and a secondary gearbox with a ration of 10:1 then set GearRatio to (12*10) or 120
	gearRatio int32

	// RA Encoder
	encoder.RAEncoder

	// RA Encoder
	position uint32
}

// Returns a new RADriver
func NewRADriver(
	stepPin machine.Pin,
	pwm PWM,
	directionPin machine.Pin,
	stepsPerRevolution int32,
	maxHz int32,
	microStep1 machine.Pin,
	microStep2 machine.Pin,
	maxMicroStepSetting MicroStep,
	enableMotorPin machine.Pin,
	wormRatio int32,
	gearRatio int32,
	encoderSPI machine.SPI,
	encoderCS machine.Pin,

) (RADriver, error) {

	if maxMicroStepSetting != MS_HALF &&
		maxMicroStepSetting != MS_QUARTER &&
		maxMicroStepSetting != MS_EIGHTH &&
		maxMicroStepSetting != MS_SIXTEENTH {
		return RADriver{}, errors.New("maxMicroStepSetting must be 2, 4, 8 or 16")
	}

	if stepsPerRevolution < 1 {
		return RADriver{}, errors.New("stepsPerRevolution must be greater than 0, typical values are 200 or 400")
	}

	if wormRatio < 1 {
		return RADriver{}, errors.New("wormRatio must be greater than 0, use 1 if not using a worm gear, typical value is 400")
	}

	if gearRatio < 1 {
		return RADriver{}, errors.New("gearRatio must be greater than 0, use 1 if not using a gearbox, typical values between 1 and 75")
	}

	raDriver := RADriver{
		stepPin:             stepPin,
		pwm:                 pwm,
		directionPin:        directionPin,
		stepsPerRevolution:  stepsPerRevolution,
		maxHz:               maxHz,
		runningHz:           0,
		microStep1:          microStep1,
		microStep2:          microStep2,
		microStepSetting:    maxMicroStepSetting,
		maxMicroStepSetting: maxMicroStepSetting,
		enableMotorPin:      enableMotorPin,
		wormRatio:           wormRatio,
		gearRatio:           gearRatio,
	}
	raDriver.ConfigureEncoder(encoderSPI, encoderCS, encoder.RES14)

	return raDriver, nil
}

func (ra *RADriver) Configure() {

	//
	// Configure the machine PWM for the RA
	// See https://datasheets.raspberrypi.com/rp2040/rp2040-datasheet.pdf
	//     4.5.2. Programmer’s Model
	//
	ra.pwm.Configure(machine.PWMConfig{Period: 0})
	chA, _ := ra.pwm.Channel(ra.stepPin)
	ra.pwm.Set(chA, ra.pwm.Top()/2)

	//
	// Microstepping
	//
	microStep1 := ra.microStep1
	microStep2 := ra.microStep2
	microStep1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	microStep2.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Default to microStepSetting of 16
	ra.setMicroStepSetting(MS_SIXTEENTH)

	// Direction
	ra.directionPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ra.SetDirection(RA_DIRECTION_NORTH)

	// Enable Motor
	ra.enableMotorPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ra.SetTracking(RA_TRACKING_OFF)

	// RA Encoder
	ra.ZeroRA()

	// Start go routine to monitor position
	go ra.monitorPositionRoutine()

}

func (ra *RADriver) setMicroStepSetting(ms MicroStep) {

	ra.microStepSetting = ms

	//  ms1  ms2  Steps       Interpolation
	//  ---  ---  ----------- -------------
	//   H    L   1/2         1/256
	//   L    H   1/4         1/256
	//   L    L   1/8         1/256
	//   H    H   1/16        1/256

	switch ra.microStepSetting {
	case 2:
		ra.microStep1.High()
		ra.microStep2.Low()
		fmt.Println("[setMicroStepSetting] microStepSetting 2-H L")
	case 4:
		ra.microStep1.Low()
		ra.microStep2.High()
		fmt.Println("[setMicroStepSetting] microStepSetting 4-L H")
	case 8:
		ra.microStep1.Low()
		ra.microStep2.Low()
		fmt.Println("[setMicroStepSetting] microStepSetting 8-H H")
	case 16:
		ra.microStep1.High()
		ra.microStep2.High()
		fmt.Println("[setMicroStepSetting] microStepSetting 16-H H")
	default:
		ra.microStep1.High()
		ra.microStep2.High()
		fmt.Println("[setMicroStepSetting] microStepSetting default 16-H H")
	}

}

// Set to run at Sidereal rate, that is the RA will do one full rotation in one sidereal day
//
// To compute the PWM cycle that is needed to drive the system at a siderial rate, for example given:
//
//	 stepsPerRevolution  = 400
//	 maxMicroStepSetting = 16
//	 wormRatio           = 144 (144:1)
//	 gearRatio           = 3   (48:16)
//	                     ============
//			    								2_764_800 (system ratio 400*16*144*3)
//
//	 The cycle Hz = system ratio / number of seconds in a sideral day
//	 The cycle perod = 1e9 / Hz
func (ra *RADriver) RunAtSiderealRate() {

	systemRatio := ra.stepsPerRevolution * int32(ra.maxMicroStepSetting) * ra.wormRatio * ra.gearRatio
	sideralHz := float64(systemRatio) / SIDEREAL_DAY_IN_SECONDS

	ra.RunAtHz(sideralHz)

}

func (ra *RADriver) RunAtHz(hz float64) {

	fmt.Printf("[RunAtHz] Set hz to: %.2f\n", hz)
	period := uint64(math.Round(1e9 / hz))

	// Save Hz on RA Driver
	ra.runningHz = int32(hz)

	// Set period for hardware PWM
	ra.pwm.SetPeriod(period)
}

func (ra *RADriver) monitorPositionRoutine() {

	for {
		position, err := ra.GetPositionRA()
		if err == nil {
			ra.position = position
		} else {
			println("[monitorPositionRoutine] Error getting position")
		}
		time.Sleep(time.Millisecond * 700) //DEVTODO - not sure if this is too short or too long?
	}
}

func (ra *RADriver) GetTracking() RaValue {

	// Enabled if pin is low, so if true return off
	if ra.enableMotorPin.Get() {
		return RA_TRACKING_OFF
	} else {
		return RA_TRACKING_ON
	}

}

func (ra *RADriver) GetPosition() uint32 {
	return ra.position
}

func (ra *RADriver) GetDirection() RaValue {
	if ra.directionPin.Get() {
		return RA_DIRECTION_NORTH
	} else {
		return RA_DIRECTION_SOUTH
	}
}

func (ra *RADriver) SetDirection(direction RaValue) {

	if direction == RA_DIRECTION_NORTH {
		ra.directionPin.High()
	} else {
		ra.directionPin.Low()
	}

}

func (ra *RADriver) SetTracking(tracking RaValue) {

	if tracking == RA_TRACKING_ON {
		ra.enableMotorPin.Low() // Enabled if pin is low
	} else {
		ra.enableMotorPin.High()
	}

}
