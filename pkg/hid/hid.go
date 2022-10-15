package hid

import (
	// "fmt"
	"machine"
	"time"
)

type Key uint8

const (
	UndefinedKey Key = iota
	ZeroKey
	OneKey
	TwoKey
	ThreeKey
	FourKey
	FiveKey
	SixKey
	SevenKey
	EightKey
	NineKey

	ScrollDnKey
	ScrollUpKey

	RightKey
	LeftKey
	UpKey
	DownKey

	EscKey
	SetupKey
	EnterKey
)

// If any key is pressed record the corresponding pin
var keyPressed Key

// As keys are pressed they are published to the keyStrokes chan
// a buffer of 100 ensure key stroke publishing is not blocked
var keyStrokes = make(chan Key, 100)

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
	zeroKey machine.Pin,
	oneKey machine.Pin,
	twoKey machine.Pin,
	threeKey machine.Pin,
	fourKey machine.Pin,
	fiveKey machine.Pin,
	sixKey machine.Pin,
	sevenKey machine.Pin,
	eightKey machine.Pin,
	nineKey machine.Pin,
	
	scrollDnKey machine.Pin,
	scrollUpKey machine.Pin,

	rightKey machine.Pin,
	leftKey machine.Pin,
	upKey machine.Pin,
	downKey machine.Pin,

	escKey machine.Pin,
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

// Configure - will configure the HID pins and assign each to a key
// starts a go routine to listen for key strokes and publishes each to the key chan
// the key channel is returned for key stroke subscribers
func (hs *Handset) Configure() chan Key {
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

	hs.zeroKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = ZeroKey })
	hs.oneKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = OneKey })
	hs.twoKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = TwoKey })
	hs.threeKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = ThreeKey })
	hs.fourKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = FourKey })
	hs.fiveKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = FiveKey })
	hs.sixKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = SixKey })
	hs.sevenKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = SevenKey })
	hs.eightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = EightKey })
	hs.nineKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = NineKey })
	hs.scrollUpKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = ScrollUpKey })
	hs.scrollDnKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = ScrollDnKey })
	hs.rightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = RightKey })
	hs.leftKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = LeftKey })
	hs.upKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = UpKey })
	hs.downKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = DownKey })
	hs.escKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = EscKey })
	hs.setupKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = SetupKey })
	hs.enterKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { keyPressed = EnterKey })

	// Start go routine that will listen for key strokes and publish them on a channel
	go publishKeys()

	return keyStrokes

}

// PublishKeys will capture the keys pressed and publish them to the keyStrokes channel
func publishKeys() {
	for {

		// If any key was pressed
		if keyPressed != UndefinedKey {

			//
			//  After a small delay if the key pressed has not changed, consider it "pressed"
			//
			key := keyPressed
			time.Sleep(time.Millisecond * 100)
			// fmt.Printf("[publishKeys] value: %v \n", key)

			if key == keyPressed {
				keyStrokes <- keyPressed
				keyPressed = UndefinedKey //reset for next key press
			}
		}
		time.Sleep(time.Millisecond * 500)
	}

}

func (hs *Handset) GetKeyName(k Key) string {

	// fmt.Printf("[GetKeyName] value: %v \n", k)
	switch k {
	case ZeroKey:
		return "0"
	case OneKey:
		return "1"
	case TwoKey:
		return "2"
	case ThreeKey:
		return "3"
	case FourKey:
		return "4"
	case FiveKey:
		return "5"
	case SixKey:
		return "6"
	case SevenKey:
		return "7"
	case EightKey:
		return "8"
	case NineKey:
		return "9"
	case ScrollDnKey:
		return "ScrollDn"
	case ScrollUpKey:
		return "ScrollUp"
	case RightKey:
		return "Right"
	case LeftKey:
		return "Left"
	case UpKey:
		return "Up"
	case DownKey:
		return "Down"
	case EscKey:
		return "ESC"
	case SetupKey:
		return "Setup"
	case EnterKey:
		return "Enter"
	default:
		return "Unknown"
	}
}

