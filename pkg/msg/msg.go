package msg

import (
	"fmt"
	"machine"
	"strconv"
	"strings"
	"time"

	"github.com/tonygilkerson/astroeq/pkg/driver"
)

const (
	TOKEN_HAT   byte = 94  // ^
	TOKEN_ABOUT byte = 126 // ~
)

// Define message types
type MsgType string
type RADriverCmd string

const (
	MSG_FOO          MsgType = "Foo"
	MSG_HANDSET      MsgType = "Handset"
	MSG_RADRIVER     MsgType = "RADriver"
	MSG_RADRIVER_CMD MsgType = "RADriverCmd"
)

const (
	RA_CMD_SET_TRACKING  RADriverCmd = "SetTracking"
	RA_CMD_SET_DIRECTION RADriverCmd = "SetDirection"
)

// Foo message use for testing I will delete it eventually
// The following is a sample message that can be sent over UART
//
//	^Foo|This is a foo message~
type FooMsg struct {
	Kind MsgType
	Name string
}

// ^Handset|somekey~
type HandsetMsg struct {
	Kind MsgType
	Keys []string
}

// RA Driver message used for sending commands to the RA Driver and for publishing it current status
// The following are sample messages
//
// ^RADriver|On|North|12345~
type RADriverMsg struct {
	Kind      MsgType
	Tracking  driver.RaValue
	Direction driver.RaValue
	Position  uint32
}

// ^RADriverCmd|SetTracking|On~
// ^RADriverCmd|SetTracking|Off~
// ^RADriverCmd|SetDirection|North~
// ^RADriverCmd|SetDirection|South~
type RADriverCmdMsg struct {
	Kind MsgType
	Cmd  RADriverCmd
	Args []string
}

type MsgInterface interface {
	FooMsg | HandsetMsg | RADriverMsg | RADriverCmdMsg
}

type UART interface {
	Configure(config machine.UARTConfig) error
	Buffered() int
	ReadByte() (byte, error)
	Write(data []byte) (n int, err error)
}

// Message Broker
type MsgBroker struct {
	uartUp      UART
	uartUpTxPin machine.Pin
	uartUpRxPin machine.Pin

	uartDn      UART
	uartDnTxPin machine.Pin
	uartDnRxPin machine.Pin

	fooCh         chan FooMsg
	handsetCh     chan HandsetMsg
	raDriverCh    chan RADriverMsg
	raDriverCmdCh chan RADriverCmdMsg
}

func NewBroker(
	uartUp UART,
	uartUpTxPin machine.Pin,
	uartUpRxPin machine.Pin,

	uartDn UART,
	uartDnTxPin machine.Pin,
	uartDnRxPin machine.Pin,

) (MsgBroker, error) {

	var mb MsgBroker

	if uartUp != nil {
		mb.uartUp = uartUp
		mb.uartUpTxPin = uartUpTxPin
		mb.uartUpRxPin = uartUpRxPin
	}

	if uartDn != nil {
		mb.uartDn = uartDn
		mb.uartDnTxPin = uartDnTxPin
		mb.uartDnRxPin = uartDnRxPin
	}

	return mb, nil

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
func (mb *MsgBroker) SetRADriverCh(ch chan RADriverMsg) {
	mb.raDriverCh = ch
}
func (mb *MsgBroker) SetRADriverCmdCh(ch chan RADriverCmdMsg) {
	mb.raDriverCmdCh = ch
}

func (mb *MsgBroker) SubscriptionReaderRoutine() {

	for {

		mb.uartReader(mb.uartUp, mb.uartDn)
		time.Sleep(time.Millisecond * 100)

		mb.uartReader(mb.uartDn, mb.uartUp)
		time.Sleep(time.Millisecond * 100)
	}
}

func (mb *MsgBroker) uartReader(readFromUart UART, forwardToUart UART) {

	// If no data, quit
	if readFromUart == nil || readFromUart.Buffered() == 0 {
		return
	}

	data, _ := readFromUart.ReadByte()

	// the "^" character is the start of a message
	if data == TOKEN_HAT {
		message := make([]byte, 0, 255) //capacity of 255

		//
		// Start loop read a message
		//
		var abortCount int

		for {

			// If we see the start of a message but don't get the end in less than a second then abort
			if abortCount > 100 {
				return
			}

			// If no data wait and try again
			if readFromUart.Buffered() == 0 {
				abortCount++
				time.Sleep(time.Millisecond * 10)
				continue
			}

			// the "~" character is the end of a message
			data, _ := readFromUart.ReadByte()

			if data == TOKEN_ABOUT {
				break
			} else {
				message = append(message, data)
			}

		}

		//
		// At this point we have an entire message, so dispatch it!
		//
		msgParts := strings.Split(string(message), "|")

		mb.DispatchMsgToChannel(msgParts)

		// Forward message for other potential consumers
		if forwardToUart != nil {

			// rewrap the message to start with ^ and end with ~
			message = append([]byte{TOKEN_HAT}, message...)
			message = append(message, TOKEN_ABOUT)

			forwardToUart.Write(message)

		}

	}
}

func (mb *MsgBroker) DispatchMsgToChannel(msgParts []string) {

	switch msgParts[0] {

	case string(MSG_FOO):
		fmt.Printf("[DispatchMsgToChannel] - %v\n", MSG_FOO)
		msg := makeFoo(msgParts)
		if mb.fooCh != nil {
			mb.fooCh <- *msg
		}
	case string(MSG_HANDSET):
		fmt.Printf("[DispatchMsgToChannel] - %v\n", MSG_HANDSET)
		msg := makeHandset(msgParts)
		if mb.handsetCh != nil {
			mb.handsetCh <- *msg
		}
	case string(MSG_RADRIVER):
		fmt.Printf("[DispatchMsgToChannel] - %v\n", MSG_RADRIVER)
		msg := makeRADriver(msgParts)
		if mb.raDriverCh != nil {
			mb.raDriverCh <- *msg
		}
	case string(MSG_RADRIVER_CMD):
		fmt.Printf("[DispatchMsgToChannel] - %v\n", MSG_RADRIVER_CMD)
		msg := makeRADriverCmd(msgParts)
		if mb.raDriverCmdCh != nil {
			mb.raDriverCmdCh <- *msg
		}
	default:
		fmt.Println("[DispatchMsgToChannel] - no match found")
	}

}

func (mb *MsgBroker) PublishFoo(foo FooMsg) {

	msgStr := "^" + string(foo.Kind)
	msgStr = msgStr + "|" + string(foo.Name) + "~"

	mb.PublishMsg(msgStr)

}

func (mb *MsgBroker) PublishRADriver(raDriverMsg RADriverMsg) {

	msgStr := "^" + string(raDriverMsg.Kind)
	msgStr = msgStr + "|" + fmt.Sprintf("%v", raDriverMsg.Tracking)
	msgStr = msgStr + "|" + fmt.Sprintf("%v", raDriverMsg.Direction)
	msgStr = msgStr + "|" + fmt.Sprintf("%v", raDriverMsg.Position) + "~"

	mb.PublishMsg(msgStr)

}

func (mb *MsgBroker) PublishRADriverCmd(raDriverCmdMsg RADriverCmdMsg) {

	msgStr := "^" + string(raDriverCmdMsg.Kind)
	msgStr = msgStr + "|" + fmt.Sprintf("%v", raDriverCmdMsg.Cmd)
	msgStr = msgStr + "|" + strings.Join(raDriverCmdMsg.Args, ",") + "~"

	mb.PublishMsg(msgStr)

}

func (mb *MsgBroker) PublishRACmdSetDirection(direction driver.RaValue) {
	var raCmdMsg RADriverCmdMsg

	raCmdMsg.Kind = MSG_RADRIVER_CMD
	raCmdMsg.Cmd = RA_CMD_SET_DIRECTION

	if direction == driver.RA_DIRECTION_NORTH {
		raCmdMsg.Args = append(raCmdMsg.Args, string(driver.RA_DIRECTION_NORTH))
	} else {
		raCmdMsg.Args = append(raCmdMsg.Args, string(driver.RA_DIRECTION_SOUTH))
	}

	mb.PublishRADriverCmd(raCmdMsg)

}

func (mb *MsgBroker) PublishRACmdSetTracking(tracking driver.RaValue) {
	var raCmdMsg RADriverCmdMsg

	raCmdMsg.Kind = MSG_RADRIVER_CMD
	raCmdMsg.Cmd = RA_CMD_SET_TRACKING
	raCmdMsg.Args = append(raCmdMsg.Args, string(tracking))

	mb.PublishRADriverCmd(raCmdMsg)

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

func makeFoo(msgParts []string) *FooMsg {

	fooMsg := new(FooMsg)

	if len(msgParts) > 0 {
		fooMsg.Kind = MSG_FOO
	}
	if len(msgParts) > 1 {
		fooMsg.Name = msgParts[1]
	}

	return fooMsg
}

func makeHandset(msgParts []string) *HandsetMsg {

	handsetMsg := new(HandsetMsg)

	if len(msgParts) > 0 {
		handsetMsg.Kind = MSG_HANDSET
	}
	if len(msgParts) > 1 {
		handsetMsg.Keys = msgParts[1:]
	}

	return handsetMsg
}

func makeRADriver(msgParts []string) *RADriverMsg {

	// DEVTODO - I need a way to make the compiler complain when this does not match the struct

	raDriverMsg := new(RADriverMsg)

	if len(msgParts) > 0 {
		raDriverMsg.Kind = MSG_RADRIVER
	}

	if len(msgParts) > 1 {
		// index 1 is "On" of "Off"
		tracking := driver.RaValue(msgParts[1])
		raDriverMsg.Tracking = tracking
	}

	if len(msgParts) > 2 {
		if RADriverCmd(msgParts[2]) == "North" {
			raDriverMsg.Direction = driver.RA_DIRECTION_NORTH
		} else {
			raDriverMsg.Direction = driver.RA_DIRECTION_SOUTH
		}
	}

	if len(msgParts) > 3 {
		p, _ := strconv.Atoi(msgParts[3])
		raDriverMsg.Position = uint32(p)
	}

	return raDriverMsg
}
func makeRADriverCmd(msgParts []string) *RADriverCmdMsg {

	raDriverCmdMsg := new(RADriverCmdMsg)

	if len(msgParts) > 0 {
		raDriverCmdMsg.Kind = MSG_RADRIVER_CMD
	}
	if len(msgParts) > 1 {
		raDriverCmdMsg.Cmd = RADriverCmd(msgParts[1])
	}

	if len(msgParts) > 2 {
		raDriverCmdMsg.Args = strings.Split(msgParts[2], ",")
	}

	return raDriverCmdMsg
}
