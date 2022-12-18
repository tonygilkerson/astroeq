package hid

import (
	"fmt"
	"image/color"
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/ssd1351"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

const VERSION string = "v0-alpha4"
const MMDDYY = "010206"
const LOCATION_LATITUDE_HOME = "+39.8491"
const LOCATION_LONGITUDE_HOME = "-83.9768"
const LOCATION_ELEVATION = "+270"

// The handset is controlled by a state machine. State is used to control what is
// displayed on the screen and what keys are valid input
type State uint8

const (
	FIRST State = iota
	SHOW_VERSION
	SET_DATE
	SET_TIME
	SET_LATITUDE
	SET_LONGITUDE
	SET_ELEVATION
	UTILITY_MENU
	OBJECTS_MENU
	LAST
	SET_DATE_Error
	SET_TIME_MSG_ERROR
)

// Each button on the handset corresponds to one of the following Keys
type Key uint8

const (
	UNDEFINED_KEY Key = iota
	ZERO_KEY
	ONE_KEY
	TWO_KEY
	THREE_KEY
	FOUR_KEY
	FIVE_KEY
	SIX_KEY
	SEVEN_KEY
	EIGHT_KEY
	NINE_KEY

	SCROLL_DN_KEY
	SCROLL_UP_KEY

	RIGHT_KEY
	LEFT_KEY
	UP_KEY
	DOWN_KEY

	ESC_KEY
	SETUP_KEY
	ENTER_KEY
)

// The Handset properties are maintained by the user via the handset.
// The user can perform basic CRUD operations on the Handset properties as well as
// use them in commands sent over the message bus.
type Handset struct {
	screen       *Screen
	isSetup      bool
	state        State
	scrollDnKey  machine.Pin
	zeroKey      machine.Pin
	scrollUP_KEY machine.Pin

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

// The Screen properties are used to determine what is written to the display
type Screen struct {
	displayDevice *ssd1351.Device
	font          *tinyfont.Font
	fontColor     color.RGBA
	direction     bool
	statusBarText string
	bodyText      string
}

// Returns a new Handset
func NewHandset(
	displayDevice *ssd1351.Device,
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
	scrollUP_KEY machine.Pin,

	rightKey machine.Pin,
	leftKey machine.Pin,
	upKey machine.Pin,
	downKey machine.Pin,

	escKey machine.Pin,
	setupKey machine.Pin,
	enterKey machine.Pin,
) (Handset, error) {

	var screen Screen
	screen.displayDevice = displayDevice

	return Handset{
		screen:               &screen,
		isSetup:              false,
		state:                FIRST,
		scrollDnKey:          scrollDnKey,
		zeroKey:              zeroKey,
		scrollUP_KEY:         scrollUP_KEY,
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
		keyPressed:           0,
		keyStrokes:           make(chan Key, 100),
		dspOut:               "",
		currentDateStr:       "2022",
		currentTimeStr:       "",
		currentTime:          time.Time{},
		locationLatitudeStr:  LOCATION_LATITUDE_HOME,
		locationLongitudeStr: LOCATION_LONGITUDE_HOME,
		locationElevationStr: LOCATION_ELEVATION,
		locationElevation:    0,
	}, nil
}

// Configure - will configure the HID pins and assign each to a key
// starts a go routine to listen for key strokes and publishes each to the key chan
// the key channel is returned for key stroke subscribers
func (hs *Handset) Configure() chan Key {

	//
	// Init Screen
	//
	hs.screen.font = &freemono.Regular9pt7b
	hs.screen.fontColor = color.RGBA{0, 0, 255, 255} // RED
	hs.screen.statusBarText = "IgGLpq|X"

	//
	// Configure Key Pins
	//
	hs.scrollDnKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.zeroKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.scrollUP_KEY.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
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

	//
	// Register interrupts
	//
	hs.zeroKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = ZERO_KEY })
	hs.oneKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = ONE_KEY })
	hs.twoKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = TWO_KEY })
	hs.threeKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = THREE_KEY })
	hs.fourKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = FOUR_KEY })
	hs.fiveKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = FIVE_KEY })
	hs.sixKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = SIX_KEY })
	hs.sevenKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = SEVEN_KEY })
	hs.eightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = EIGHT_KEY })
	hs.nineKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = NINE_KEY })
	hs.scrollUP_KEY.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = SCROLL_UP_KEY })
	hs.scrollDnKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = SCROLL_DN_KEY })
	hs.rightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = RIGHT_KEY })
	hs.leftKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = LEFT_KEY })
	hs.upKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = UP_KEY })
	hs.downKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = DOWN_KEY })
	hs.escKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = ESC_KEY })
	hs.setupKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = SETUP_KEY })
	hs.enterKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = ENTER_KEY })

	//
	// Start go routine that will listen for key strokes and publish them on a channel
	//
	go hs.publishKeysRoutine()

	return hs.keyStrokes

}

// PublishKeys will capture the keys pressed and publish them to the keyStrokes channel
func (hs *Handset) publishKeysRoutine() {
	for {

		// If any key was pressed
		if hs.keyPressed != UNDEFINED_KEY {

			//
			//  After a small delay if the key pressed has not changed, consider it "pressed"
			//
			key := hs.keyPressed
			time.Sleep(time.Millisecond * 150)
			// fmt.Printf("[publishKeysRoutine] value: %v \n", key)

			if key == hs.keyPressed {
				hs.keyStrokes <- hs.keyPressed
				hs.keyPressed = UNDEFINED_KEY //reset for next key press
			}
		}
		time.Sleep(time.Millisecond * 500)
	}

}

func (hs *Handset) GetKeyString(k Key) string {

	// fmt.Printf("[GetKeyString] value: %v \n", k)
	switch k {
	case ZERO_KEY:
		return "0"
	case ONE_KEY:
		return "1"
	case TWO_KEY:
		return "2"
	case THREE_KEY:
		return "3"
	case FOUR_KEY:
		return "4"
	case FIVE_KEY:
		return "5"
	case SIX_KEY:
		return "6"
	case SEVEN_KEY:
		return "7"
	case EIGHT_KEY:
		return "8"
	case NINE_KEY:
		return "9"
	case SCROLL_DN_KEY:
		return "ScrollDn"
	case SCROLL_UP_KEY:
		return "ScrollUp"
	case RIGHT_KEY:
		return "Right"
	case LEFT_KEY:
		return "Left"
	case UP_KEY:
		return "Up"
	case DOWN_KEY:
		return "Down"
	case ESC_KEY:
		return "ESC"
	case SETUP_KEY:
		return "Setup"
	case ENTER_KEY:
		return "Enter"
	default:
		return "Undefined"
	}
}

func (hs *Handset) GetKeyFromString(s string) Key {
	switch s {
	case "0":
		return ZERO_KEY
	case "1":
		return ONE_KEY
	case "2":
		return TWO_KEY
	case "3":
		return THREE_KEY
	case "4":
		return FOUR_KEY
	case "5":
		return FIVE_KEY
	case "6":
		return SIX_KEY
	case "7":
		return SEVEN_KEY
	case "8":
		return EIGHT_KEY
	case "9":
		return NINE_KEY
	case "ScrollDn":
		return SCROLL_DN_KEY
	case "ScrollUp":
		return SCROLL_UP_KEY
	case "Right":
		return RIGHT_KEY
	case "Left":
		return LEFT_KEY
	case "Up":
		return UP_KEY
	case "Down":
		return DOWN_KEY
	case "ESC":
		return ESC_KEY
	case "Setup":
		return SETUP_KEY
	case "Enter":
		return ENTER_KEY
	default:
		return UNDEFINED_KEY
	}
}

func (hs *Handset) StateMachine(key Key) string {

	switch hs.state {

	case FIRST:

		if key == ONE_KEY {
			hs.state++
		} else if key == TWO_KEY {
			hs.state = UTILITY_MENU
		} else if key == THREE_KEY {
			hs.state = OBJECTS_MENU
		}

	case SHOW_VERSION:
		hs.isSetup = false
		doNav(key, &hs.state)

	case SET_DATE:

		if key == ENTER_KEY {
			var err error
			// RFC3339 example: "2006-01-02T15:04:05+05:00"
			hs.currentTime, err = time.Parse(time.RFC3339, hs.currentDateStr+"T10:00:00+00:00")

			if err != nil {
				hs.state = SET_DATE_Error
			} else {
				hs.state = SET_DATE + 1
			}

		} else if !doNav(key, &hs.state) {

			if key == LEFT_KEY && len(hs.currentDateStr) > 0 {
				hs.currentDateStr = hs.currentDateStr[:len(hs.currentDateStr)-1]

			} else if len(hs.currentDateStr) == 4 && keyIsDigit(key) {
				hs.currentDateStr = hs.currentDateStr + "-" + hs.GetKeyString(key)

			} else if len(hs.currentDateStr) == 7 && keyIsDigit(key) {
				hs.currentDateStr = hs.currentDateStr + "-" + hs.GetKeyString(key)

			} else if len(hs.currentDateStr) < 10 && keyIsDigit(key) {
				hs.currentDateStr = hs.currentDateStr + hs.GetKeyString(key)

			}
		}

	case SET_DATE_Error:
		if key == ESC_KEY {
			hs.state = SET_DATE
		}

	case SET_TIME:

		if key == ENTER_KEY {
			var err error
			// RFC3339 example: "2006-01-02T15:04:05+05:00"
			hs.currentTime, err = time.Parse(time.RFC3339, hs.currentDateStr+"T"+hs.currentTimeStr+":00")

			if err != nil {
				hs.state = SET_TIME_MSG_ERROR
			} else {
				hs.state = SET_TIME + 1
			}

		} else if !doNav(key, &hs.state) {

			if key == LEFT_KEY && len(hs.currentTimeStr) > 0 {
				hs.currentTimeStr = hs.currentTimeStr[:len(hs.currentTimeStr)-1]

			} else if len(hs.currentTimeStr) == 2 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + ":" + hs.GetKeyString(key)

			} else if len(hs.currentTimeStr) == 5 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + ":" + hs.GetKeyString(key)

			} else if len(hs.currentTimeStr) == 8 {

				if key == SCROLL_DN_KEY {
					hs.currentTimeStr = hs.currentTimeStr + "-"
				} else if key == SCROLL_UP_KEY {
					hs.currentTimeStr = hs.currentTimeStr + "+"
				}

			} else if len(hs.currentTimeStr) < 11 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + hs.GetKeyString(key)

			}
		}

	case SET_TIME_MSG_ERROR:
		if key == ESC_KEY {
			hs.state = SET_TIME
		}

	case SET_LATITUDE:

		if !doNav(key, &hs.state) {

			if key == LEFT_KEY && len(hs.locationLatitudeStr) > 0 {
				hs.locationLatitudeStr = hs.locationLatitudeStr[:len(hs.locationLatitudeStr)-1]

			} else if len(hs.locationLatitudeStr) == 3 && keyIsDigit(key) {
				hs.locationLatitudeStr = hs.locationLatitudeStr + "." + hs.GetKeyString(key)

			} else if len(hs.locationLatitudeStr) == 0 {

				if key == SCROLL_DN_KEY {
					hs.locationLatitudeStr = "-"
				} else if key == SCROLL_UP_KEY {
					hs.locationLatitudeStr = "+"
				}

			} else if len(hs.locationLatitudeStr) < 8 && keyIsDigit(key) {
				hs.locationLatitudeStr = hs.locationLatitudeStr + hs.GetKeyString(key)

			}
		}

	case SET_LONGITUDE:

		if !doNav(key, &hs.state) {

			if key == LEFT_KEY && len(hs.locationLongitudeStr) > 0 {
				hs.locationLongitudeStr = hs.locationLongitudeStr[:len(hs.locationLongitudeStr)-1]

			} else if len(hs.locationLongitudeStr) == 3 && keyIsDigit(key) {
				hs.locationLongitudeStr = hs.locationLongitudeStr + "." + hs.GetKeyString(key)

			} else if len(hs.locationLongitudeStr) == 0 {

				if key == SCROLL_DN_KEY {
					hs.locationLongitudeStr = "-"
				} else if key == SCROLL_UP_KEY {
					hs.locationLongitudeStr = "+"
				}

			} else if len(hs.locationLongitudeStr) < 8 && keyIsDigit(key) {
				hs.locationLongitudeStr = hs.locationLongitudeStr + hs.GetKeyString(key)

			}
		}

	case SET_ELEVATION:

		if key == ENTER_KEY {
			elevation, _ := strconv.Atoi(hs.locationElevationStr)
			hs.locationElevation = int16(elevation)
			// DEVTODO - If we get to this point then we can turn we are considered to be setup
			//           However it is possible to go back into the setup and "unset" a field
			//           In this case we are not considered setup but I am not resetting the
			//           hs.isSetup field.  I am looking for a clean easy way to do this but for now
			//           hs.isSetup can be wrong under certain situation.
			hs.isSetup = true
			hs.state = FIRST

		} else if !doNav(key, &hs.state) {

			if key == LEFT_KEY && len(hs.locationElevationStr) > 0 {
				hs.locationElevationStr = hs.locationElevationStr[:len(hs.locationElevationStr)-1]

			} else if len(hs.locationElevationStr) == 0 {

				if key == SCROLL_DN_KEY {
					hs.locationElevationStr = "-"
				} else if key == SCROLL_UP_KEY {
					hs.locationElevationStr = "+"
				}

			} else if len(hs.locationElevationStr) < 5 && keyIsDigit(key) {
				hs.locationElevationStr = hs.locationElevationStr + hs.GetKeyString(key)

			}
		}

	case UTILITY_MENU:

		if key == ESC_KEY {
			hs.state = FIRST
		}

	case OBJECTS_MENU:

		if key == ESC_KEY {
			hs.state = FIRST
		}

	case LAST:

		if key == ESC_KEY {
			hs.state = FIRST
		}

	}

	//
	// Set prompt
	//  11 char per line at Regular9pt7b
	//
	switch hs.state {
	case FIRST:

		hs.dspOut = "1 Setup\n" +
			"2 Utility\n" +
			"3 Objects\n" +
			"\n"

		if !hs.isSetup {
			hs.dspOut = hs.dspOut + ">Not Setup<"
		}

	case SHOW_VERSION:
		hs.dspOut = "VERSION\n" + VERSION

	case SET_DATE:
		hs.dspOut = "Set Date\nYYYY-MM-DD\n-----------\n" + hs.currentDateStr

	case SET_DATE_Error:
		hs.dspOut = "Set Date\nYYYY-MM-DD\n-----------\n" + hs.currentDateStr + "\n>>MSG_ERROR<<"

	case SET_TIME:
		hs.dspOut = "Set Time\nHH:MM:SS+hh\n-----------\n" + hs.currentTimeStr

	case SET_TIME_MSG_ERROR:
		hs.dspOut = "Set Time\nHH:MM:SS+hh\n-----------\n" + hs.currentTimeStr + "\n>>MSG_ERROR<<"

	case SET_LATITUDE:
		hs.dspOut = "Set\nLatitude\n+DD.dddd\n-----------\n" + hs.locationLatitudeStr

	case SET_LONGITUDE:
		hs.dspOut = "Set\nLongitude\n+DD.dddd\n-----------\n" + hs.locationLongitudeStr

	case SET_ELEVATION:
		hs.dspOut = "Set\nElevation\n+DDDD\n-----------\n" + hs.locationElevationStr

	case UTILITY_MENU:
		hs.dspOut = "Utility\nMenu\n\nTODO\n"

	case OBJECTS_MENU:
		hs.dspOut = "Objects\nMenu\n\nTODO\n"

	case LAST:
		hs.dspOut = ">>END<<"

	default:
		hs.dspOut = "Bad State\n"

	}

	return hs.dspOut
}

func (hs *Handset) RenderScreen() {

	status := [10]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}

	// DEVTODO - try to do better than clearing the screen each time
	hs.screen.displayDevice.FillScreen(color.RGBA{0, 0, 0, 0})

	// Compute the status bar text
	status[0] = 'S'
	if hs.screen.direction {
		status[0] = 'N'
	}
	statusText := fmt.Sprintf("%s\n-----------", status)
	hs.screen.statusBarText = statusText

	// Status Bar
	tinyfont.WriteLine(
		hs.screen.displayDevice,
		hs.screen.font,
		3, 10,
		statusText,
		hs.screen.fontColor)

	// hs.screen.displayDevice.DrawFastHLine(5,5,150,hs.screen.fontColor)

	// Body
	tinyfont.WriteLine(
		hs.screen.displayDevice,
		hs.screen.font,
		3, 45,
		hs.screen.bodyText,
		hs.screen.fontColor)
}

func (hs *Handset) SetScreenBodyText(body string) {
	hs.screen.bodyText = body
}

// Util functions
func keyIsDigit(key Key) bool {
	switch key {
	case ZERO_KEY:
		return true
	case ONE_KEY:
		return true
	case TWO_KEY:
		return true
	case THREE_KEY:
		return true
	case FOUR_KEY:
		return true
	case FIVE_KEY:
		return true
	case SIX_KEY:
		return true
	case SEVEN_KEY:
		return true
	case EIGHT_KEY:
		return true
	case NINE_KEY:
		return true
	default:
		return false
	}
}

func doNav(key Key, state *State) bool {

	if key == ESC_KEY {

		if *state > FIRST {
			*state--
		}
		return true

	} else if key == ENTER_KEY {

		if *state < LAST {
			*state++
		}
		return true

	} else {

		return false
	}

}
