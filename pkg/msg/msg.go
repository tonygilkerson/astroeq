package msg

import (
	"fmt"
	"machine"
	"strings"
	"time"
)

// Define message types
type MsgType string
type LogLevel string

const (
	Foo     MsgType = "Foo"
	Log     MsgType = "Log"
	Handset MsgType = "Handset"
)

const (
	Debug LogLevel = "Debug"
	Info  LogLevel = "Info"
	Warn  LogLevel = "Warn"
	Error LogLevel = "Error"
)

// DEVTODO delete foo message eventually
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
	Keys  []string
}

type MsgInterface interface {
	FooMsg | LogMsg | HandsetMsg
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

	fooCh     chan FooMsg
	logCh     chan LogMsg
	handsetCh chan HandsetMsg
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
		logLevel: Info, // default

		uartUp:      uartUp,
		uartUpTxPin: uartUpTxPin,
		uartUpRxPin: uartUpRxPin,

		uartDn:      uartDn,
		uartDnTxPin: uartDnTxPin,
		uartDnRxPin: uartDnRxPin,

		fooCh: nil,
		logCh: nil,
		handsetCh: nil,
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

func (mb *MsgBroker) SetFooCh(c chan FooMsg) {
	mb.fooCh = c
}
func (mb *MsgBroker) SetHandsetCh(c chan HandsetMsg) {
	mb.handsetCh = c
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
	case string(Handset):
		fmt.Println("[DispatchMessage] - Handset")
		msg := unmarshallHandset(msgParts)
		if mb.handsetCh != nil {
			mb.handsetCh <- *msg
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
func (mb *MsgBroker) PublishHandset(handsetMsg HandsetMsg) {

	msgStr := "^" + string(handsetMsg.Kind)

	for _, key := range handsetMsg.Keys {
		msgStr = msgStr + "|" + key
	}
	msgStr = msgStr + "~"

	mb.PublishMsg(msgStr)

}

func (mb *MsgBroker) PublishMsg(msg string) {

	if mb.uartUp != nil {
		mb.uartUp.Write([]byte(msg))
	}

	if mb.uartDn != nil {
		mb.uartDn.Write([]byte(msg))
	}
}

func unmarshallFoo(msgParts []string) *FooMsg {

	fooMsg := new(FooMsg)

	if len(msgParts) > 0 {
		fooMsg.Kind = Foo
	}
	if len(msgParts) > 1 {
		fooMsg.Name = msgParts[1]
	}

	return fooMsg
}

func unmarshallLog(msgParts []string) *LogMsg {

	logMsg := new(LogMsg)

	if len(msgParts) > 0 {
		logMsg.Kind = Log
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
		handsetMsg.Kind = Handset
	}
	if len(msgParts) > 1 {
		handsetMsg.Keys = msgParts[1:]
	}

	return handsetMsg
}

func (mb *MsgBroker) InfoLog(src string, body string) {
	mb.PublishLog(makeLogMsg(Info, src, body))
}
func makeLogMsg(level LogLevel, src string, body string) LogMsg {

	var logMsg LogMsg
	logMsg.Kind = Log
	logMsg.Level = level
	logMsg.Source = src
	logMsg.Body = body

	return logMsg
}
