// +build arm

// LCD driver for i2c connected LCDs
package lcd

// Adapted from github.com/davecheney/i2c, which was:
//   Adapted from http://think-bowl.com/raspberry-pi/installing-the-think-bowl-i2c-libraries-for-python/
//
// See https://orientdisplay.com/wp-content/uploads/2018/08/AMC1602AI2C-Full-1.pdf
// Also see: https://www.sunfounder.com/learn/sensor-kit-v2-0-for-raspberry-pi-b-plus/lesson-30-i2c-lcd1602-sensor-kit-v2-0-for-b-plus.html
// TODO: Need to find a real spec sheet.
//

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// I2C
	i2c_ADDR  = 0x27
	i2c_SLAVE = 0x0703

	// LCD Commands
	_CMD_Clear_Display        = 0x01
	_CMD_Return_Home          = 0x02
	_CMD_Entry_Mode           = 0x04
	_CMD_Display_Control      = 0x08
	_CMD_Cursor_Display_Shift = 0x10
	_CMD_Function_Set         = 0x20
	_CMD_DDRAM_Set            = 0x80

	// LCD Options
	_OPT_Increment      = 0x02 // CMD_Entry_Mode
	_OPT_Enable_Display = 0x04 // CMD_Display_Control
	_OPT_Enable_Cursor  = 0x02 // CMD_Display_Control
	_OPT_Enable_Blink   = 0x01 // CMD_Display_Control
	_OPT_Display_Shift  = 0x08 // CMD_Cursor_Display_Shift
	_OPT_Shift_Right    = 0x04 // CMD_Cursor_Display_Shift 0 = Left
	_OPT_2_Lines        = 0x08 // CMD_Function_Set 0 = 1 line
	_OPT_5x10_Dots      = 0x04 // CMD_Function_Set 0 = 5x7 dots

	// LCD instruction offsets
	_OFFSET_RS        = 0
	_OFFSET_EN        = 2
	_OFFSET_BACKLIGHT = 3
	_OFFSET_D4        = 4
	_OFFSET_D5        = 5
	_OFFSET_D6        = 6
	_OFFSET_D7        = 7
)

type Lcd struct {
	on              bool
	i2c             *os.File
	backlight_state bool
	cols, rows      int
	m               sync.Mutex
}

func ioctl(fd, cmd, arg uintptr) (err error) {
	_, _, e := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if e != 0 {
		err = e
	}
	return err
}

func (lcd *Lcd) writeI2C(b byte) (int, error) {
	if lcd.i2c == nil {
		panic("Attempt to write to LCD I2C without calling lcd.Enable()")
	}
	var buf [1]byte

	buf[0] = b
	return lcd.i2c.Write(buf[:])
}

func pinInterpret(pin, data byte, value bool) byte {
	if value {
		// Construct mask using pin
		var mask byte = 0x01 << (pin)
		data = data | mask
	} else {
		// Construct mask using pin
		var mask byte = 0x01<<(pin) ^ 0xFF
		data = data & mask
	}
	return data
}

func (lcd *Lcd) enable(data byte) {
	// Determine if black light is on and insure it does not turn off or on
	data = pinInterpret(_OFFSET_BACKLIGHT, data, lcd.backlight_state)
	lcd.writeI2C(data)
	lcd.writeI2C(pinInterpret(_OFFSET_EN, data, true))
	lcd.writeI2C(data)
}

func (lcd *Lcd) write(data byte, command bool) {
	var i2c_data byte

	// Add data for high nibble
	hi_nibble := data >> 4
	i2c_data = pinInterpret(_OFFSET_D4, i2c_data, (hi_nibble&0x01 == 0x01))
	i2c_data = pinInterpret(_OFFSET_D5, i2c_data, ((hi_nibble>>1)&0x01 == 0x01))
	i2c_data = pinInterpret(_OFFSET_D6, i2c_data, ((hi_nibble>>2)&0x01 == 0x01))
	i2c_data = pinInterpret(_OFFSET_D7, i2c_data, ((hi_nibble>>3)&0x01 == 0x01))

	// # Set the register selector to 1 if this is data
	if !command {
		i2c_data = pinInterpret(_OFFSET_RS, i2c_data, true)
	}

	//  Toggle Enable
	lcd.enable(i2c_data)

	i2c_data = 0x00

	// Add data for high nibble
	low_nibble := data & 0x0F
	i2c_data = pinInterpret(_OFFSET_D4, i2c_data, (low_nibble&0x01 == 0x01))
	i2c_data = pinInterpret(_OFFSET_D5, i2c_data, ((low_nibble>>1)&0x01 == 0x01))
	i2c_data = pinInterpret(_OFFSET_D6, i2c_data, ((low_nibble>>2)&0x01 == 0x01))
	i2c_data = pinInterpret(_OFFSET_D7, i2c_data, ((low_nibble>>3)&0x01 == 0x01))

	// Set the register selector to 1 if this is data
	if !command {
		i2c_data = pinInterpret(_OFFSET_RS, i2c_data, true)
	}

	lcd.enable(i2c_data)
}

func (lcd *Lcd) command(data byte) {
	lcd.write(data, true)
}

func (lcd *Lcd) writeBuf(buf []byte) (int, error) {
	for _, c := range buf {
		lcd.write(c, false)
	}
	return len(buf), nil
}

func (lcd *Lcd) getLCDaddress(line, pos byte) (byte, error) {
	var address byte
	if line > byte(lcd.rows) {
		return 0, fmt.Errorf("invalid line number %d, max %d", line, lcd.rows)
	}
	if pos > byte(lcd.cols) {
		return 0, fmt.Errorf("invalid column number %d, max %d", pos, lcd.cols)
	}

	switch line {
	case 1:
		address = pos
	case 2:
		address = 0x40 + pos
	case 3:
		address = 0x14 + pos
	case 4:
		address = 0x54 + pos
	}

	return address, nil
}

func capstring(s string, l int) string {
	if len(s) > l {
		s = s[:l]
	}
	return s
}

func pad(s string, l int) string {
	if len(s) < l {
		for i := len(s); i < l; i++ {
			s += " "
		}
	}
	return s
}

func shift(s string, l int) string {
	for i := 0; i < l; i++ {
		s = " " + s
	}
	return s
}

func sfmt(format string, a ...interface{}) string {
	out := fmt.Sprintf(format, a...)
	out = strings.Replace(out, "\n", "", -1)
	return out
}

func (lcd *Lcd) print(line byte, out string) (int, error) {
	out = capstring(out, lcd.cols)
	address, err := lcd.getLCDaddress(line, 0)
	if err != nil {
		return 0, err
	}

	lcd.m.Lock()
	defer lcd.m.Unlock()

	lcd.command(_CMD_DDRAM_Set + address) // Do this here to prevent race between print() and SetPosition()
	return lcd.writeBuf([]byte(out))
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// NewLcd creates a new LCD driver, and sets the dimensions of the display
func NewLcd(rows, cols int) *Lcd {

	lcd := Lcd{
		on:   false,
		i2c:  nil,
		cols: cols,
		rows: rows,
	}
	return &lcd
}

// Enable enables the i2c hardware and LCD driver hardware.  This must be called before any other LCD
// function is called.
func (lcd *Lcd) EnableHW() error {
	i2c, err := os.OpenFile("/dev/i2c-1", os.O_RDWR, 0600)
	if err != nil {
		err = fmt.Errorf("NewLcd(): Can't open i2c device: %s", err)
		return nil
	}
	if err := ioctl(i2c.Fd(), i2c_SLAVE, uintptr(i2c_ADDR)); err != nil {
		err = fmt.Errorf("NewLcd(): ioctl() error: %s", err)
		return nil
	}

	// Activate LCD

	lcd.i2c = i2c
	
	var data byte
	data = pinInterpret(_OFFSET_D4, data, true)
	data = pinInterpret(_OFFSET_D5, data, true)
	lcd.enable(data)
	time.Sleep(200 * time.Millisecond)
	lcd.enable(data)
	time.Sleep(100 * time.Millisecond)
	lcd.enable(data)
	time.Sleep(100 * time.Millisecond)

	// Initialize 4-bit mode
	data = pinInterpret(_OFFSET_D4, data, false)
	lcd.enable(data)
	time.Sleep(10 * time.Millisecond)

	lcd.command(_CMD_Function_Set | _OPT_2_Lines)
	lcd.command(_CMD_Display_Control | _OPT_Enable_Display | _OPT_Enable_Cursor)
	lcd.command(_CMD_Clear_Display)
	lcd.command(_CMD_Entry_Mode | _OPT_Increment | _OPT_Display_Shift)

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
	if !lcd.on {
		return
	}
	lcd.m.Lock()
	defer lcd.m.Unlock()
	lcd.writeI2C(pinInterpret(_OFFSET_BACKLIGHT, 0x00, true))
	lcd.backlight_state = true
}

func (lcd *Lcd) BacklightOff() {
	if !lcd.on {
		return
	}
	lcd.m.Lock()
	defer lcd.m.Unlock()
	lcd.writeI2C(pinInterpret(_OFFSET_BACKLIGHT, 0x00, false))
	lcd.backlight_state = false
}

// Clear clears the display
func (lcd *Lcd) Clear() {
	if !lcd.on {
		return
	}
	lcd.m.Lock()
	defer lcd.m.Unlock()
	lcd.command(_CMD_Clear_Display)
}

// Home moves the cursor to 0, 0
func (lcd *Lcd) Home() {
	if !lcd.on {
		return
	}
	lcd.m.Lock()
	defer lcd.m.Unlock()
	lcd.command(_CMD_Return_Home)
}

// SetPosition moves the cursor to the line and column specified
func (lcd *Lcd) SetPosition(line, col byte) error {
	if !lcd.on {
		return nil
	}
	address, err := lcd.getLCDaddress(line, col)
	if err != nil {
		return err
	}

	lcd.m.Lock()
	defer lcd.m.Unlock()
	lcd.command(_CMD_DDRAM_Set + address)
	return nil
}

// ClearLine clears the current line
func (lcd *Lcd) ClearLine(line byte) {
	if !lcd.on {
		return
	}
	s := pad("", lcd.cols)
	lcd.print(line, s)
}

// CenterF formats using default formats, centering the line on the display
// Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.
func (lcd *Lcd) Centerf(line byte, format string, a ...interface{}) (int, error) {
	if !lcd.on {
		return 0, nil
	}
	out := sfmt(format, a...)
	out = shift(out, (lcd.cols-len(out))/2)
	out = pad(out, lcd.cols)
	return lcd.print(line, out)
}

// RightJustifyf formats using default formats, right justifying the line on the display
// Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.
func (lcd *Lcd) RightJustifyf(line byte, format string, a ...interface{}) (int, error) {
	if !lcd.on {
		return 0, nil
	}
	out := sfmt(format, a...)
	out = shift(out, lcd.cols-len(out))
	return lcd.print(line, out)
}

// Printf formats using the default formats.
// Uses the same modifiers as fmt.Printf(), and returns the number of bytes written and any error encountered.
func (lcd *Lcd) Printf(line byte, format string, a ...interface{}) (int, error) {
	if !lcd.on {
		return 0, nil
	}
	out := sfmt(format, a...)
	out = pad(out, lcd.cols)
	return lcd.print(line, out)
}
