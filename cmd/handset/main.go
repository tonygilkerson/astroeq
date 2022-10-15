package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/tonygilkerson/astroeq/pkg/hid"
)

func main() {

	// run light
	runLight()

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

	scrollUpKey := machine.GP4
	scrollDnKey := machine.GP2

	rightKey := machine.GP14
	leftKey := machine.GP15
	upKey := machine.GP16
	downKey := machine.GP17

	escKey := machine.GP18
	setupKey := machine.GP19
	enterKey := machine.GP20

	handset, _ := hid.NewHandset(
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
		scrollUpKey,
		scrollDnKey,
		rightKey,
		leftKey,
		upKey,
		downKey,
		escKey,
		setupKey,
		enterKey)

	keyStrokes := handset.Configure()

	//
	// Capture key strokes
	//
	for k := range keyStrokes {
		keyName := handset.GetKeyName(k)

		fmt.Printf("[main] KeyName: %s\n", keyName)
	}

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
