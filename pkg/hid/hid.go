package hid

import (
	"machine"
	"strconv"
	"time"
)

const Version string = "v0-alpha4"
const MMDDYY = "010206"
const LocationLatitudeHome = "+39.8491"
const LocationLongitudeHome = "-83.9768"
const LocationElevation = "+270"

const (
	First = iota
	ShowVersion
	SetDate
	SetTime
	SetLatitude
	SetLongitude
	SetElevation
	Last
	SetDateError
	SetTimeError
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
	currentTimeStr string
	currentTime    time.Time

	// The latitude and longitude of your current location in DD format
	//
	// Some reference info:
	// Decimal degrees (DD): 41.40338, 2.17403
	// Degrees, minutes, and seconds (DMS): 41°24'12.2"N 2°10'26.5"E
	// Degrees and decimal minutes (DMM): 41 24.2028, 2 10.4418
	// 8822 Dayton-springfield Rd is:
	//   DD:
	//		Latitude:   39.8490726
	//    Longitude: -83.9767929
	//    Elevation:  15.96z
	//   DMS:
	//		Latitude:  N 39° 56'
	//		Longitude: W 83° 50'
	//    Elevation: 270m

	locationLatitudeStr  string
	locationLongitudeStr string

	locationElevationStr string
	locationElevation    int16
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
		state:                First,
		scrollDnKey:          scrollDnKey,
		zeroKey:              zeroKey,
		scrollUpKey:          scrollUpKey,
		sevenKey:             sevenKey,
		eightKey:             eightKey,
		nineKey:              nineKey,
		fourKey:              fourKey,
		fiveKey:              fiveKey,
		sixKey:               sixKey,
		oneKey:               oneKey,
		twoKey:               twoKey,
		threeKey:             threeKey,
		rightKey:             rightKey,
		leftKey:              leftKey,
		upKey:                upKey,
		downKey:              downKey,
		escKey:               escKey,
		setupKey:             setupKey,
		enterKey:             enterKey,
		keyStrokes:           make(chan Key, 100),
		locationLatitudeStr:  LocationLatitudeHome,
		locationLongitudeStr: LocationLongitudeHome,
		locationElevationStr: LocationElevation,
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
			time.Sleep(time.Millisecond * 150)
			// fmt.Printf("[publishKeys] value: %v \n", key)

			if key == hs.keyPressed {
				hs.keyStrokes <- hs.keyPressed
				hs.keyPressed = UndefinedKey //reset for next key press
			}
		}
		time.Sleep(time.Millisecond * 500)
	}

}

func (hs *Handset) GetKeyString(k Key) string {

	// fmt.Printf("[GetKeyString] value: %v \n", k)
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
		return "Undefined"
	}
}

func (hs *Handset) GetKeyFromString(s string) Key {
	switch s {
	case "0":
		return ZeroKey
	case "1":
		return OneKey
	case "2":
		return TwoKey
	case "3":
		return ThreeKey
	case "4":
		return FourKey
	case "5":
		return FiveKey
	case "6":
		return SixKey
	case "7":
		return SevenKey
	case "8":
		return EightKey
	case "9":
		return NineKey
	case "ScrollDn":
		return ScrollDnKey
	case "ScrollUp":
		return ScrollUpKey
	case "Right":
		return RightKey
	case "Left":
		return LeftKey
	case "Up":
		return UpKey
	case "Down":
		return DownKey
	case "ESC":
		return EscKey
	case "Setup":
		return SetupKey
	case "Enter":
		return EnterKey
	default:
		return UndefinedKey
	}
}

func (hs *Handset) StateMachine(key Key) string {

	switch hs.state {

	case First:
		hs.state++

	case ShowVersion:
		doNav(key, &hs.state)

	case SetDate:

		if key == EnterKey {
			var err error
			// RFC3339 example: "2006-01-02T15:04:05+05:00"
			hs.currentTime, err = time.Parse(time.RFC3339, hs.currentDateStr+"T10:00:00+00:00")

			if err != nil {
				hs.state = SetDateError
			} else {
				hs.state = SetDate + 1
			}

		} else if !doNav(key, &hs.state) {

			if key == LeftKey && len(hs.currentDateStr) > 0 {
				hs.currentDateStr = hs.currentDateStr[:len(hs.currentDateStr)-1]

			} else if len(hs.currentDateStr) == 4 && keyIsDigit(key) {
				hs.currentDateStr = hs.currentDateStr + "-" + hs.GetKeyString(key)

			} else if len(hs.currentDateStr) == 7 && keyIsDigit(key) {
				hs.currentDateStr = hs.currentDateStr + "-" + hs.GetKeyString(key)

			} else if len(hs.currentDateStr) < 10 && keyIsDigit(key) {
				hs.currentDateStr = hs.currentDateStr + hs.GetKeyString(key)

			}
		}

	case SetDateError:
		if key == EscKey {
			hs.state = SetDate
		}

	case SetTime:

		if key == EnterKey {
			var err error
			// RFC3339 example: "2006-01-02T15:04:05+05:00"
			hs.currentTime, err = time.Parse(time.RFC3339, hs.currentDateStr+"T"+hs.currentTimeStr+":00")

			if err != nil {
				hs.state = SetTimeError
			} else {
				hs.state = SetTime + 1
			}

		} else if !doNav(key, &hs.state) {

			if key == LeftKey && len(hs.currentTimeStr) > 0 {
				hs.currentTimeStr = hs.currentTimeStr[:len(hs.currentTimeStr)-1]

			} else if len(hs.currentTimeStr) == 2 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + ":" + hs.GetKeyString(key)

			} else if len(hs.currentTimeStr) == 5 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + ":" + hs.GetKeyString(key)

			} else if len(hs.currentTimeStr) == 8 {

				if key == ScrollDnKey {
					hs.currentTimeStr = hs.currentTimeStr + "-"
				} else if key == ScrollUpKey {
					hs.currentTimeStr = hs.currentTimeStr + "+"
				}

			} else if len(hs.currentTimeStr) < 11 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + hs.GetKeyString(key)

			}
		}

	case SetTimeError:
		if key == EscKey {
			hs.state = SetTime
		}

	case SetLatitude:

		if !doNav(key, &hs.state) {

			if key == LeftKey && len(hs.locationLatitudeStr) > 0 {
				hs.locationLatitudeStr = hs.locationLatitudeStr[:len(hs.locationLatitudeStr)-1]

			} else if len(hs.locationLatitudeStr) == 3 && keyIsDigit(key) {
				hs.locationLatitudeStr = hs.locationLatitudeStr + "." + hs.GetKeyString(key)

			} else if len(hs.locationLatitudeStr) == 0 {

				if key == ScrollDnKey {
					hs.locationLatitudeStr = "-"
				} else if key == ScrollUpKey {
					hs.locationLatitudeStr = "+"
				}

			} else if len(hs.locationLatitudeStr) < 8 && keyIsDigit(key) {
				hs.locationLatitudeStr = hs.locationLatitudeStr + hs.GetKeyString(key)

			}
		}

	case SetLongitude:

		if !doNav(key, &hs.state) {

			if key == LeftKey && len(hs.locationLongitudeStr) > 0 {
				hs.locationLongitudeStr = hs.locationLongitudeStr[:len(hs.locationLongitudeStr)-1]

			} else if len(hs.locationLongitudeStr) == 3 && keyIsDigit(key) {
				hs.locationLongitudeStr = hs.locationLongitudeStr + "." + hs.GetKeyString(key)

			} else if len(hs.locationLongitudeStr) == 0 {

				if key == ScrollDnKey {
					hs.locationLongitudeStr = "-"
				} else if key == ScrollUpKey {
					hs.locationLongitudeStr = "+"
				}

			} else if len(hs.locationLongitudeStr) < 8 && keyIsDigit(key) {
				hs.locationLongitudeStr = hs.locationLongitudeStr + hs.GetKeyString(key)

			}
		}

	case SetElevation:

		if key == EnterKey {
			elevation, _ := strconv.Atoi(hs.locationElevationStr)
			hs.locationElevation = int16(elevation)
			hs.state++

		} else if !doNav(key, &hs.state) {

			if key == LeftKey && len(hs.locationElevationStr) > 0 {
				hs.locationElevationStr = hs.locationElevationStr[:len(hs.locationElevationStr)-1]

			} else if len(hs.locationElevationStr) == 0 {

				if key == ScrollDnKey {
					hs.locationElevationStr = "-"
				} else if key == ScrollUpKey {
					hs.locationElevationStr = "+"
				}

			} else if len(hs.locationElevationStr) < 5 && keyIsDigit(key) {
				hs.locationElevationStr = hs.locationElevationStr + hs.GetKeyString(key)

			}
		}

	case Last:
		doNav(key, &hs.state)

	}

	//
	// Set prompt
	//  11 char per line at Regular9pt7b
	//
	switch hs.state {
	case First:
		hs.dspOut = "Version\n" + Version

	case ShowVersion:
		hs.dspOut = "Version\n" + Version

	case SetDate:
		hs.dspOut = "Set Date\nYYYY-MM-DD\n-----------\n" + hs.currentDateStr

	case SetDateError:
		hs.dspOut = "Set Date\nYYYY-MM-DD\n-----------\n" + hs.currentDateStr + "\n>>ERROR<<"

	case SetTime:
		hs.dspOut = "Set Time\nHH:MM:SS+hh\n-----------\n" + hs.currentTimeStr

	case SetTimeError:
		hs.dspOut = "Set Time\nHH:MM:SS+hh\n-----------\n" + hs.currentTimeStr + "\n>>ERROR<<"

	case SetLatitude:
		hs.dspOut = "Set\nLatitude\n+DD.dddd\n-----------\n" + hs.locationLatitudeStr

	case SetLongitude:
		hs.dspOut = "Set\nLongitude\n+DD.dddd\n-----------\n" + hs.locationLongitudeStr

	case SetElevation:
		hs.dspOut = "Set\nElevation\n+DDDD\n-----------\n" + hs.locationElevationStr

	case Last:
		hs.dspOut = "Press Esc\nto go back"

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

func doNav(key Key, state *State) bool {

	if key == EscKey {

		if *state > First {
			*state--
		}
		return true

	} else if key == EnterKey {

		if *state < Last {
			*state++
		}
		return true

	} else {

		return false
	}

}
