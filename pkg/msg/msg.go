package msg

import (
	"fmt"
	"machine"
	"strconv"
	"strings"
	"time"

	"github.com/tonygilkerson/astroeq/pkg/driver"
)

// Define message types
type MsgType string
type LogLevel string
type RADriverCmd string

const (
	MSG_FOO      MsgType = "Foo"
	MSG_LOG      MsgType = "Log"
	MSG_HANDSET  MsgType = "Handset"
	MSG_RADRIVER MsgType = "RADriver"
)

const (
	MSG_DEBUG LogLevel = "Debug"
	MSG_INFO  LogLevel = "Info"
	MSG_WARN  LogLevel = "Warn"
	MSG_ERROR LogLevel = "Error"
)

const (
	RA_CMD_INFO          RADriverCmd = "INFO"
	RA_CMD_SET_DIR_NORTH RADriverCmd = "SetDirNorth"
	RA_CMD_SET_DIR_SOUTH RADriverCmd = "SetDirSouth"
	RA_CMD_TRACKING_ON   RADriverCmd = "SetTrackingOn"
	RA_CMD_TRACKING_OFF  RADriverCmd = "SetTrackingOff"
)

// Foo message use for testing I will delete it eventually
// The following is a sample message that can be sent over UART
//
//	^Foo|This is a foo message~
type FooMsg struct {
	Kind MsgType
	Name string
}
type LogMsg struct {
	Kind   MsgType
	Level  LogLevel
	Source string
	Body   string
}
type HandsetMsg struct {
	Kind MsgType
	Keys []string
}

// DEVTODO - Creat RADriverMsg and RADriverCmd
// DEVTODO - combine SetDirNorth and SetDirSouth into SetDirection and have it take as a parm driver.RA_DIR_NORTH
// RA Driver message used for sending commands to the RA Driver and for publishing it current status
// The following are sample messages
//
// ^RADriver|SetDirNorth~
// ^RADriver|SetDirSouth~
// ^RADriver|INFO|true|North|12345~
type RADriverMsg struct {
	Kind      MsgType
	Cmd       RADriverCmd
	Tracking  bool
	Direction driver.RaDirection
	Position  uint32
}

type MsgInterface interface {
	FooMsg | LogMsg | HandsetMsg | RADriverMsg
}

type UART interface {
	Configure(config machine.UARTConfig) error
	Buffered() int
	ReadByte() (byte, error)
	Write(data []byte) (n int, err error)
}

// Message Broker
type MsgBroker struct {
	logLevel LogLevel

	uartUp      UART
	uartUpTxPin machine.Pin
	uartUpRxPin machine.Pin

	uartDn      UART
	uartDnTxPin machine.Pin
	uartDnRxPin machine.Pin

	fooCh      chan FooMsg
	logCh      chan LogMsg
	handsetCh  chan HandsetMsg
	raDriverCh chan RADriverMsg
}

func NewBroker(
	uartUp UART,
	uartUpTxPin machine.Pin,
	uartUpRxPin machine.Pin,

	uartDn UART,
	uartDnTxPin machine.Pin,
	uartDnRxPin machine.Pin,

) (MsgBroker, error) {

	return MsgBroker{
		logLevel: MSG_INFO, // default

		uartUp:      uartUp,
		uartUpTxPin: uartUpTxPin,
		uartUpRxPin: uartUpRxPin,

		uartDn:      uartDn,
		uartDnTxPin: uartDnTxPin,
		uartDnRxPin: uartDnRxPin,

		fooCh:      nil,
		logCh:      nil,
		handsetCh:  nil,
		raDriverCh: nil,
	}, nil

}

func (mb *MsgBroker) Configure() {

	// Upstream UART
	if mb.uartUp != nil {
		mb.uartUp.Configure(machine.UARTConfig{TX: mb.uartUpTxPin, RX: mb.uartUpRxPin})
	}

	// Downstream UART
	if mb.uartDn != nil {
		mb.uartDn.Configure(machine.UARTConfig{TX: mb.uartDnTxPin, RX: mb.uartDnRxPin})
	}
}

func (mb *MsgBroker) SetFooCh(ch chan FooMsg) {
	mb.fooCh = ch
}
func (mb *MsgBroker) SetHandsetCh(ch chan HandsetMsg) {
	mb.handsetCh = ch
}
func (mb *MsgBroker) SetLogCh(ch chan LogMsg) {
	mb.logCh = ch
}
func (mb *MsgBroker) SetRADriverCh(ch chan RADriverMsg) {
	mb.raDriverCh = ch
}

// Look for messages that look like this
//
//	^Log|Info|HID|A log message from the HID~
func (mb *MsgBroker) SubscriptionReaderRoutine() {

	//
	// Look for start of message loop
	//
	for {

		// If no data wait and try again
		if mb.uartUp.Buffered() == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		data, _ := mb.uartUp.ReadByte()

		// the "^" character is the start of a message
		if data == 94 {
			message := make([]byte, 0, 255) //capacity of 255

			//
			// Start loop read a message
			//
			for {

				// If no data wait and try again
				if mb.uartUp.Buffered() == 0 {
					time.Sleep(time.Millisecond * 1)
					continue
				}

				// the "~" character is the end of a message
				data, _ := mb.uartUp.ReadByte()

				if data == 126 {
					break
				} else {
					message = append(message, data)
				}

			}

			//
			// At this point we have an entire message, so dispatch it!
			//
			msgParts := strings.Split(string(message[:]), "|")
			mb.DispatchMsgToChannel(msgParts)

		}
	}
}

func (mb *MsgBroker) DispatchMsgToChannel(msgParts []string) {

	switch msgParts[0] {

	case string(MSG_FOO):
		fmt.Printf("[DispatchMsgToChannel] - %v\n", MSG_FOO)
		msg := unmarshallFoo(msgParts)
		if mb.fooCh != nil {
			mb.fooCh <- *msg
		}
	case string(MSG_LOG):
		fmt.Printf("[DispatchMsgToChannel] - %v\n", MSG_LOG)
		msg := unmarshallLog(msgParts)
		if mb.logCh != nil {
			mb.logCh <- *msg
		}
	case string(MSG_HANDSET):
		fmt.Printf("[DispatchMsgToChannel] - %v\n", MSG_HANDSET)
		msg := unmarshallHandset(msgParts)
		if mb.handsetCh != nil {
			mb.handsetCh <- *msg
		}
	case string(MSG_RADRIVER):
		fmt.Printf("[DispatchMsgToChannel] - %v\n", MSG_RADRIVER)
		msg := unmarshallRADriver(msgParts)
		if mb.raDriverCh != nil {
			mb.raDriverCh <- *msg
		}
	default:
		fmt.Println("[DispatchMsgToChannel] - no match found")
	}

}

func (mb *MsgBroker) SetLogLevel(level LogLevel) {
	mb.logLevel = level
}

func (mb *MsgBroker) isLoggable(level LogLevel) bool {

	switch level {
	case MSG_DEBUG:
		return true

	case MSG_INFO:
		if mb.logLevel == MSG_ERROR || mb.logLevel == MSG_WARN || mb.logLevel == MSG_INFO {
			return true
		}
	case MSG_WARN:
		if mb.logLevel == MSG_ERROR || mb.logLevel == MSG_WARN {
			return true
		}
	case MSG_ERROR:
		if mb.logLevel == MSG_ERROR {
			return true
		}

	}

	return false

}

func (mb *MsgBroker) PublishFoo(foo FooMsg) {

	msgStr := "^" + string(foo.Kind)
	msgStr = msgStr + "|" + string(foo.Name) + "~"

	mb.PublishMsg(msgStr)

}

func (mb *MsgBroker) PublishLog(log LogMsg) {

	// If not loggable do nothing
	if !mb.isLoggable(log.Level) {
		return
	}

	msgStr := "^" + string(log.Kind)
	msgStr = msgStr + "|" + string(log.Level)
	msgStr = msgStr + "|" + string(log.Source)
	msgStr = msgStr + "|" + string(log.Body) + "~"

	mb.PublishMsg(msgStr)

}

// DEVTODO - I delete PublishHandset soon
// func (mb *MsgBroker) PublishHandset(handsetMsg HandsetMsg) {

// 	msgStr := "^" + string(handsetMsg.Kind)

// 	for _, key := range handsetMsg.Keys {
// 		msgStr = msgStr + "|" + key
// 	}
// 	msgStr = msgStr + "~"

// 	mb.PublishMsg(msgStr)

// }
func (mb *MsgBroker) PublishRADriver(raDriverMsg RADriverMsg) {

	msgStr := "^" + string(raDriverMsg.Kind)
	msgStr = msgStr + "|" + fmt.Sprintf("%v", raDriverMsg.Cmd)
	msgStr = msgStr + "|" + fmt.Sprintf("%v", raDriverMsg.Tracking)
	msgStr = msgStr + "|" + fmt.Sprintf("%v", raDriverMsg.Direction)
	msgStr = msgStr + "|" + fmt.Sprintf("%v", raDriverMsg.Position) + "~"

	mb.PublishMsg(msgStr)

}

func (mb *MsgBroker) PublishRACmdSetDir(direction driver.RaDirection) {

	msgStr := "^" + string(MSG_RADRIVER)
	if direction == driver.RA_DIR_NORTH {
		msgStr = msgStr + "|" + string(RA_CMD_SET_DIR_NORTH)
		msgStr = msgStr + "|" + string(driver.RA_DIR_NORTH) + "~"
	} else {
		msgStr = msgStr + "|" + string(RA_CMD_SET_DIR_SOUTH)
		msgStr = msgStr + "|" + string(driver.RA_DIR_SOUTH) + "~"
	}

	mb.PublishMsg(msgStr)

}
func (mb *MsgBroker) PublishRACmdSetTracking(tracking bool) {

	msgStr := "^" + string(MSG_RADRIVER)
	if tracking {
		msgStr = msgStr + "|" + string(RA_CMD_TRACKING_ON) + "~"
	} else {
		msgStr = msgStr + "|" + string(RA_CMD_TRACKING_OFF) + "~"
	}

	mb.PublishMsg(msgStr)

}

func (mb *MsgBroker) PublishMsg(msg string) {

	if mb.uartUp != nil {
		mb.uartUp.Write([]byte(msg))
		// Print a new line between messages for readability in the serial monitor
		mb.uartUp.Write([]byte("\n"))
	}

	if mb.uartDn != nil {
		mb.uartDn.Write([]byte(msg))
		// Print a new line between messages for readability in the serial monitor
		mb.uartDn.Write([]byte("\n"))
	}
}

func unmarshallFoo(msgParts []string) *FooMsg {

	fooMsg := new(FooMsg)

	if len(msgParts) > 0 {
		fooMsg.Kind = MSG_FOO
	}
	if len(msgParts) > 1 {
		fooMsg.Name = msgParts[1]
	}

	return fooMsg
}

func unmarshallLog(msgParts []string) *LogMsg {

	logMsg := new(LogMsg)

	if len(msgParts) > 0 {
		logMsg.Kind = MSG_LOG
	}
	if len(msgParts) > 1 {
		logMsg.Level = LogLevel(msgParts[1])
	}
	if len(msgParts) > 2 {
		logMsg.Source = msgParts[2]
	}
	if len(msgParts) > 3 {
		logMsg.Body = msgParts[3]
	}

	return logMsg
}

func unmarshallHandset(msgParts []string) *HandsetMsg {

	handsetMsg := new(HandsetMsg)

	if len(msgParts) > 0 {
		handsetMsg.Kind = MSG_HANDSET
	}
	if len(msgParts) > 1 {
		handsetMsg.Keys = msgParts[1:]
	}

	return handsetMsg
}

func unmarshallRADriver(msgParts []string) *RADriverMsg {

	raDriverMsg := new(RADriverMsg)

	if len(msgParts) > 0 {
		raDriverMsg.Kind = MSG_RADRIVER
	}
	if len(msgParts) > 1 {
		raDriverMsg.Cmd = RADriverCmd(msgParts[1])
	}

	if len(msgParts) > 2 {
		if RADriverCmd(msgParts[2]) == "North" {
			raDriverMsg.Direction = driver.RA_DIR_NORTH
		} else {
			raDriverMsg.Direction = driver.RA_DIR_SOUTH
		}
	}
	if len(msgParts) > 3 {
		raDriverMsg.Direction = driver.RaDirection(msgParts[3])
	}

	if len(msgParts) > 4 {
		p, _ := strconv.Atoi(msgParts[4])
		raDriverMsg.Position = uint32(p)
	}

	return raDriverMsg
}

func (mb *MsgBroker) InfoLog(src string, body string) {
	mb.PublishLog(makeLogMsg(MSG_INFO, src, body))
}
func makeLogMsg(level LogLevel, src string, body string) LogMsg {

	var logMsg LogMsg
	logMsg.Kind = MSG_LOG
	logMsg.Level = level
	logMsg.Source = src
	logMsg.Body = body

	return logMsg
}
