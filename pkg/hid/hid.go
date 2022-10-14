package hid

import (
	"machine"
)

// If any key is pressed record the corresponding pin
var keyPressed machine.Pin

type Handset struct {
	scrollDnKey machine.Pin
	zeroKey     machine.Pin
	scrollUpKey machine.Pin

	sevenKey machine.Pin
	eightKey machine.Pin
	nineKey  machine.Pin

	fourKey machine.Pin
	fiveKey machine.Pin
	sixKey  machine.Pin

	oneKey   machine.Pin
	twoKey   machine.Pin
	threeKey machine.Pin

	rightKey machine.Pin
	leftKey  machine.Pin
	upKey    machine.Pin
	downKey  machine.Pin

	escKey   machine.Pin
	setupKey machine.Pin
	enterKey machine.Pin
}

// Returns a new Handset
func NewHandset(
	scrollDnKey machine.Pin,
	zeroKey     machine.Pin,
	scrollUpKey machine.Pin,

	sevenKey machine.Pin,
	eightKey machine.Pin,
	nineKey  machine.Pin,

	fourKey machine.Pin,
	fiveKey machine.Pin,
	sixKey  machine.Pin,

	oneKey   machine.Pin,
	twoKey   machine.Pin,
	threeKey machine.Pin,

	rightKey machine.Pin,
	leftKey  machine.Pin,
	upKey    machine.Pin,
	downKey  machine.Pin,

	escKey   machine.Pin,
	setupKey machine.Pin,
	enterKey machine.Pin,
) (Handset, error) {
	return Handset{
		scrollDnKey: scrollDnKey,
		zeroKey:     zeroKey,
		scrollUpKey: scrollUpKey,
		sevenKey:    sevenKey,
		eightKey:    eightKey,
		nineKey:     nineKey,
		fourKey:     fourKey,
		fiveKey:     fiveKey,
		sixKey:      sixKey,
		oneKey:      oneKey,
		twoKey:      twoKey,
		threeKey:    threeKey,
		rightKey:    rightKey,
		leftKey:     leftKey,
		upKey:       upKey,
		downKey:     downKey,
		escKey:      escKey,
		setupKey:    setupKey,
		enterKey:    enterKey,
	}, nil
}

func (hs *Handset) Configure() {
	hs.scrollDnKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.zeroKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.scrollUpKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.sevenKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.eightKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.nineKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.fourKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.fiveKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.sixKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.oneKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.twoKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.threeKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.rightKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.leftKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.upKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.downKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.escKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.setupKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.enterKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	hs.scrollDnKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.zeroKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.scrollUpKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.sevenKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.eightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.nineKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.fourKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.fiveKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.sixKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.oneKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.twoKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.threeKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.rightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.leftKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.upKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.downKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.escKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.setupKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
	hs.enterKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = p })
}