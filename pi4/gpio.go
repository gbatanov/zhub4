package pi4

import (
	"fmt"
	"log"

	rpio "github.com/stianeikeland/go-rpio/v4"
)

var Pi4Available bool = false

type Pi4 struct {
}

func (pi4 *Pi4) Ringer() {

}

func init() {
	// here you can check gpio availability
	fmt.Println("Init in pi4::gpio")
	err := rpio.Open()
	if err != nil {
		Pi4Available = false
		log.Println("GPIO isn't present")
	} else {
		Pi4Available = true
		rpio.Close()
	}
}
