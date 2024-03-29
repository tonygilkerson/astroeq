// This package contains the Struct and Methods for controlling an AMTT Encoder
// See this datasheet: https://www.cuidevices.com/product/resource/amt22.pdf
package encoder

import (
	"errors"
	"fmt"
	"machine"
	"math"
	"time"
)

// AMT22 constants
const AMT22_NOP byte = 0x00
const AMT22_RESET byte = 0x60
const AMT22_ZERO byte = 0x70
const RES12 int8 = 12
const RES14 int8 = 14
const MAX_ENCODER_READING uint32 = 16_384

// Encoder
type RAEncoder struct {
	cs                     machine.Pin
	resolution             int8
	spi                    machine.SPI
	previousEncoderReading uint32
	raPosition             uint32
	rotationCount          int16
}

// DEVTODO delete me soon
// func NewRA(spi machine.SPI, cs machine.Pin, resolution int8) RAEncoder {

// 	return RAEncoder{
// 		spi:        spi,
// 		cs:         cs,
// 		resolution: resolution,
// 	}

// }

// Configure RA encoder
func (raEncoder *RAEncoder) ConfigureEncoder(spi machine.SPI, cs machine.Pin, resolution int8) {

	raEncoder.spi = spi
	raEncoder.cs = cs
	raEncoder.resolution = resolution

	//
	// Channel select for encoder on the SPI bus
	// initialize high i.e. Not listening
	//
	raEncoder.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	raEncoder.cs.High()

}

// Zero the RA encoder
func (raEncoder *RAEncoder) ZeroRA() {

	fmt.Println("[ZeroRA] - Set RA to position zero!")
	raEncoder.WriteRead(AMT22_NOP, AMT22_ZERO)

	raEncoder.raPosition = 0
	raEncoder.previousEncoderReading = 0
	raEncoder.rotationCount = 0

	// allow time to reset
	time.Sleep(time.Millisecond * 240)

	p, e := raEncoder.GetPositionRA()
	fmt.Printf("[ZeroRA] - Check to see if it works, current position is: %v or error: %v", p, e)

}

func (raEncoder *RAEncoder) GetPositionRA() (position uint32, err error) {

	var encoderReading uint16
	r1, r2 := raEncoder.WriteRead(AMT22_NOP, AMT22_NOP)

	// Put r1 into into the upper 8 bits
	responseUpper := uint16(r1)
	responseUpper = responseUpper << 8

	// Put r2 into into the upper 8 bits
	responseLower := uint16(r2)

	// Combine upper and lowe into the same byte
	response := responseUpper | responseLower

	if parityCheck(response) == true {
		// Use the lower 14 bits for the encoderPosition
		encoderReading = response & 0x3FFF

		// Check if the difference between current and previous position is large
		// If so then we must have made a full rotation
		if math.Abs(float64(raEncoder.previousEncoderReading)-float64(encoderReading)) > float64(MAX_ENCODER_READING/2) {

			// Next check to see if we are going forward or backwards
			if uint32(encoderReading) < (MAX_ENCODER_READING/2) && raEncoder.previousEncoderReading > (MAX_ENCODER_READING/2) {
				// the encoder has moved beyond it's max in the "forward" direction, if so add to the rotationCount
				raEncoder.rotationCount++
			} else if uint32(encoderReading) > (MAX_ENCODER_READING/2) && raEncoder.previousEncoderReading < (MAX_ENCODER_READING/2) {
				// the encoder has moved beyond it's min in the "backward" direction, if so subtract from the rotationCount
				raEncoder.rotationCount--
			}

		}

		// It does not make sense to go negative
		if raEncoder.rotationCount < 0 {
			raEncoder.rotationCount = 0
		}

		// Save ra position and its previous position
		raEncoder.raPosition = uint32(encoderReading) + (uint32(raEncoder.rotationCount) * MAX_ENCODER_READING)
		raEncoder.previousEncoderReading = uint32(encoderReading)

		// println("[GetPositionRA] encoderReading: ", encoderReading, " ra.rotationCount: ", ra.rotationCount , " ra.raPosition: ", ra.raPosition)
		return raEncoder.raPosition, nil

	} else {
		return 0, errors.New("Bad parity check")
	}

}

func (raEncoder *RAEncoder) WriteRead(b1 byte, b2 byte) (r1, r2 byte) {

	// Select RA channel
	raEncoder.cs.Low()
	time.Sleep(time.Microsecond * 3) // wait min time see datasheet

	// byte 1
	r1, _ = raEncoder.spi.Transfer(b1)
	time.Sleep(time.Microsecond * 3) // wait min time see datasheet

	// byte 2
	r2, _ = machine.SPI0.Transfer(b2)
	time.Sleep(time.Microsecond * 3) // wait min time see datasheet

	// de-select RA channel
	raEncoder.cs.High()

	return r1, r2

}

func parityCheck(n uint16) bool {
	/*
		In the case of odd parity, For a given set of bits, if the count of bits with a value of 1 is even,
		the parity bit value is set to 1 making the total count of 1s in the whole set (including the parity bit) an odd number.
		If the count of bits with a value of 1 is odd, the count is already odd so the parity bit's value is 0

		The following example taken from the data sheet: https://www.mouser.com/datasheet/2/670/amt22_v-1776172.pdf

		Example:
		Full response: 0x61AB (as bits 01100001 10101011)
		14-bit position: 0x21AB (8619 decimal)

		Checkbit Formula:
			Odd:  K1 = !(H5^H3^H1^L7^L5^L3^L1)
			Even: K0 = !(H4^H2^H0^L6^L4^L2^L0)

			From the above response 0x61AB:
			Odd:  0 = !(1^0^0^1^1^1^1) = correct  - There are five 1s, five is odd  thus the parity bit should be 0
			Even: 1 = !(0^0^1^0^0^0^1) = correct  - There are two  1s, two  is even thus the parity bit should be 1

					 H6
					 | H5
					 | | H4
					 | | | H3
					 | | | | H2
					 | | | | | H1
			  	 | | | | | | H0
			  	 | | | | | | |
			0 1  1 0 0 0 0 1 1  0 1 0 1 0 1 1
			|	|                 | | | | | | |
			|	K0                | | | | | | L0
			K1	  							| | | | | L1
													| | | | L2
													| | | L3
													| | L4
													| L5
													L6

		If parity is good then use the lower 14 bits for the position,
	*/

	//
	// Loop over the lower 14 bits, counting the 1s
	// Count the number of 1s in odd and even positions
	//
	oddCount := 0
	evenCount := 0
	var i int8

	for i = 0; i < 14; i++ {
		if isKthBitSet(n, i) {
			if i%2 == 0 {
				evenCount++
			} else {
				oddCount++
			}
		}
	}

	//
	// Are the counts even or odd.
	// if even the parity bit is expected to be a 1
	//
	var isEvenCountEven bool = false
	if evenCount%2 == 0 {
		isEvenCountEven = true
	}

	var isOddCountEven bool = false
	if oddCount%2 == 0 {
		isOddCountEven = true
	}

	//
	// Get the High and Low parity bits K1 and K0
	//
	highParityBitK1 := isKthBitSet(n, 15) // The 16th bit
	lowParityBitK0 := isKthBitSet(n, 14)  // The 15th bit

	//
	// If k1 and k0 match what we found then all is good
	//
	if isEvenCountEven == lowParityBitK0 && isOddCountEven == highParityBitK1 {
		return true
	} else {
		return false
	}

}

func isKthBitSet(n uint16, k int8) bool {
	// k starts at 0
	flag := n & (1 << k)

	if flag != 0 {
		return true
	} else {
		return false
	}

}
