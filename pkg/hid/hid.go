package hid

import (
	"fmt"
	"image/color"
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/ssd1351"
	"tinygo.org/x/tinyfont"

	"github.com/tonygilkerson/astroeq/pkg/driver"
	"github.com/tonygilkerson/astroeq/pkg/grid"
	"github.com/tonygilkerson/astroeq/pkg/msg"
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
	ZERO State = iota
	FIRST
	SHOW_VERSION
	SET_DATE
	SET_TIME
	SET_LATITUDE
	SET_LONGITUDE
	SET_ELEVATION
	UTILITY_MENU
	SET_RA_TRACKING
	SET_RA_DIRECTION
	OBJECTS_MENU
	LAST
	SET_DATE_Error
	SET_TIME_MSG_ERROR
)

// Each button on the handset corresponds to one of the following Keys
type Key uint8

const (
	KEY_UNDEFINED Key = iota
	KEY_ZERO
	KEY_ONE
	KEY_TWO
	KEY_THREE
	KEY_FOUR
	KEY_FIVE
	KEY_SIX
	KEY_SEVEN
	KEY_EIGHT
	KEY_NINE

	KEY_SCROLL_DN
	KEY_SCROLL_UP

	KEY_RIGHT
	KEY_LEFT
	KEY_UP
	KEY_DOWN

	KEY_ESC
	KEY_SETUP
	KEY_ENTER
	KEY_REFRESH
)

// The Handset properties are maintained by the user via the handset.
// The user can perform basic CRUD operations on the Handset properties as well as
// use them in commands sent over the message bus.
type Handset struct {
	Screen       *Screen
	msgBroker    *msg.MsgBroker
	isSetup      bool
	state        State
	scrollDnKey  machine.Pin
	zeroKey      machine.Pin
	scrollKEY_UP machine.Pin

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
	grid.Grid
	displayDevice ssd1351.Device
	font          tinyfont.Font
	fontColor     color.RGBA
	BodyText      string
	// RA Data
	Tracking  driver.RaValue
	Direction driver.RaValue
	Position  uint32
}

// Returns a new Handset
func NewHandset(
	displayDevice ssd1351.Device,
	font tinyfont.Font,
	fontColor color.RGBA,
	displayRows int,
	displayCols int,
	msgBroker *msg.MsgBroker,
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
	scrollKEY_UP machine.Pin,

	rightKey machine.Pin,
	leftKey machine.Pin,
	upKey machine.Pin,
	downKey machine.Pin,

	escKey machine.Pin,
	setupKey machine.Pin,
	enterKey machine.Pin,
) (Handset, error) {

	var screen Screen
	// var grid grid.Grid  DEVTODO del me soon

	screen.displayDevice = displayDevice
	// screen.grid = grid    DEVTODO del me soon
	screen.ConfigureGrid(displayRows, displayCols)
	screen.font = font
	screen.fontColor = fontColor

	return Handset{
		Screen:               &screen,
		msgBroker:            msgBroker,
		isSetup:              false,
		state:                FIRST,
		scrollDnKey:          scrollDnKey,
		zeroKey:              zeroKey,
		scrollKEY_UP:         scrollKEY_UP,
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
	// Configure Key Pins
	//
	hs.scrollDnKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.zeroKey.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	hs.scrollKEY_UP.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
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
	hs.zeroKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_ZERO })
	hs.oneKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_ONE })
	hs.twoKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_TWO })
	hs.threeKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_THREE })
	hs.fourKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_FOUR })
	hs.fiveKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_FIVE })
	hs.sixKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_SIX })
	hs.sevenKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_SEVEN })
	hs.eightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_EIGHT })
	hs.nineKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_NINE })
	hs.scrollKEY_UP.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_SCROLL_UP })
	hs.scrollDnKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_SCROLL_DN })
	hs.rightKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_RIGHT })
	hs.leftKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_LEFT })
	hs.upKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_UP })
	hs.downKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_DOWN })
	hs.escKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_ESC })
	hs.setupKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_SETUP })
	hs.enterKey.SetInterrupt(machine.PinFalling, func(p machine.Pin) { hs.keyPressed = KEY_ENTER })

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
		if hs.keyPressed != KEY_UNDEFINED {

			//
			//  After a small delay if the key pressed has not changed, consider it "pressed"
			//
			key := hs.keyPressed
			time.Sleep(time.Millisecond * 150)
			// fmt.Printf("[publishKeysRoutine] value: %v \n", key)

			if key == hs.keyPressed {
				hs.keyStrokes <- hs.keyPressed
				hs.keyPressed = KEY_UNDEFINED //reset for next key press
			}
		}
		time.Sleep(time.Millisecond * 500)
	}

}

func (hs *Handset) GetKeyString(k Key) string {

	// fmt.Printf("[GetKeyString] value: %v \n", k)
	switch k {
	case KEY_ZERO:
		return "0"
	case KEY_ONE:
		return "1"
	case KEY_TWO:
		return "2"
	case KEY_THREE:
		return "3"
	case KEY_FOUR:
		return "4"
	case KEY_FIVE:
		return "5"
	case KEY_SIX:
		return "6"
	case KEY_SEVEN:
		return "7"
	case KEY_EIGHT:
		return "8"
	case KEY_NINE:
		return "9"
	case KEY_SCROLL_DN:
		return "ScrollDn"
	case KEY_SCROLL_UP:
		return "ScrollUp"
	case KEY_RIGHT:
		return "Right"
	case KEY_LEFT:
		return "Left"
	case KEY_UP:
		return "Up"
	case KEY_DOWN:
		return "Down"
	case KEY_ESC:
		return "ESC"
	case KEY_SETUP:
		return "Setup"
	case KEY_ENTER:
		return "Enter"
	case KEY_REFRESH:
		return "Refresh"
	default:
		return "Undefined"
	}
}

func (hs *Handset) GetKeyFromString(s string) Key {
	switch s {
	case "0":
		return KEY_ZERO
	case "1":
		return KEY_ONE
	case "2":
		return KEY_TWO
	case "3":
		return KEY_THREE
	case "4":
		return KEY_FOUR
	case "5":
		return KEY_FIVE
	case "6":
		return KEY_SIX
	case "7":
		return KEY_SEVEN
	case "8":
		return KEY_EIGHT
	case "9":
		return KEY_NINE
	case "ScrollDn":
		return KEY_SCROLL_DN
	case "ScrollUp":
		return KEY_SCROLL_UP
	case "Right":
		return KEY_RIGHT
	case "Left":
		return KEY_LEFT
	case "Up":
		return KEY_UP
	case "Down":
		return KEY_DOWN
	case "ESC":
		return KEY_ESC
	case "Setup":
		return KEY_SETUP
	case "Enter":
		return KEY_ENTER
	case "Refresh":
		return KEY_REFRESH
	default:
		return KEY_UNDEFINED
	}
}

func (hs *Handset) StateMachine(key Key) string {

	switch hs.state {

	case ZERO:
		doNav(key, &hs.state)

	case FIRST:

		if !doNav(key, &hs.state) {
			if key == KEY_ONE {
				hs.state++
			} else if key == KEY_TWO {
				hs.state = UTILITY_MENU
			} else if key == KEY_THREE {
				hs.state = OBJECTS_MENU
			}
		}

	case SHOW_VERSION:
		hs.isSetup = false
		doNav(key, &hs.state)

	case SET_RA_TRACKING:

		if !doNav(key, &hs.state) {
			if key == KEY_ONE {
				hs.msgBroker.PublishRACmdSetTracking(driver.RA_TRACKING_ON)
				hs.state++
			} else if key == KEY_TWO {
				hs.msgBroker.PublishRACmdSetTracking(driver.RA_TRACKING_OFF)
				hs.state++
			}
		}

	case SET_RA_DIRECTION:

		if !doNav(key, &hs.state) {
			if key == KEY_ONE {
				hs.msgBroker.PublishRACmdSetDirection(driver.RA_DIRECTION_NORTH)
				hs.state = UTILITY_MENU
			} else if key == KEY_TWO {
				hs.msgBroker.PublishRACmdSetDirection(driver.RA_DIRECTION_SOUTH)
				hs.state = UTILITY_MENU
			}
		}

	case SET_DATE:

		if key == KEY_ENTER {
			var err error
			// RFC3339 example: "2006-01-02T15:04:05+05:00"
			hs.currentTime, err = time.Parse(time.RFC3339, hs.currentDateStr+"T10:00:00+00:00")

			if err != nil {
				hs.state = SET_DATE_Error
			} else {
				hs.state = SET_DATE + 1
			}

		} else if !doNav(key, &hs.state) {

			if key == KEY_LEFT && len(hs.currentDateStr) > 0 {
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
		if key == KEY_ESC {
			hs.state = SET_DATE
		}

	case SET_TIME:

		if key == KEY_ENTER {
			var err error
			// RFC3339 example: "2006-01-02T15:04:05+05:00"
			hs.currentTime, err = time.Parse(time.RFC3339, hs.currentDateStr+"T"+hs.currentTimeStr+":00")

			if err != nil {
				hs.state = SET_TIME_MSG_ERROR
			} else {
				hs.state = SET_TIME + 1
			}

		} else if !doNav(key, &hs.state) {

			if key == KEY_LEFT && len(hs.currentTimeStr) > 0 {
				hs.currentTimeStr = hs.currentTimeStr[:len(hs.currentTimeStr)-1]

			} else if len(hs.currentTimeStr) == 2 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + ":" + hs.GetKeyString(key)

			} else if len(hs.currentTimeStr) == 5 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + ":" + hs.GetKeyString(key)

			} else if len(hs.currentTimeStr) == 8 {

				if key == KEY_SCROLL_DN {
					hs.currentTimeStr = hs.currentTimeStr + "-"
				} else if key == KEY_SCROLL_UP {
					hs.currentTimeStr = hs.currentTimeStr + "+"
				}

			} else if len(hs.currentTimeStr) < 11 && keyIsDigit(key) {
				hs.currentTimeStr = hs.currentTimeStr + hs.GetKeyString(key)

			}
		}

	case SET_TIME_MSG_ERROR:
		if key == KEY_ESC {
			hs.state = SET_TIME
		}

	case SET_LATITUDE:

		if !doNav(key, &hs.state) {

			if key == KEY_LEFT && len(hs.locationLatitudeStr) > 0 {
				hs.locationLatitudeStr = hs.locationLatitudeStr[:len(hs.locationLatitudeStr)-1]

			} else if len(hs.locationLatitudeStr) == 3 && keyIsDigit(key) {
				hs.locationLatitudeStr = hs.locationLatitudeStr + "." + hs.GetKeyString(key)

			} else if len(hs.locationLatitudeStr) == 0 {

				if key == KEY_SCROLL_DN {
					hs.locationLatitudeStr = "-"
				} else if key == KEY_SCROLL_UP {
					hs.locationLatitudeStr = "+"
				}

			} else if len(hs.locationLatitudeStr) < 8 && keyIsDigit(key) {
				hs.locationLatitudeStr = hs.locationLatitudeStr + hs.GetKeyString(key)

			}
		}

	case SET_LONGITUDE:

		if !doNav(key, &hs.state) {

			if key == KEY_LEFT && len(hs.locationLongitudeStr) > 0 {
				hs.locationLongitudeStr = hs.locationLongitudeStr[:len(hs.locationLongitudeStr)-1]

			} else if len(hs.locationLongitudeStr) == 3 && keyIsDigit(key) {
				hs.locationLongitudeStr = hs.locationLongitudeStr + "." + hs.GetKeyString(key)

			} else if len(hs.locationLongitudeStr) == 0 {

				if key == KEY_SCROLL_DN {
					hs.locationLongitudeStr = "-"
				} else if key == KEY_SCROLL_UP {
					hs.locationLongitudeStr = "+"
				}

			} else if len(hs.locationLongitudeStr) < 8 && keyIsDigit(key) {
				hs.locationLongitudeStr = hs.locationLongitudeStr + hs.GetKeyString(key)

			}
		}

	case SET_ELEVATION:

		if key == KEY_ENTER {
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

			if key == KEY_LEFT && len(hs.locationElevationStr) > 0 {
				hs.locationElevationStr = hs.locationElevationStr[:len(hs.locationElevationStr)-1]

			} else if len(hs.locationElevationStr) == 0 {

				if key == KEY_SCROLL_DN {
					hs.locationElevationStr = "-"
				} else if key == KEY_SCROLL_UP {
					hs.locationElevationStr = "+"
				}

			} else if len(hs.locationElevationStr) < 5 && keyIsDigit(key) {
				hs.locationElevationStr = hs.locationElevationStr + hs.GetKeyString(key)

			}
		}

	case UTILITY_MENU:

		if key == KEY_ESC {
			hs.state = FIRST
		} else if key == KEY_ONE {
			hs.state++
		}

	case OBJECTS_MENU:

		if key == KEY_ESC {
			hs.state = FIRST
		}

	case LAST:

		if key == KEY_ESC {
			hs.state = FIRST
		}

	}

	//
	// Set prompt
	//  11 char per line at Regular9pt7b
	//
	switch hs.state {
	case ZERO:
		hs.dspOut = fmt.Sprintf("RA: %v", hs.Screen.Position)

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

	case SET_RA_TRACKING:
		hs.dspOut = "RA Tracking\n1 - On\n2 - Off\n" + string(hs.Screen.Direction)

	case SET_RA_DIRECTION:
		hs.dspOut = "RA Dir\n1 - North\n2 - South\n" + string(hs.Screen.Direction)

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
		hs.dspOut = "1 RA Setup\n" +
			"2 todo\n" +
			"\n"

	case OBJECTS_MENU:
		hs.dspOut = "Objects\nMenu\n\nTODO\n"

	case LAST:
		hs.dspOut = ">>END<<"

	default:
		hs.dspOut = "Bad State\n"

	}

	return hs.dspOut
}
func (hs *Handset) GetStatusLine() string {
	status := []byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}

	//
	// Compute the status bar text
	//

	if hs.Screen.Tracking == driver.RA_TRACKING_ON {
		// On - tracking
		status[0] = 'T'
	} else {
		// Off - not tracking
		status[0] = 'X'
	}

	if hs.Screen.Direction == driver.RA_DIRECTION_NORTH {
		// Direction = North
		status[1] = 'N'
	} else {
		// Direction = South
		status[1] = 'S'
	}

	return string(status)
}

func (hs *Handset) RenderScreen() {

	hs.Screen.LoadGrid(hs.GetStatusLine() + "\n-----------\n" + hs.Screen.BodyText)
	hs.Screen.WriteLines()

}

func (screen *Screen) WriteLines() {

	var x, y int16
	black := color.RGBA{0, 0, 0, 255}

	for r, row := range screen.GetCells() {
		for c, col := range row {
			cell := col
			// DEVTODO might need to add width and height back, also the x and y seem backward do I have the screen rotated
			//         undo the hard code when you figure it out
			x = int16(10*c) + 10
			y = int16(15*r) + 15

			// x = int16(screen.grid.GetWidth()*c) + 5
			// y = int16(screen.grid.GetHeight()*r) + 20
			if cell.IsDirty() {
				cells := screen.GetCells()
				tinyfont.WriteLine(&screen.displayDevice, &screen.font, x, y, string(cells[r][c].GetPrevChar()), black) // erase the previous character
				tinyfont.WriteLine(&screen.displayDevice, &screen.font, x, y, string(cells[r][c].GetChar()), screen.fontColor)
			}
		}
	}
}

// Util functions
func keyIsDigit(key Key) bool {
	switch key {
	case KEY_ZERO:
		return true
	case KEY_ONE:
		return true
	case KEY_TWO:
		return true
	case KEY_THREE:
		return true
	case KEY_FOUR:
		return true
	case KEY_FIVE:
		return true
	case KEY_SIX:
		return true
	case KEY_SEVEN:
		return true
	case KEY_EIGHT:
		return true
	case KEY_NINE:
		return true
	default:
		return false
	}
}

func doNav(key Key, state *State) bool {

	if key == KEY_ESC {

		if *state > ZERO {
			*state--
		}
		return true

	} else if key == KEY_ENTER {

		if *state < LAST {
			*state++
		}
		return true

	} else {

		return false
	}

}
