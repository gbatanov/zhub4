package serial3

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

const SOF byte = 0xFE

type Uart struct {
	port       string
	os         string
	comport    *serial.Port
	Flag       bool
	portOpened bool
}

func init() {
	fmt.Println("Init in serial3")
	// TODO: check availability serial port
}

func UartCreate(port string, os string) *Uart {
	uart := Uart{port: port, os: os}
	return &uart
}

func (u *Uart) Open() error {
	comport, err := u.openPort()
	if err != nil {
		log.Println("Error open port ", u.port)
		return err
	}
	u.Flag = true
	u.comport = comport
	u.portOpened = true
	return nil
}

// Opening the given port
func (u Uart) openPort() (*serial.Port, error) {

	// bitrate for linux (checked!) - 115200 230400 460800 500000 576000
	// bitrate fo Mac - 115200 only!
	var baud int = 576000
	if u.os == "darwin" {
		baud = 115200
	}
	c := &serial.Config{Name: u.port, Baud: baud, ReadTimeout: time.Second * 3}
	return serial.OpenPort(c)

}

func (u *Uart) Stop() {
	if u.portOpened {
		u.Flag = false
		u.comport.Flush()
		u.comport.Close()
		u.portOpened = false
		log.Println("comport closed")
	}
}

// write a sequence of bytes to serial port
func (u Uart) Write(text []byte) error {
	n, err := u.comport.Write(text)
	if err != nil {
		return err
	}
	if n != len(text) {
		return errors.New("Write error")
	}
	return nil
}

// The cycle of receiving commands from the zhub
// in this serial port library version we get chunks 64 byte size !!!
func (u *Uart) Loop(cmdinput chan []byte) {
	for u.Flag {
		BufRead := make([]byte, 256)
		n, err := u.comport.Read(BufRead)
		if err != nil {
			if n != 0 {
				u.Flag = false
			}
		} else if err == nil && n > 0 {
			cmdinput <- BufRead[:n]
		} else {
			// if there is no command, wait 1 sec
			time.Sleep(1 * time.Second)
		}
	}
}

func (u Uart) Send_command_to_device(buff []byte) error {

	err := u.Write(buff)
	return err
}
