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
type RADriverCmd string

const (
	FOO     MsgType = "Foo"
	LOG     MsgType = "Log"
	HANDSET MsgType = "Handset"
	RADRIVER MsgType = "RADriver"
)

const (
	DEBUG LogLevel = "Debug"
	INFO  LogLevel = "Info"
	WARN  LogLevel = "Warn"
	ERROR LogLevel = "Error"
)

const (
	SET_DIR_NORTH RADriverCmd = "SetDirNorth"
	SET_DIR_SOUTH RADriverCmd = "SetDirSouth"
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

// RA Driver message used for sending commands to the RA Driver and for publishing it current status
// The following are sample messages
//
// ^RADriver|SetDirNorth~
type RADriverMsg struct {
	Kind MsgType
	Cmd  RADriverCmd
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
		logLevel: INFO, // default

		uartUp:      uartUp,
		uartUpTxPin: uartUpTxPin,
		uartUpRxPin: uartUpRxPin,

		uartDn:      uartDn,
		uartDnTxPin: uartDnTxPin,
		uartDnRxPin: uartDnRxPin,

		fooCh:     nil,
		logCh:     nil,
		handsetCh: nil,
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

	case string(FOO):
		fmt.Printf("[DispatchMessage] - %v\n",FOO)
		msg := unmarshallFoo(msgParts)
		if mb.fooCh != nil {
			mb.fooCh <- *msg
		}
	case string(LOG):
		fmt.Printf("[DispatchMessage] - %v\n",LOG)
		msg := unmarshallLog(msgParts)
		if mb.logCh != nil {
			mb.logCh <- *msg
		}
	case string(HANDSET):
		fmt.Printf("[DispatchMessage] - %v\n",HANDSET)
		msg := unmarshallHandset(msgParts)
		if mb.handsetCh != nil {
			mb.handsetCh <- *msg
		}
	case string(RADRIVER):
		fmt.Printf("[DispatchMessage] - %v\n",RADRIVER)
		msg := unmarshallRADriver(msgParts)
		if mb.raDriverCh != nil {
			mb.raDriverCh <- *msg
		}
	default:
		fmt.Println("[DispatchMessage] - no match found")
	}

}

func (mb *MsgBroker) SetLogLevel(level LogLevel) {
	mb.logLevel = level
}

func (mb *MsgBroker) isLoggable(level LogLevel) bool {

	switch level {
	case DEBUG:
		return true

	case INFO:
		if mb.logLevel == ERROR || mb.logLevel == WARN || mb.logLevel == INFO {
			return true
		}
	case WARN:
		if mb.logLevel == ERROR || mb.logLevel == WARN {
			return true
		}
	case ERROR:
		if mb.logLevel == ERROR {
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
		fooMsg.Kind = FOO
	}
	if len(msgParts) > 1 {
		fooMsg.Name = msgParts[1]
	}

	return fooMsg
}

func unmarshallLog(msgParts []string) *LogMsg {

	logMsg := new(LogMsg)

	if len(msgParts) > 0 {
		logMsg.Kind = LOG
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
		handsetMsg.Kind = HANDSET
	}
	if len(msgParts) > 1 {
		handsetMsg.Keys = msgParts[1:]
	}

	return handsetMsg
}

func unmarshallRADriver(msgParts []string) *RADriverMsg { 

	raDriverMsg := new(RADriverMsg)

	if len(msgParts) > 0 {
		raDriverMsg.Kind = RADRIVER
	}
	if len(msgParts) > 1 {
		raDriverMsg.Cmd = RADriverCmd(msgParts[1])
	}

	return raDriverMsg
}

func (mb *MsgBroker) InfoLog(src string, body string) {
	mb.PublishLog(makeLogMsg(INFO, src, body))
}
func makeLogMsg(level LogLevel, src string, body string) LogMsg {

	var logMsg LogMsg
	logMsg.Kind = LOG
	logMsg.Level = level
	logMsg.Source = src
	logMsg.Body = body

	return logMsg
}
