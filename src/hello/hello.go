package main

import (
	"lcd"
	"time"
)

func main() {

	l, err := lcd.NewLcd(2, 16)
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
