// +build !arm

// stub LCD driver for i2c connected LCDs
package lcd

// Stub interface for non-ARM/Raspberry Pi platforms

type Lcd struct {
	on              bool
	rows, cols      int
}

// NewLcd creates a new LCD driver, and sets the dimensions of the display
func NewLcd(rows, cols int) *Lcd {
	lcd := Lcd{
		on:   false,
		cols: cols,
		rows: rows,
	}

	return &lcd, nil
}

// Enable enables the i2c hardware and LCD driver hardware.  This must be called before any
// other LCD function is called.
func (lcd *Lcd) EnableHW() error {
	return nil
}

// On enables the LCD function set to operate.
// Intended to support command line flags in main program
func (lcd *Lcd) On() {
	lcd.on = true
}

// Off disables the LCD functions
func (lcd *Lcd) Off() {
	lcd.on = false
}

func (lcd *Lcd) BacklightOn() {
	return
}

func (lcd *Lcd) BacklightOff() {
	return
}

// Clear clears the display
func (lcd *Lcd) Clear() {
	return
}

// Home moves the cursor to 0, 0
func (lcd *Lcd) Home() {
	return
}

// SetPosition moves the cursor to the line and column specified
func (lcd *Lcd) SetPosition(line, col byte) error {
	return nil
}

// ClearLine clears the current line
func (lcd *Lcd) ClearLine(line byte) {
	return
}

// CenterF formats using default formats, centering the line on the display
// Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.
func (lcd *Lcd) Centerf(line byte, format string, a ...interface{}) (int, error) {
	return 0, nil
}

// RightJustifyf formats using default formats, right justifying the line on the display
// Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.
func (lcd *Lcd) RightJustifyf(line byte, format string, a ...interface{}) (int, error) {
	return 0, nil
}

// Printf formats using the default formats.
// Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.
func (lcd *Lcd) Printf(line byte, format string, a ...interface{}) (int, error) {
	return 0, nil
}
