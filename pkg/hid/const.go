package hid

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
