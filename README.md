# lcd
Golang package for driving i2c connected LCD displays

PACKAGE DOCUMENTATION

package lcd

LCD driver for i2c connected LCDs

TYPES

	type Lcd struct {
		// Has unexported fields.
	}

	func NewLcd(rows, cols int) (*Lcd, error)
NewLcd creates a new LCD driver, and sets the dimensions of the display

	func (lcd *Lcd) BacklightOff()

	func (lcd *Lcd) BacklightOn()

	func (lcd *Lcd) Centerf(line byte, format string, a ...interface{}) (int, error)
CenterF formats using default formats, centering the line on the display
Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.

	func (lcd *Lcd) Clear()
Clear clears the display

	func (lcd *Lcd) ClearLine(line byte)
ClearLine clears the current line

	func (lcd *Lcd) Home()
Home moves the cursor to 0, 0

	func (lcd *Lcd) Off()
Off disables the LCD functions

	func (lcd *Lcd) On()
On enables the LCD function set to operate. Intended to support command line flags in main program

	func (lcd *Lcd) Printf(line byte, format string, a ...interface{}) (int, error)
Printf formats using the default formats. Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.

	func (lcd *Lcd) RightJustifyf(line byte, format string, a ...interface{}) (int, error)
RightJustifyf formats using default formats, right justifying the line on the display Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.

	func (lcd *Lcd) SetPosition(line, col byte) error
SetPosition moves the cursor to the line and column specified

