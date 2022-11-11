package msg

import (
	"fmt"
	"machine"
	"reflect"
	"strings"
	"time"
)

// Define message types
type MsgType string
type LogLevel string

const (
	Foo MsgType = "Foo"
	Log MsgType = "Log"
)

const (
	Debug LogLevel = "Debug"
	Info  LogLevel = "Info"
	Warn  LogLevel = "Warn"
	Error LogLevel = "Error"
)

// DEVTODO delete foo message eventually
type LogMsg struct {
	Kind   MsgType
	Level  LogLevel
	Source string
	Body   string
}
type FooMsg struct {
	Kind MsgType
	Name string
}

type MsgInterface interface {
	FooMsg | LogMsg
}

// Message Broker
type MsgBroker struct {
	logLevel LogLevel

	uartUp      *machine.UART
	uartUpTxPin machine.Pin
	uartUpRxPin machine.Pin

	uartDn      *machine.UART
	uartDnTxPin machine.Pin
	uartDnRxPin machine.Pin

	fooCh chan FooMsg
	logCh chan LogMsg
}

func NewBroker(

	uartUp *machine.UART,
	uartUpTxPin machine.Pin,
	uartUpRxPin machine.Pin,

	uartDn *machine.UART,
	uartDnTxPin machine.Pin,
	uartDnRxPin machine.Pin,
) (MsgBroker, error) {

	return MsgBroker{
		logLevel: Warn, // By default broker Warn and Error

		uartUp:      uartUp,
		uartUpTxPin: uartUpTxPin,
		uartUpRxPin: uartUpRxPin,

		uartDn:      uartDn,
		uartDnTxPin: uartDnTxPin,
		uartDnRxPin: uartDnRxPin,

		fooCh: nil,
		logCh: nil,
	}, nil

}

func (mb *MsgBroker) Configure() {
	// Upstream UART
	mb.uartUp.Configure(machine.UARTConfig{TX: mb.uartUpTxPin, RX: mb.uartUpRxPin})

	// Downstream UART
	mb.uartDn.Configure(machine.UARTConfig{TX: mb.uartUpTxPin, RX: mb.uartUpRxPin})
}

func (mb *MsgBroker) SetFooCh(c chan FooMsg) {
	mb.fooCh = c
}
func (mb *MsgBroker) SetLogCh(c chan LogMsg) {
	mb.logCh = c
}

// Look for messages that look like this
//
//	^Log|Info|HID|A log message from the HID~
func (mb *MsgBroker) SubscriptionReader() {

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

				// the "~" character it the end of a message
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
			mb.DispatchMessage(msgParts)

		}
	}
}

func (mb *MsgBroker) DispatchMessage(msgParts []string) {

	switch msgParts[0] {
	case string(Foo):
		fmt.Println("[DispatchMessage] - Foo")
		msg := unmarshallFoo(msgParts)
		if mb.fooCh != nil {
			mb.fooCh <- *msg
		}
	case string(Log):
		fmt.Println("[DispatchMessage] - Log")
		msg := unmarshallLog(msgParts)
		if mb.logCh != nil {
			mb.logCh <- *msg
		}
	default:
		fmt.Println("[DispatchMessage] - default")
	}

}

func (mb *MsgBroker) SetLogLevel(level LogLevel) {
	mb.logLevel = level
}

func (mb *MsgBroker) isLoggable(level LogLevel) bool {

	switch level {
	case Debug:
		return true

	case Info:
		if mb.logLevel == Error || mb.logLevel == Warn || mb.logLevel == Info {
			return true
		}
	case Warn:
		if mb.logLevel == Error || mb.logLevel == Warn {
			return true
		}
	case Error:
		if mb.logLevel == Error {
			return true
		}

	}

	return false

}

func PublishMsg[M MsgInterface](m M, mb MsgBroker) {

	//
	// reflect to get message properties
	//
	msg := reflect.ValueOf(&m).Elem()
	msgKind := fmt.Sprintf(",%v", msg.Field(0).Interface())


	//
	// If it is a log message check to see if it is loggable
	//
	if msgKind == string(Log) {
		msgLogLevel := LogLevel(fmt.Sprintf(",%v", msg.Field(1).Interface()))
		if !mb.isLoggable(msgLogLevel) {
			fmt.Println("msg.PublishMsg Don't publish log message due to logging level")
			return
		}
	}

	


	//
	// Create msgStr
	//
	msgStr := fmt.Sprintf("^%v", msg.Field(0).Interface())
	for i := 1; i < msg.NumField(); i++ {
		msgStr = msgStr + fmt.Sprintf(",%v", msg.Field(i).Interface())
	}
	msgStr = msgStr + "~"

	//
	// Write to uart
	//
	if mb.uartUp != nil {
		mb.uartUp.Write([]byte(msgStr))
	}
	if mb.uartDn != nil {
		mb.uartDn.Write([]byte(msgStr))
	}

}

func unmarshallFoo(msgParts []string) *FooMsg {

	msg := new(FooMsg)

	if len(msgParts) > 0 {
		msg.Kind = Foo
	}
	if len(msgParts) > 1 {
		msg.Name = msgParts[1]
	}

	return msg
}

func unmarshallLog(msgParts []string) *LogMsg {

	msg := new(LogMsg)

	if len(msgParts) > 0 {
		msg.Kind = Log
	}
	if len(msgParts) > 1 {
		msg.Level = LogLevel(msgParts[1])
	}
	if len(msgParts) > 2 {
		msg.Source = msgParts[2]
	}
	if len(msgParts) > 3 {
		msg.Body = msgParts[3]
	}

	return msg
}
