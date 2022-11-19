package hid

import (
	// "fmt"
	"fmt"
	"machine"
	"time"
)

const Version string = "v0.0.1"
const MMDDYY = "010206"

const (
	First = iota
	ShowVersion
	SetDate
	SetDateError
	Next
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

type State uint8

type Handset struct {
	state       State
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

	// If any key is pressed record the corresponding pin
	keyPressed Key

	// As keys are pressed they are published to the keyStrokes chan
	// a buffer of 100 ensure key stroke publishing is not blocked
	keyStrokes chan Key

	// Display output
	dspOut string

	// Current data set by user at startup in MMDDYY format
	currentDateStr string
	currentDate    time.Time
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
		state:       First,
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
		keyStrokes:  make(chan Key, 100),
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

	hs.zeroKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = ZeroKey })
	hs.oneKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = OneKey })
	hs.twoKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = TwoKey })
	hs.threeKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = ThreeKey })
	hs.fourKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = FourKey })
	hs.fiveKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = FiveKey })
	hs.sixKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = SixKey })
	hs.sevenKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = SevenKey })
	hs.eightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = EightKey })
	hs.nineKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = NineKey })
	hs.scrollUpKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = ScrollUpKey })
	hs.scrollDnKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = ScrollDnKey })
	hs.rightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = RightKey })
	hs.leftKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = LeftKey })
	hs.upKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = UpKey })
	hs.downKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = DownKey })
	hs.escKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = EscKey })
	hs.setupKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = SetupKey })
	hs.enterKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = EnterKey })

	// Start go routine that will listen for key strokes and publish them on a channel
	go hs.publishKeys()

	return hs.keyStrokes

}

// PublishKeys will capture the keys pressed and publish them to the keyStrokes channel
func (hs *Handset) publishKeys() {
	for {

		// If any key was pressed
		if hs.keyPressed != UndefinedKey {

			//
			//  After a small delay if the key pressed has not changed, consider it "pressed"
			//
			key := hs.keyPressed
			time.Sleep(time.Millisecond * 100)
			// fmt.Printf("[publishKeys] value: %v \n", key)

			if key == hs.keyPressed {
				hs.keyStrokes <- hs.keyPressed
				hs.keyPressed = UndefinedKey //reset for next key press
			}
		}
		time.Sleep(time.Millisecond * 400)
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

func (hs *Handset) StateMachine(key Key) string {

	switch hs.state {

	case First:
		hs.state = ShowVersion

	case ShowVersion:
		if key == EnterKey {
			hs.state = SetDate
			hs.currentDateStr = ""
		}

	case SetDate:

		if key == EscKey {
			hs.state = ShowVersion

		} else if key == EnterKey {
			var err error
			hs.currentDate, err = time.Parse(MMDDYY, hs.currentDateStr)

			if err != nil {
				hs.state = SetDateError
			} else {
				hs.state = Next
			}

		} else if key == LeftKey && len(hs.currentDateStr) > 0 {
			hs.currentDateStr = hs.currentDateStr[:len(hs.currentDateStr)-1]

		} else {
			//
			// Build up a date string until we have 6 characters to make mmddyy
			// then try to convert it to a date
			if len(hs.currentDateStr) < 6 && keyIsDigit(key) {
				hs.currentDateStr = hs.currentDateStr + hs.GetKeyName(key)
			}
		}

	case SetDateError:
		if key == EscKey {
			hs.state = SetDate
		}

	case Next:

		if key == EscKey {
			hs.state = SetDate
		}

		if key == EnterKey {
			hs.state = ShowVersion
		}

	}

	//
	// Set prompt
	//
	switch hs.state {
	case First:
		hs.dspOut = "Version\n" + Version
	case ShowVersion:
		hs.dspOut = "Version\n" + Version
	case SetDate:
		hs.dspOut = "Set Date\nMMDDYY:\n" + hs.currentDateStr
	case SetDateError:
		hs.dspOut = fmt.Sprintf("Set Date\nMMDDYY:\n%s\nERR", hs.currentDateStr)
	case Next:
		hs.dspOut = "Next"
	default:
		hs.dspOut = "Bad State\n"
	}

	return hs.dspOut
}

func keyIsDigit(key Key) bool {
	switch key {
	case ZeroKey:
		return true
	case OneKey:
		return true
	case TwoKey:
		return true
	case ThreeKey:
		return true
	case FourKey:
		return true
	case FiveKey:
		return true
	case SixKey:
		return true
	case SevenKey:
		return true
	case EightKey:
		return true
	case NineKey:
		return true
	default:
		return false
	}
}

func setDate(mmddyy string)
