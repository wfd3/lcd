package main

import (
	"github.com/wfd3/lcd"
	"time"
)

func main() {

	l := lcd.NewLcd(2, 16)
        err := l.EnableHW()
	if err != nil {
		panic(err)
        }
	
        l.On()
        l.BacklightOn()
        l.Clear()
        l.SetPosition(1,1)
        l.Printf(1, "Hello")
	l.Centerf(2, "There")
	l.RightJustifyf(3, "LCD") // On a 2 line display, this won't be seen.
	time.Sleep(10 * time.Second)
	l.Clear()
	l.BacklightOff()
	l.Off()
}
